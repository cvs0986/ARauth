package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/arauth-identity/iam/api/middleware"
	"github.com/arauth-identity/iam/auth/claims"
	"github.com/arauth-identity/iam/identity/audit"
	"github.com/arauth-identity/iam/identity/impersonation"
	"github.com/arauth-identity/iam/identity/models"
)

// ImpersonationHandler handles impersonation requests
type ImpersonationHandler struct {
	impersonationService impersonation.ServiceInterface
	auditService         audit.ServiceInterface
}

// NewImpersonationHandler creates a new impersonation handler
func NewImpersonationHandler(impersonationService impersonation.ServiceInterface, auditService audit.ServiceInterface) *ImpersonationHandler {
	return &ImpersonationHandler{
		impersonationService: impersonationService,
		auditService:         auditService,
	}
}

// StartImpersonation handles POST /api/v1/users/:id/impersonate
func (h *ImpersonationHandler) StartImpersonation(c *gin.Context) {
	// Get impersonator from context (current user)
	userClaims, exists := c.Get("user_claims")
	if !exists {
		middleware.RespondWithError(c, http.StatusUnauthorized, "unauthorized",
			"User claims not found", nil)
		return
	}

	claimsObj, ok := userClaims.(*claims.Claims)
	if !ok {
		middleware.RespondWithError(c, http.StatusUnauthorized, "unauthorized",
			"Invalid user claims", nil)
		return
	}

	impersonatorID, err := uuid.Parse(claimsObj.Subject)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_user_id",
			"Invalid impersonator user ID", nil)
		return
	}

	// Get target user ID from URL
	targetUserIDStr := c.Param("id")
	targetUserID, err := uuid.Parse(targetUserIDStr)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_user_id",
			"Invalid target user ID", nil)
		return
	}

	// Parse request body for optional reason
	var req struct {
		Reason *string `json:"reason,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil && c.Request.ContentLength > 0 {
		// Only error if there's content - empty body is OK
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_request",
			"Request validation failed", middleware.FormatValidationErrors(err))
		return
	}

	// Permission check: SYSTEM users can impersonate anyone, TENANT users need permission
	// For now, we'll check in middleware or add explicit permission check here
	// TODO: Add permission check for tenant.admin.impersonate

	// Start impersonation
	result, err := h.impersonationService.StartImpersonation(c.Request.Context(), impersonatorID, targetUserID, req.Reason)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "impersonation_failed",
			err.Error(), nil)
		return
	}

	// Log audit event
	if actor, err := extractActorFromContext(c); err == nil {
		sourceIP, userAgent := extractSourceInfo(c)
		target := &models.AuditTarget{
			Type:       "user",
			ID:         targetUserID,
			Identifier: "impersonation_target",
		}
		tenantID := result.Session.TenantID
		_ = h.auditService.LogEvent(c.Request.Context(), models.EventTypeUserImpersonated, actor, target, &tenantID, sourceIP, userAgent, map[string]interface{}{
			"session_id":      result.Session.ID,
			"impersonator_id": impersonatorID,
			"target_user_id":  targetUserID,
			"reason":          req.Reason,
		}, models.ResultSuccess, nil)
	}

	c.JSON(http.StatusOK, result)
}

// EndImpersonation handles DELETE /api/v1/impersonation/:session_id
func (h *ImpersonationHandler) EndImpersonation(c *gin.Context) {
	// Get user from context
	userClaims, exists := c.Get("user_claims")
	if !exists {
		middleware.RespondWithError(c, http.StatusUnauthorized, "unauthorized",
			"User claims not found", nil)
		return
	}

	claimsObj, ok := userClaims.(*claims.Claims)
	if !ok {
		middleware.RespondWithError(c, http.StatusUnauthorized, "unauthorized",
			"Invalid user claims", nil)
		return
	}

	endedBy, err := uuid.Parse(claimsObj.Subject)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_user_id",
			"Invalid user ID", nil)
		return
	}

	// Get session ID from URL
	sessionIDStr := c.Param("session_id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_session_id",
			"Invalid session ID", nil)
		return
	}

	// End impersonation
	err = h.impersonationService.EndImpersonation(c.Request.Context(), sessionID, endedBy)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "impersonation_end_failed",
			err.Error(), nil)
		return
	}

	// Log audit event
	if actor, err := extractActorFromContext(c); err == nil {
		sourceIP, userAgent := extractSourceInfo(c)
		target := &models.AuditTarget{
			Type:       "impersonation_session",
			ID:         sessionID,
			Identifier: "impersonation_session",
		}
		_ = h.auditService.LogEvent(c.Request.Context(), models.EventTypeUserImpersonationEnded, actor, target, nil, sourceIP, userAgent, map[string]interface{}{
			"session_id": sessionID,
			"ended_by":   endedBy,
		}, models.ResultSuccess, nil)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Impersonation session ended",
	})
}

// ListImpersonationSessions handles GET /api/v1/impersonation
func (h *ImpersonationHandler) ListImpersonationSessions(c *gin.Context) {
	// Get user from context
	userClaims, exists := c.Get("user_claims")
	if !exists {
		middleware.RespondWithError(c, http.StatusUnauthorized, "unauthorized",
			"User claims not found", nil)
		return
	}

	claimsObj, ok := userClaims.(*claims.Claims)
	if !ok {
		middleware.RespondWithError(c, http.StatusUnauthorized, "unauthorized",
			"Invalid user claims", nil)
		return
	}

	// Parse query parameters
	var filters impersonation.ImpersonationFilters
	if impersonatorIDStr := c.Query("impersonator_id"); impersonatorIDStr != "" {
		if id, err := uuid.Parse(impersonatorIDStr); err == nil {
			filters.ImpersonatorID = &id
		}
	}
	if targetUserIDStr := c.Query("target_user_id"); targetUserIDStr != "" {
		if id, err := uuid.Parse(targetUserIDStr); err == nil {
			filters.TargetUserID = &id
		}
	}
	if tenantIDStr := c.Query("tenant_id"); tenantIDStr != "" {
		if id, err := uuid.Parse(tenantIDStr); err == nil {
			filters.TenantID = &id
		}
	}
	filters.ActiveOnly = c.Query("active_only") == "true"
	filters.Page = 1
	filters.PageSize = 20
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := parseUint(pageStr); err == nil {
			filters.Page = page
		}
	}
	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if pageSize, err := parseUint(pageSizeStr); err == nil {
			filters.PageSize = pageSize
		}
	}

	// For tenant users, filter by their tenant
	if claimsObj.TenantID != "" {
		if tenantID, err := uuid.Parse(claimsObj.TenantID); err == nil {
			filters.TenantID = &tenantID
		}
	}

	// Get sessions
	sessions, err := h.impersonationService.GetActiveSessions(c.Request.Context(), &filters)
	if err != nil {
		middleware.RespondWithError(c, http.StatusInternalServerError, "query_failed",
			err.Error(), nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sessions": sessions,
		"total":    len(sessions),
	})
}

// GetImpersonationSession handles GET /api/v1/impersonation/:session_id
func (h *ImpersonationHandler) GetImpersonationSession(c *gin.Context) {
	// Get session ID from URL
	sessionIDStr := c.Param("session_id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_session_id",
			"Invalid session ID", nil)
		return
	}

	// Get session
	session, err := h.impersonationService.GetSession(c.Request.Context(), sessionID)
	if err != nil {
		middleware.RespondWithError(c, http.StatusNotFound, "session_not_found",
			err.Error(), nil)
		return
	}

	c.JSON(http.StatusOK, session)
}

// Helper function to parse uint from string
func parseUint(s string) (int, error) {
	result, err := strconv.Atoi(s)
	return result, err
}

