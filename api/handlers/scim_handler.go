package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/arauth-identity/iam/api/middleware"
	"github.com/arauth-identity/iam/identity/models"
	"github.com/arauth-identity/iam/identity/scim"
)

// SCIMHandler handles SCIM 2.0 API requests
type SCIMHandler struct {
	provisioningService scim.ProvisioningServiceInterface
	tokenService        scim.TokenServiceInterface
}

// NewSCIMHandler creates a new SCIM handler
func NewSCIMHandler(
	provisioningService scim.ProvisioningServiceInterface,
	tokenService scim.TokenServiceInterface,
) *SCIMHandler {
	return &SCIMHandler{
		provisioningService: provisioningService,
		tokenService:        tokenService,
	}
}

// getTenantIDFromContext extracts tenant ID from SCIM token context
func (h *SCIMHandler) getTenantIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	tenantID, exists := c.Get("scim_tenant_id")
	if !exists {
		return uuid.Nil, false
	}
	return tenantID.(uuid.UUID), true
}

// CreateUser handles POST /scim/v2/Users
func (h *SCIMHandler) CreateUser(c *gin.Context) {
	tenantID, ok := h.getTenantIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, models.SCIMError{
			Schemas: []string{"urn:ietf:params:scim:api:messages:2.0:Error"},
			Detail:  "Tenant context required",
			Status:  "401",
		})
		return
	}

	var scimUser models.SCIMUser
	if err := c.ShouldBindJSON(&scimUser); err != nil {
		c.JSON(http.StatusBadRequest, models.SCIMError{
			Schemas: []string{"urn:ietf:params:scim:api:messages:2.0:Error"},
			Detail:  "Invalid request body",
			Status:  "400",
		})
		return
	}

	createdUser, err := h.provisioningService.CreateUser(c.Request.Context(), tenantID, &scimUser)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.SCIMError{
			Schemas: []string{"urn:ietf:params:scim:api:messages:2.0:Error"},
			Detail:  err.Error(),
			Status:  "400",
		})
		return
	}

	c.JSON(http.StatusCreated, createdUser)
}

// GetUser handles GET /scim/v2/Users/:id
func (h *SCIMHandler) GetUser(c *gin.Context) {
	tenantID, ok := h.getTenantIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, models.SCIMError{
			Schemas: []string{"urn:ietf:params:scim:api:messages:2.0:Error"},
			Detail:  "Tenant context required",
			Status:  "401",
		})
		return
	}

	userID := c.Param("id")
	scimUser, err := h.provisioningService.GetUser(c.Request.Context(), tenantID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.SCIMError{
			Schemas: []string{"urn:ietf:params:scim:api:messages:2.0:Error"},
			Detail:  "User not found",
			Status:  "404",
		})
		return
	}

	c.JSON(http.StatusOK, scimUser)
}

// ListUsers handles GET /scim/v2/Users
func (h *SCIMHandler) ListUsers(c *gin.Context) {
	tenantID, ok := h.getTenantIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, models.SCIMError{
			Schemas: []string{"urn:ietf:params:scim:api:messages:2.0:Error"},
			Detail:  "Tenant context required",
			Status:  "401",
		})
		return
	}

	// Parse SCIM query parameters
	filters := &scim.UserFilters{
		Filter:     c.Query("filter"),
		StartIndex: 1,
		Count:      100,
	}

	if startIndexStr := c.Query("startIndex"); startIndexStr != "" {
		if idx, err := strconv.Atoi(startIndexStr); err == nil && idx > 0 {
			filters.StartIndex = idx
		}
	}

	if countStr := c.Query("count"); countStr != "" {
		if cnt, err := strconv.Atoi(countStr); err == nil && cnt > 0 {
			if cnt > 100 {
				cnt = 100
			}
			filters.Count = cnt
		}
	}

	users, total, err := h.provisioningService.ListUsers(c.Request.Context(), tenantID, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.SCIMError{
			Schemas: []string{"urn:ietf:params:scim:api:messages:2.0:Error"},
			Detail:  err.Error(),
			Status:  "500",
		})
		return
	}

	response := models.SCIMListResponse{
		Schemas:      []string{"urn:ietf:params:scim:api:messages:2.0:ListResponse"},
		TotalResults: total,
		ItemsPerPage: filters.Count,
		StartIndex:   filters.StartIndex,
		Resources:    make([]interface{}, len(users)),
	}

	for i, user := range users {
		response.Resources[i] = user
	}

	c.JSON(http.StatusOK, response)
}

