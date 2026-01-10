package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/arauth-identity/iam/api/middleware"
	"github.com/arauth-identity/iam/auth/claims"
	"github.com/arauth-identity/iam/identity/audit"
	"github.com/arauth-identity/iam/identity/models"
	"github.com/arauth-identity/iam/storage/interfaces"
)

// AuditHandler handles audit event-related HTTP requests
type AuditHandler struct {
	auditService audit.ServiceInterface
}

// NewAuditHandler creates a new audit handler
func NewAuditHandler(auditService audit.ServiceInterface) *AuditHandler {
	return &AuditHandler{
		auditService: auditService,
	}
}

// QueryEvents handles GET /api/v1/audit/events
func (h *AuditHandler) QueryEvents(c *gin.Context) {
	// Get tenant ID from context (if tenant-scoped)
	var tenantID *uuid.UUID
	if tenantIDStr, exists := middleware.GetTenantID(c); exists {
		tenantID = &tenantIDStr
	}

	// Build filters from query parameters
	filters := &interfaces.AuditEventFilters{
		TenantID: tenantID,
	}

	// Event type filter
	if eventType := c.Query("event_type"); eventType != "" {
		filters.EventType = &eventType
	}

	// Actor user ID filter
	if actorUserIDStr := c.Query("actor_user_id"); actorUserIDStr != "" {
		if actorUserID, err := uuid.Parse(actorUserIDStr); err == nil {
			filters.ActorUserID = &actorUserID
		}
	}

	// Target type filter
	if targetType := c.Query("target_type"); targetType != "" {
		filters.TargetType = &targetType
	}

	// Target ID filter
	if targetIDStr := c.Query("target_id"); targetIDStr != "" {
		if targetID, err := uuid.Parse(targetIDStr); err == nil {
			filters.TargetID = &targetID
		}
	}

	// Result filter
	if result := c.Query("result"); result != "" {
		filters.Result = &result
	}

	// Date range filters
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if startDate, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			filters.StartDate = &startDate
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if endDate, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			filters.EndDate = &endDate
		}
	}

	// Pagination
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	filters.Page = page

	pageSize := interfaces.DefaultPageSize
	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			if ps > interfaces.MaxPageSize {
				ps = interfaces.MaxPageSize
			}
			pageSize = ps
		}
	}
	filters.PageSize = pageSize

	// Query events
	events, total, err := h.auditService.QueryEvents(c.Request.Context(), filters)
	if err != nil {
		middleware.RespondWithError(c, http.StatusInternalServerError, "query_failed",
			"Failed to query audit events", nil)
		return
	}

	// Calculate pagination metadata
	totalPages := (total + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}

	c.JSON(http.StatusOK, gin.H{
		"events": events,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

// GetEvent handles GET /api/v1/audit/events/:id
func (h *AuditHandler) GetEvent(c *gin.Context) {
	eventIDStr := c.Param("id")
	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_id",
			"Invalid event ID format", nil)
		return
	}

	event, err := h.auditService.GetEvent(c.Request.Context(), eventID)
	if err != nil {
		middleware.RespondWithError(c, http.StatusNotFound, "not_found",
			"Audit event not found", nil)
		return
	}

	// Check tenant access (if tenant-scoped)
	if event.TenantID != nil {
		tenantID, exists := middleware.GetTenantID(c)
		if !exists || *event.TenantID != tenantID {
			middleware.RespondWithError(c, http.StatusForbidden, "forbidden",
				"Access denied", nil)
			return
		}
	}

	c.JSON(http.StatusOK, event)
}

// extractActorFromContext extracts actor information from Gin context
func extractActorFromContext(c *gin.Context) (models.AuditActor, error) {
	claimsObj, exists := c.Get("user_claims")
	if !exists {
		return models.AuditActor{}, fmt.Errorf("user claims not found in context")
	}

	userClaims := claimsObj.(*claims.Claims)

	userID, err := uuid.Parse(userClaims.Subject)
	if err != nil {
		return models.AuditActor{}, fmt.Errorf("invalid user ID in claims: %w", err)
	}

	username := userClaims.Username
	if username == "" {
		username = userClaims.Email
	}

	principalType := userClaims.PrincipalType
	if principalType == "" {
		principalType = "TENANT" // Default to TENANT if not specified
	}

	return models.AuditActor{
		UserID:        userID,
		Username:      username,
		PrincipalType: principalType,
	}, nil
}

// extractSourceInfo extracts source IP and user agent from request
func extractSourceInfo(c *gin.Context) (sourceIP, userAgent string) {
	// Get source IP
	sourceIP = c.ClientIP()
	if forwardedFor := c.GetHeader("X-Forwarded-For"); forwardedFor != "" {
		sourceIP = forwardedFor
	} else if realIP := c.GetHeader("X-Real-IP"); realIP != "" {
		sourceIP = realIP
	}

	// Get user agent
	userAgent = c.GetHeader("User-Agent")

	return sourceIP, userAgent
}

