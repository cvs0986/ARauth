package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/arauth-identity/iam/api/middleware"
	"github.com/arauth-identity/iam/identity/webhook"
)

// WebhookHandler handles webhook-related HTTP requests
type WebhookHandler struct {
	webhookService webhook.ServiceInterface
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(webhookService webhook.ServiceInterface) *WebhookHandler {
	return &WebhookHandler{
		webhookService: webhookService,
	}
}

// CreateWebhook handles POST /api/v1/webhooks
func (h *WebhookHandler) CreateWebhook(c *gin.Context) {
	tenantID, exists := middleware.GetTenantID(c)
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_required", "message": "Tenant ID is required"})
		return
	}

	var req webhook.CreateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_request",
			"Request validation failed", middleware.FormatValidationErrors(err))
		return
	}

	w, err := h.webhookService.CreateWebhook(c.Request.Context(), tenantID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "creation_failed", "message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, w)
}

// GetWebhook handles GET /api/v1/webhooks/:id
func (h *WebhookHandler) GetWebhook(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id", "message": "Invalid webhook ID"})
		return
	}

	w, err := h.webhookService.GetWebhook(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Webhook not found"})
		return
	}

	c.JSON(http.StatusOK, w)
}

// ListWebhooks handles GET /api/v1/webhooks
func (h *WebhookHandler) ListWebhooks(c *gin.Context) {
	tenantID, exists := middleware.GetTenantID(c)
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_required", "message": "Tenant ID is required"})
		return
	}

	webhooks, err := h.webhookService.GetWebhooksByTenant(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "list_failed", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, webhooks)
}

// UpdateWebhook handles PUT /api/v1/webhooks/:id
func (h *WebhookHandler) UpdateWebhook(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id", "message": "Invalid webhook ID"})
		return
	}

	var req webhook.UpdateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest, "invalid_request",
			"Request validation failed", middleware.FormatValidationErrors(err))
		return
	}

	w, err := h.webhookService.UpdateWebhook(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update_failed", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, w)
}

// DeleteWebhook handles DELETE /api/v1/webhooks/:id
func (h *WebhookHandler) DeleteWebhook(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id", "message": "Invalid webhook ID"})
		return
	}

	if err := h.webhookService.DeleteWebhook(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete_failed", "message": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// GetDeliveries handles GET /api/v1/webhooks/:id/deliveries
func (h *WebhookHandler) GetDeliveries(c *gin.Context) {
	idStr := c.Param("id")
	webhookID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id", "message": "Invalid webhook ID"})
		return
	}

	// Parse pagination parameters
	limit := 50
	offset := 0
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
			if limit > 100 {
				limit = 100
			}
		}
	}
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	deliveries, total, err := h.webhookService.GetDeliveriesByWebhook(c.Request.Context(), webhookID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "list_failed", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"deliveries": deliveries,
		"total":      total,
		"limit":      limit,
		"offset":     offset,
	})
}

// GetDelivery handles GET /api/v1/webhooks/:id/deliveries/:delivery_id
func (h *WebhookHandler) GetDelivery(c *gin.Context) {
	deliveryIDStr := c.Param("delivery_id")
	deliveryID, err := uuid.Parse(deliveryIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id", "message": "Invalid delivery ID"})
		return
	}

	delivery, err := h.webhookService.GetDeliveryByID(c.Request.Context(), deliveryID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Delivery not found"})
		return
	}

	c.JSON(http.StatusOK, delivery)
}