// UpdateUser handles PUT /scim/v2/Users/:id
func (h *SCIMHandler) UpdateUser(c *gin.Context) {
	tenantID, ok := h.getTenantIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, models.SCIMError{
			Schemas: []string{"urn:ietf:params:scim:api:messages:2.0:Error"},
			Detail:  "Tenant context required",
			Status:  "401",
		})
		return
	}

	userID := c.Param("id")

	var scimUser models.SCIMUser
	if err := c.ShouldBindJSON(&scimUser); err != nil {
		c.JSON(http.StatusBadRequest, models.SCIMError{
			Schemas: []string{"urn:ietf:params:scim:api:messages:2.0:Error"},
			Detail:  "Invalid request body",
			Status:  "400",
		})
		return
	}

	updatedUser, err := h.provisioningService.UpdateUser(c.Request.Context(), tenantID, userID, &scimUser)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.SCIMError{
			Schemas: []string{"urn:ietf:params:scim:api:messages:2.0:Error"},
			Detail:  err.Error(),
			Status:  "400",
		})
		return
	}

	c.JSON(http.StatusOK, updatedUser)
}

// DeleteUser handles DELETE /scim/v2/Users/:id
func (h *SCIMHandler) DeleteUser(c *gin.Context) {
	tenantID, ok := h.getTenantIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, models.SCIMError{
			Schemas: []string{"urn:ietf:params:scim:api:messages:2.0:Error"},
			Detail:  "Tenant context required",
			Status:  "401",
		})
		return
	}

	userID := c.Param("id")

	err := h.provisioningService.DeleteUser(c.Request.Context(), tenantID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.SCIMError{
			Schemas: []string{"urn:ietf:params:scim:api:messages:2.0:Error"},
			Detail:  "User not found",
			Status:  "404",
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// CreateGroup handles POST /scim/v2/Groups
func (h *SCIMHandler) CreateGroup(c *gin.Context) {
	tenantID, ok := h.getTenantIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, models.SCIMError{
			Schemas: []string{"urn:ietf:params:scim:api:messages:2.0:Error"},
			Detail:  "Tenant context required",
			Status:  "401",
		})
		return
	}

	var scimGroup models.SCIMGroup
	if err := c.ShouldBindJSON(&scimGroup); err != nil {
		c.JSON(http.StatusBadRequest, models.SCIMError{
			Schemas: []string{"urn:ietf:params:scim:api:messages:2.0:Error"},
			Detail:  "Invalid request body",
			Status:  "400",
		})
		return
	}

	createdGroup, err := h.provisioningService.CreateGroup(c.Request.Context(), tenantID, &scimGroup)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.SCIMError{
			Schemas: []string{"urn:ietf:params:scim:api:messages:2.0:Error"},
			Detail:  err.Error(),
			Status:  "400",
		})
		return
	}

	c.JSON(http.StatusCreated, createdGroup)
}

// GetGroup handles GET /scim/v2/Groups/:id
func (h *SCIMHandler) GetGroup(c *gin.Context) {
	tenantID, ok := h.getTenantIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, models.SCIMError{
			Schemas: []string{"urn:ietf:params:scim:api:messages:2.0:Error"},
			Detail:  "Tenant context required",
			Status:  "401",
		})
		return
	}

	groupID := c.Param("id")
	scimGroup, err := h.provisioningService.GetGroup(c.Request.Context(), tenantID, groupID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.SCIMError{
			Schemas: []string{"urn:ietf:params:scim:api:messages:2.0:Error"},
			Detail:  "Group not found",
			Status:  "404",
		})
		return
	}

	c.JSON(http.StatusOK, scimGroup)
}

