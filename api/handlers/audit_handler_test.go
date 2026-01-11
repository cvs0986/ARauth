package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arauth-identity/iam/storage/interfaces"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuditServiceForExport is a specific mock to ensure clean state
// We could reuse MockAuditService from auth_handler_test.go but it's better to be explicit here
// to avoid package level compilation issues if logic differs.
// However, since they are in the same package `handlers`, we can reuse.
// Let's assume we can reuse MockAuditService if it's exported in the package.
// Earlier grep showed it's in auth_handler_test.go package handlers.

func TestAuditHandler_ExportEvents(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tenantID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockService := &MockAuditService{}
		handler := NewAuditHandler(mockService)

		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("tenant_id", tenantID)
			c.Next()
		})
		router.GET("/api/v1/audit/export", handler.ExportEvents)

		csvData := []byte("Header\nRow1")
		filename := "export.csv"

		mockService.On("ExportEvents", mock.Anything, mock.MatchedBy(func(filters *interfaces.AuditEventFilters) bool {
			return filters.TenantID != nil && *filters.TenantID == tenantID
		})).Return(csvData, filename, nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/audit/export?event_type=login", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "text/csv", w.Header().Get("Content-Type"))
		assert.Contains(t, w.Header().Get("Content-Disposition"), filename)
		assert.Equal(t, csvData, w.Body.Bytes())
	})

	t.Run("export_failure", func(t *testing.T) {
		mockService := &MockAuditService{}
		handler := NewAuditHandler(mockService)

		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("tenant_id", tenantID)
			c.Next()
		})
		router.GET("/api/v1/audit/export", handler.ExportEvents)

		mockService.On("ExportEvents", mock.Anything, mock.Anything).Return([]byte(nil), "", assert.AnError)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/audit/export", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
