package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/arauth-identity/iam/api/middleware"
	"github.com/arauth-identity/iam/auth/introspection"
)

// IntrospectionHandler handles token introspection requests (RFC 7662)
type IntrospectionHandler struct {
	introspectionService introspection.ServiceInterface
}

// NewIntrospectionHandler creates a new token introspection handler
func NewIntrospectionHandler(introspectionService introspection.ServiceInterface) *IntrospectionHandler {
	return &IntrospectionHandler{
		introspectionService: introspectionService,
	}
}

// IntrospectToken handles POST /api/v1/introspect (RFC 7662)
func (h *IntrospectionHandler) IntrospectToken(c *gin.Context) {
	var req struct {
		Token         string `json:"token" form:"token" binding:"required"`
		TokenTypeHint string `json:"token_type_hint,omitempty" form:"token_type_hint"`
	}

	// Support both JSON and form-encoded requests (RFC 7662 allows both)
	if c.ContentType() == "application/x-www-form-urlencoded" {
		if err := c.ShouldBind(&req); err != nil {
			middleware.RespondWithError(c, http.StatusBadRequest, "invalid_request",
				"Request validation failed", middleware.FormatValidationErrors(err))
			return
		}
	} else {
		if err := c.ShouldBindJSON(&req); err != nil {
			middleware.RespondWithError(c, http.StatusBadRequest, "invalid_request",
				"Request validation failed", middleware.FormatValidationErrors(err))
			return
		}
	}

	// Introspect the token
	tokenInfo, err := h.introspectionService.IntrospectToken(c.Request.Context(), req.Token, req.TokenTypeHint)
	if err != nil {
		middleware.RespondWithError(c, http.StatusInternalServerError, "introspection_error",
			"Failed to introspect token", err.Error())
		return
	}

	// Return token info (RFC 7662 response)
	c.JSON(http.StatusOK, tokenInfo)
}

