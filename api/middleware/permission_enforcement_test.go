package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestRequirePermission_MissingPermission tests that missing permission returns 403
func TestRequirePermission_MissingPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	router := gin.New()
	router.Use(func(c *gin.Context) {
		// Simulate authenticated user with no permissions
		c.Set("user_permissions", []string{})
		c.Next()
	})
	router.GET("/test", RequirePermission("users", "read"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "forbidden")
}

// TestRequirePermission_CorrectPermission tests that correct permission returns 200
func TestRequirePermission_CorrectPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	router := gin.New()
	router.Use(func(c *gin.Context) {
		// Simulate authenticated user with correct permission
		c.Set("user_permissions", []string{"users:read"})
		c.Next()
	})
	router.GET("/test", RequirePermission("users", "read"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "success")
}

// TestRequirePermission_MultiplePermissions tests that user with multiple permissions can access
func TestRequirePermission_MultiplePermissions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	router := gin.New()
	router.Use(func(c *gin.Context) {
		// Simulate authenticated user with multiple permissions
		c.Set("user_permissions", []string{"users:read", "users:update", "roles:create"})
		c.Next()
	})
	router.GET("/test", RequirePermission("users", "update"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestRequirePermission_WrongPermission tests that wrong permission returns 403
func TestRequirePermission_WrongPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	router := gin.New()
	router.Use(func(c *gin.Context) {
		// Simulate authenticated user with wrong permission
		c.Set("user_permissions", []string{"users:read"})
		c.Next()
	})
	router.GET("/test", RequirePermission("users", "delete"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// TestRequirePermission_NoPermissionsContext tests that missing permissions context returns 403
func TestRequirePermission_NoPermissionsContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	router := gin.New()
	// No permissions set in context
	router.GET("/test", RequirePermission("users", "read"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "Permission information not available")
}

// TestRequirePermission_GranularPermissions tests granular permission enforcement
func TestRequirePermission_GranularPermissions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name             string
		userPermissions  []string
		requiredResource string
		requiredAction   string
		expectedStatus   int
	}{
		{
			name:             "users:roles:assign granted",
			userPermissions:  []string{"users:roles:assign"},
			requiredResource: "users",
			requiredAction:   "roles:assign",
			expectedStatus:   http.StatusOK,
		},
		{
			name:             "users:roles:assign denied (has users:read)",
			userPermissions:  []string{"users:read"},
			requiredResource: "users",
			requiredAction:   "roles:assign",
			expectedStatus:   http.StatusForbidden,
		},
		{
			name:             "users:identities:link granted",
			userPermissions:  []string{"users:identities:link"},
			requiredResource: "users",
			requiredAction:   "identities:link",
			expectedStatus:   http.StatusOK,
		},
		{
			name:             "users:impersonate granted",
			userPermissions:  []string{"users:impersonate"},
			requiredResource: "users",
			requiredAction:   "impersonate",
			expectedStatus:   http.StatusOK,
		},
		{
			name:             "federation:create granted",
			userPermissions:  []string{"federation:create"},
			requiredResource: "federation",
			requiredAction:   "create",
			expectedStatus:   http.StatusOK,
		},
		{
			name:             "oauth:scopes:create granted",
			userPermissions:  []string{"oauth:scopes:create"},
			requiredResource: "oauth",
			requiredAction:   "scopes:create",
			expectedStatus:   http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set("user_permissions", tt.userPermissions)
				c.Next()
			})
			router.GET("/test", RequirePermission(tt.requiredResource, tt.requiredAction), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestRequireSystemPermission_TenantUserDenied tests that tenant users cannot access system routes
func TestRequireSystemPermission_TenantUserDenied(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	router := gin.New()
	router.Use(func(c *gin.Context) {
		// Simulate TENANT user (not SYSTEM)
		c.Set("principal_type", "TENANT")
		c.Set("user_permissions", []string{"tenant:create"})
		c.Next()
	})
	router.GET("/test", RequireSystemPermission("tenant", "create"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// TestPermissionEnforcement_CriticalOperations tests critical operations require specific permissions
func TestPermissionEnforcement_CriticalOperations(t *testing.T) {
	gin.SetMode(gin.TestMode)

	criticalOps := []struct {
		operation string
		resource  string
		action    string
	}{
		{"User deletion", "users", "delete"},
		{"Role deletion", "roles", "delete"},
		{"Permission creation", "permissions", "create"},
		{"Impersonation", "users", "impersonate"},
		{"Federation config", "federation", "update"},
		{"OAuth scope deletion", "oauth", "scopes:delete"},
	}

	for _, op := range criticalOps {
		t.Run(op.operation+" requires permission", func(t *testing.T) {
			router := gin.New()
			router.Use(func(c *gin.Context) {
				// User has no permissions
				c.Set("user_permissions", []string{})
				c.Next()
			})
			router.POST("/test", RequirePermission(op.resource, op.action), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/test", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusForbidden, w.Code, "%s should be forbidden without permission", op.operation)
		})
	}
}