// ListGroups handles GET /scim/v2/Groups
func (h *SCIMHandler) ListGroups(c *gin.Context) {
	tenantID, ok := h.getTenantIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, models.SCIMError{
			Schemas: []string{"urn:ietf:params:scim:api:messages:2.0:Error"},
			Detail:  "Tenant context required",
			Status:  "401",
		})
		return
	}

	filters := &scim.GroupFilters{
		Filter:     c.Query("filter"),
		StartIndex: 1,
		Count:      100,
	}

	if startIndexStr := c.Query("startIndex"); startIndexStr != "" {
		if idx, err := strconv.Atoi(startIndexStr); err == nil && idx > 0 {
			filters.StartIndex = idx
		}
	}

	if countStr := c.Query("count"); countStr != "" {
		if cnt, err := strconv.Atoi(countStr); err == nil && cnt > 0 {
			if cnt > 100 {
				cnt = 100
			}
			filters.Count = cnt
		}
	}

	groups, total, err := h.provisioningService.ListGroups(c.Request.Context(), tenantID, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.SCIMError{
			Schemas: []string{"urn:ietf:params:scim:api:messages:2.0:Error"},
			Detail:  err.Error(),
			Status:  "500",
		})
		return
	}

	response := models.SCIMListResponse{
		Schemas:      []string{"urn:ietf:params:scim:api:messages:2.0:ListResponse"},
		TotalResults: total,
		ItemsPerPage: filters.Count,
		StartIndex:   filters.StartIndex,
		Resources:    make([]interface{}, len(groups)),
	}

	for i, group := range groups {
		response.Resources[i] = group
	}

	c.JSON(http.StatusOK, response)
}

// UpdateGroup handles PUT /scim/v2/Groups/:id
func (h *SCIMHandler) UpdateGroup(c *gin.Context) {
	tenantID, ok := h.getTenantIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, models.SCIMError{
			Schemas: []string{"urn:ietf:params:scim:api:messages:2.0:Error"},
			Detail:  "Tenant context required",
			Status:  "401",
		})
		return
	}

	groupID := c.Param("id")

	var scimGroup models.SCIMGroup
	if err := c.ShouldBindJSON(&scimGroup); err != nil {
		c.JSON(http.StatusBadRequest, models.SCIMError{
			Schemas: []string{"urn:ietf:params:scim:api:messages:2.0:Error"},
			Detail:  "Invalid request body",
			Status:  "400",
		})
		return
	}

	updatedGroup, err := h.provisioningService.UpdateGroup(c.Request.Context(), tenantID, groupID, &scimGroup)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.SCIMError{
			Schemas: []string{"urn:ietf:params:scim:api:messages:2.0:Error"},
			Detail:  err.Error(),
			Status:  "400",
		})
		return
	}

	c.JSON(http.StatusOK, updatedGroup)
}

