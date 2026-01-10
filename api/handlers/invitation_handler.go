package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/arauth-identity/iam/api/middleware"
	"github.com/arauth-identity/iam/identity/invitation"
)

// InvitationHandler handles invitation-related HTTP requests
type InvitationHandler struct {
	invitationService invitation.ServiceInterface
}

// NewInvitationHandler creates a new invitation handler
func NewInvitationHandler(invitationService invitation.ServiceInterface) *InvitationHandler {
	return &InvitationHandler{
		invitationService: invitationService,
	}
}

// CreateInvitation handles POST /api/v1/invitations
func (h *InvitationHandler) CreateInvitation(c *gin.Context) {
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

	invitedBy, err := uuid.Parse(userID.(string))
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_user_id",
			"Invalid user ID format", nil)
		return
	}

	var req invitation.CreateInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_request",
			"Request validation failed", middleware.FormatValidationErrors(err))
		return
	}

	inv, err := h.invitationService.CreateInvitation(c.Request.Context(), tenantID, invitedBy, &req)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "creation_failed",
			err.Error(), nil)
		return
	}

	c.JSON(http.StatusCreated, inv)
}

// GetInvitation handles GET /api/v1/invitations/:id
func (h *InvitationHandler) GetInvitation(c *gin.Context) {
	tenantID, ok := middleware.RequireTenant(c)
	if !ok {
		return
	}

	invitationIDStr := c.Param("id")
	invitationID, err := uuid.Parse(invitationIDStr)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_id",
			"Invalid invitation ID format", nil)
		return
	}

	inv, err := h.invitationService.GetInvitation(c.Request.Context(), invitationID)
	if err != nil {
		middleware.RespondWithError(c, http.StatusNotFound, "not_found",
			"Invitation not found", nil)
		return
	}

	// Verify tenant ownership
	if inv.TenantID != tenantID {
		middleware.RespondWithError(c, http.StatusNotFound, "not_found",
			"Invitation not found", nil)
		return
	}

	c.JSON(http.StatusOK, inv)
}

// ListInvitations handles GET /api/v1/invitations
func (h *InvitationHandler) ListInvitations(c *gin.Context) {
	tenantID, ok := middleware.RequireTenant(c)
	if !ok {
		return
	}

	// Parse query parameters
	filters := &invitation.ListInvitationsFilters{
		Email:    c.Query("email"),
		Status:   c.Query("status"),
		Page:     1,
		PageSize: 10,
	}

	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			filters.Page = page
		}
	}

	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 {
			if pageSize > 100 {
				pageSize = 100
			}
			filters.PageSize = pageSize
		}
	}

	if invitedByStr := c.Query("invited_by"); invitedByStr != "" {
		if invitedBy, err := uuid.Parse(invitedByStr); err == nil {
			filters.InvitedBy = &invitedBy
		}
	}

	invitations, total, err := h.invitationService.ListInvitations(c.Request.Context(), tenantID, filters)
	if err != nil {
		middleware.RespondWithError(c, http.StatusInternalServerError, "list_failed",
			"Failed to list invitations", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"invitations": invitations,
		"pagination": gin.H{
			"page":        filters.Page,
			"page_size":   filters.PageSize,
			"total":       total,
			"total_pages": (total + filters.PageSize - 1) / filters.PageSize,
		},
	})
}

// ResendInvitation handles POST /api/v1/invitations/:id/resend
func (h *InvitationHandler) ResendInvitation(c *gin.Context) {
	tenantID, ok := middleware.RequireTenant(c)
	if !ok {
		return
	}

	invitationIDStr := c.Param("id")
	invitationID, err := uuid.Parse(invitationIDStr)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_id",
			"Invalid invitation ID format", nil)
		return
	}

	// Verify tenant ownership
	inv, err := h.invitationService.GetInvitation(c.Request.Context(), invitationID)
	if err != nil {
		middleware.RespondWithError(c, http.StatusNotFound, "not_found",
			"Invitation not found", nil)
		return
	}

	if inv.TenantID != tenantID {
		middleware.RespondWithError(c, http.StatusNotFound, "not_found",
			"Invitation not found", nil)
		return
	}

	if err := h.invitationService.ResendInvitation(c.Request.Context(), invitationID); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "resend_failed",
			err.Error(), nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Invitation resent successfully"})
}

// AcceptInvitation handles POST /api/v1/invitations/accept
func (h *InvitationHandler) AcceptInvitation(c *gin.Context) {
	var req struct {
		Token     string `json:"token" binding:"required"`
		Username  string `json:"username" binding:"required,min=3,max=255"`
		Password  string `json:"password" binding:"required,min=12"`
		FirstName *string `json:"first_name,omitempty"`
		LastName  *string `json:"last_name,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_request",
			"Request validation failed", middleware.FormatValidationErrors(err))
		return
	}

	acceptReq := &invitation.AcceptInvitationRequest{
		Username:  req.Username,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	user, err := h.invitationService.AcceptInvitation(c.Request.Context(), req.Token, acceptReq)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "accept_failed",
			err.Error(), nil)
		return
	}

	c.JSON(http.StatusOK, user)
}

// DeleteInvitation handles DELETE /api/v1/invitations/:id
func (h *InvitationHandler) DeleteInvitation(c *gin.Context) {
	tenantID, ok := middleware.RequireTenant(c)
	if !ok {
		return
	}

	invitationIDStr := c.Param("id")
	invitationID, err := uuid.Parse(invitationIDStr)
	if err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_id",
			"Invalid invitation ID format", nil)
		return
	}

	// Verify tenant ownership
	inv, err := h.invitationService.GetInvitation(c.Request.Context(), invitationID)
	if err != nil {
		middleware.RespondWithError(c, http.StatusNotFound, "not_found",
			"Invitation not found", nil)
		return
	}

	if inv.TenantID != tenantID {
		middleware.RespondWithError(c, http.StatusNotFound, "not_found",
			"Invitation not found", nil)
		return
	}

	if err := h.invitationService.DeleteInvitation(c.Request.Context(), invitationID); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "delete_failed",
			err.Error(), nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Invitation deleted successfully"})
}

