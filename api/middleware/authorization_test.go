package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestRequirePermission(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userPermissions []string
		requiredPerm   string
		expectedStatus int
	}{
		{
			name:           "user has required permission",
			userPermissions: []string{"users:read", "users:write"},
			requiredPerm:   "users:read",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "user lacks required permission",
			userPermissions: []string{"users:read"},
			requiredPerm:   "users:write",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "no permissions in context",
			userPermissions: nil,
			requiredPerm:   "users:read",
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			if tt.userPermissions != nil {
				c.Set("user_permissions", tt.userPermissions)
			}
			c.Next()
		})
			// RequirePermission takes resource and action
			parts := strings.Split(tt.requiredPerm, ":")
			resource := parts[0]
			action := parts[1]
			router.GET("/test", RequirePermission(resource, action), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			req, _ := http.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestHasPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userPermissions []string
		requiredPerm   string
		expectedResult bool
	}{
		{
			name:           "user has required permission",
			userPermissions: []string{"users:read", "users:write"},
			requiredPerm:   "users:read",
			expectedResult: true,
		},
		{
			name:           "user lacks required permission",
			userPermissions: []string{"users:read"},
			requiredPerm:   "users:write",
			expectedResult: false,
		},
		{
			name:           "no permissions in context",
			userPermissions: nil,
			requiredPerm:   "users:read",
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/test", func(c *gin.Context) {
				if tt.userPermissions != nil {
					c.Set("user_permissions", tt.userPermissions)
				}
				// HasPermission takes resource and action
				parts := strings.Split(tt.requiredPerm, ":")
				resource := parts[0]
				action := parts[1]
				result := HasPermission(c, resource, action)
				assert.Equal(t, tt.expectedResult, result)
				c.JSON(http.StatusOK, gin.H{"has_permission": result})
			})

			req, _ := http.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

func TestGetTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		setupFunc   func(*gin.Context)
		expectFound bool
		expectID    uuid.UUID
	}{
		{
			name: "tenant ID in context",
			setupFunc: func(c *gin.Context) {
				tenantID := uuid.New()
				c.Set("tenant_id", tenantID)
			},
			expectFound: true,
		},
		{
			name:        "no tenant ID in context",
			setupFunc:   func(c *gin.Context) {},
			expectFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/test", func(c *gin.Context) {
				tt.setupFunc(c)
				tenantID, found := GetTenantID(c)
				assert.Equal(t, tt.expectFound, found)
				if tt.expectFound {
					assert.NotEqual(t, uuid.Nil, tenantID)
				}
				c.JSON(http.StatusOK, gin.H{"found": found})
			})

			req, _ := http.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

