package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTenantRepository is a mock for testing
type MockTenantRepository struct {
	mock.Mock
}

func (m *MockTenantRepository) GetByID(ctx interface{}, id uuid.UUID) (interface{}, error) {
	args := m.Called(ctx, id)
	return args.Get(0), args.Error(1)
}

func (m *MockTenantRepository) GetByDomain(ctx interface{}, domain string) (interface{}, error) {
	args := m.Called(ctx, domain)
	return args.Get(0), args.Error(1)
}

func TestTenantMiddleware_Header(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockTenantRepository)
	tenantID := uuid.New()

	router := gin.New()
	router.Use(TenantMiddleware(mockRepo))
	router.GET("/test", func(c *gin.Context) {
		id, exists := GetTenantID(c)
		assert.True(t, exists)
		assert.Equal(t, tenantID, id)
		c.JSON(http.StatusOK, gin.H{"tenant_id": id.String()})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Tenant-ID", tenantID.String())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		setupFunc      func(*gin.Context)
		expectFound    bool
	}{
		{
			name: "tenant ID in context",
			setupFunc: func(c *gin.Context) {
				c.Set("tenant_id", uuid.New())
			},
			expectFound: true,
		},
		{
			name:           "no tenant ID in context",
			setupFunc:      func(c *gin.Context) {},
			expectFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/test", func(c *gin.Context) {
				tt.setupFunc(c)
				tenantID, ok := RequireTenant(c)
				assert.Equal(t, tt.expectFound, ok)
				if ok {
					assert.NotEqual(t, uuid.Nil, tenantID)
					c.JSON(http.StatusOK, gin.H{"tenant_id": tenantID.String()})
				} else {
					c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_required"})
				}
			})

			req, _ := http.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if tt.expectFound {
				assert.Equal(t, http.StatusOK, w.Code)
			} else {
				assert.Equal(t, http.StatusBadRequest, w.Code)
			}
		})
	}
}