// DeleteGroup handles DELETE /scim/v2/Groups/:id
func (h *SCIMHandler) DeleteGroup(c *gin.Context) {
	tenantID, ok := h.getTenantIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, models.SCIMError{
			Schemas: []string{"urn:ietf:params:scim:api:messages:2.0:Error"},
			Detail:  "Tenant context required",
			Status:  "401",
		})
		return
	}

	groupID := c.Param("id")

	err := h.provisioningService.DeleteGroup(c.Request.Context(), tenantID, groupID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.SCIMError{
			Schemas: []string{"urn:ietf:params:scim:api:messages:2.0:Error"},
			Detail:  "Group not found",
			Status:  "404",
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// BulkOperations handles POST /scim/v2/Bulk
func (h *SCIMHandler) BulkOperations(c *gin.Context) {
	tenantID, ok := h.getTenantIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, models.SCIMError{
			Schemas: []string{"urn:ietf:params:scim:api:messages:2.0:Error"},
			Detail:  "Tenant context required",
			Status:  "401",
		})
		return
	}

	var bulkRequest struct {
		Schemas    []string              `json:"schemas"`
		FailOnErrors int                `json:"failOnErrors,omitempty"`
		Operations []scim.BulkOperation `json:"Operations"`
	}

	if err := c.ShouldBindJSON(&bulkRequest); err != nil {
		c.JSON(http.StatusBadRequest, models.SCIMError{
			Schemas: []string{"urn:ietf:params:scim:api:messages:2.0:Error"},
			Detail:  "Invalid request body",
			Status:  "400",
		})
		return
	}

	response, err := h.provisioningService.BulkCreate(c.Request.Context(), tenantID, bulkRequest.Operations)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.SCIMError{
			Schemas: []string{"urn:ietf:params:scim:api:messages:2.0:Error"},
			Detail:  err.Error(),
			Status:  "400",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ServiceProviderConfig handles GET /scim/v2/ServiceProviderConfig
func (h *SCIMHandler) ServiceProviderConfig(c *gin.Context) {
	config := gin.H{
		"schemas":    []string{"urn:ietf:params:scim:schemas:core:2.0:ServiceProviderConfig"},
		"patch":      gin.H{"supported": true},
		"bulk":       gin.H{"supported": true, "maxOperations": 100, "maxPayloadSize": 1048576},
		"filter":     gin.H{"supported": true, "maxResults": 200},
		"changePassword": gin.H{"supported": false},
		"sort":       gin.H{"supported": false},
		"etag":       gin.H{"supported": false},
		"authenticationSchemes": []gin.H{
			{
				"type":        "oauthbearertoken",
				"name":        "OAuth Bearer Token",
				"description": "Authentication using OAuth 2.0 Bearer tokens",
			},
		},
	}

	c.JSON(http.StatusOK, config)
}

// ResourceTypes handles GET /scim/v2/ResourceTypes
func (h *SCIMHandler) ResourceTypes(c *gin.Context) {
	resourceTypes := []gin.H{
		{
			"schemas":     []string{"urn:ietf:params:scim:schemas:core:2.0:ResourceType"},
			"id":          "User",
			"name":        "User",
			"endpoint":    "/Users",
			"description": "User Account",
			"schema":      "urn:ietf:params:scim:schemas:core:2.0:User",
			"meta": gin.H{
				"location":     "/v2/ResourceTypes/User",
				"resourceType": "ResourceType",
			},
		},
		{
			"schemas":     []string{"urn:ietf:params:scim:schemas:core:2.0:ResourceType"},
			"id":          "Group",
			"name":        "Group",
			"endpoint":    "/Groups",
			"description": "Group",
			"schema":      "urn:ietf:params:scim:schemas:core:2.0:Group",
			"meta": gin.H{
				"location":     "/v2/ResourceTypes/Group",
				"resourceType": "ResourceType",
			},
		},
	}

	c.JSON(http.StatusOK, resourceTypes)
}

// Schemas handles GET /scim/v2/Schemas
func (h *SCIMHandler) Schemas(c *gin.Context) {
	schemas := []gin.H{
		{
			"id":          "urn:ietf:params:scim:schemas:core:2.0:User",
			"name":        "User",
			"description": "User Account",
			"attributes": []gin.H{
				{"name": "userName", "type": "string", "required": true, "caseExact": false},
				{"name": "name", "type": "complex", "required": false},
				{"name": "emails", "type": "complex", "multiValued": true, "required": false},
				{"name": "active", "type": "boolean", "required": false},
			},
		},
		{
			"id":          "urn:ietf:params:scim:schemas:core:2.0:Group",
			"name":        "Group",
			"description": "Group",
			"attributes": []gin.H{
				{"name": "displayName", "type": "string", "required": true},
				{"name": "members", "type": "complex", "multiValued": true, "required": false},
			},
		},
	}

	c.JSON(http.StatusOK, schemas)
}

