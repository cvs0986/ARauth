package handlers

import (
	"net/http"

	"github.com/arauth-identity/iam/api/middleware"
	"github.com/arauth-identity/iam/identity/audit"
	"github.com/arauth-identity/iam/identity/models"
	"github.com/arauth-identity/iam/identity/session"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SessionHandler handles session-related HTTP requests
type SessionHandler struct {
	sessionService session.ServiceInterface
	auditService   audit.ServiceInterface
}

// NewSessionHandler creates a new session handler
func NewSessionHandler(sessionService session.ServiceInterface, auditService audit.ServiceInterface) *SessionHandler {
	return &SessionHandler{
		sessionService: sessionService,
		auditService:   auditService,
	}
}

// ListSessions handles GET /api/v1/sessions
func (h *SessionHandler) ListSessions(c *gin.Context) {
	// Get tenant ID from context (set by tenant middleware)
	tenantID, ok := middleware.RequireTenant(c)
	if !ok {
		return
	}

	// Get user ID from JWT claims
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, http.StatusUnauthorized, "unauthorized",
			"User ID not found in token", nil)
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_user_id",
			"Invalid user ID format", nil)
		return
	}

	// List sessions for the user
	sessions, err := h.sessionService.ListSessions(c.Request.Context(), userUUID, tenantID)
	if err != nil {
		middleware.RespondWithError(c, http.StatusInternalServerError, "internal_error",
			"Failed to list sessions", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sessions": sessions,
		"count":    len(sessions),
	})
}

// RevokeSessionRequest represents the request to revoke a session
type RevokeSessionRequest struct {
	AuditReason string `json:"audit_reason" binding:"required,min=10"`
}

// RevokeSession handles POST /api/v1/sessions/:id/revoke
func (h *SessionHandler) RevokeSession(c *gin.Context) {
	// Get tenant ID from context (set by tenant middleware)
	tenantID, ok := middleware.RequireTenant(c)
	if !ok {
		return
	}

	// Parse session ID from URL
	sessionIDStr := c.Param("id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_session_id",
			"Invalid session ID format", nil)
		return
	}

	// Validate request body
	var req RevokeSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_request",
			"Request validation failed", middleware.FormatValidationErrors(err))
		return
	}

	// Get user ID from JWT claims
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, http.StatusUnauthorized, "unauthorized",
			"User ID not found in token", nil)
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_user_id",
			"Invalid user ID format", nil)
		return
	}

	// Verify session belongs to user's tenant (tenant isolation)
	// List user's sessions and check if the session ID is in the list
	sessions, err := h.sessionService.ListSessions(c.Request.Context(), userUUID, tenantID)
	if err != nil {
		middleware.RespondWithError(c, http.StatusInternalServerError, "internal_error",
			"Failed to verify session ownership", nil)
		return
	}

	// Check if session belongs to user
	sessionFound := false
	var targetSession *session.Session
	for _, s := range sessions {
		if s.ID == sessionID {
			sessionFound = true
			targetSession = s
			break
		}
	}

	if !sessionFound {
		middleware.RespondWithError(c, http.StatusForbidden, "forbidden",
			"Session not found or does not belong to your tenant", nil)
		return
	}

	// Revoke the session
	if err := h.sessionService.RevokeSession(c.Request.Context(), sessionID, req.AuditReason); err != nil {
		middleware.RespondWithError(c, http.StatusInternalServerError, "internal_error",
			"Failed to revoke session", nil)
		return
	}

	// Log audit event
	actor, err := extractActorFromContext(c)
	if err != nil {
		// Log error but don't fail the request - session is already revoked
	}

	target := &models.AuditTarget{
		Type: "session",
		ID:   sessionID,
	}

	sourceIP, userAgent := extractSourceInfo(c)

	metadata := map[string]interface{}{
		"audit_reason": req.AuditReason,
		"session_id":   sessionID.String(),
		"user_id":      targetSession.UserID.String(),
		"username":     targetSession.Username,
		"action":       "session_revoked",
	}

	h.auditService.LogUserUpdated(c.Request.Context(), actor, target, &tenantID, sourceIP, userAgent, metadata)

	c.JSON(http.StatusOK, gin.H{
		"message":    "Session revoked successfully",
		"session_id": sessionID.String(),
	})
}
