package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestValidationMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(ValidationMiddleware())
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestFormatValidationErrors(t *testing.T) {
	// Create a validator instance
	v := validator.New()

	// Test struct with validation tags
	type TestStruct struct {
		Email    string `validate:"required,email"`
		Username string `validate:"required,min=3"`
		Age      int    `validate:"min=18"`
	}

	testData := TestStruct{
		Email:    "invalid-email",
		Username: "ab", // Too short
		Age:      15,   // Too young
	}

	err := v.Struct(testData)
	assert.Error(t, err)

	validationErrors := FormatValidationErrors(err)
	assert.NotEmpty(t, validationErrors)
	assert.Greater(t, len(validationErrors), 0)

	// Check that we have error messages
	for _, ve := range validationErrors {
		assert.NotEmpty(t, ve.Field)
		assert.NotEmpty(t, ve.Message)
	}
}

func TestFormatValidationErrors_NonValidationError(t *testing.T) {
	err := assert.AnError
	validationErrors := FormatValidationErrors(err)

	assert.Len(t, validationErrors, 1)
	assert.Equal(t, "general", validationErrors[0].Field)
	assert.NotEmpty(t, validationErrors[0].Message)
}

func TestGetValidationMessage(t *testing.T) {
	// This is tested indirectly through FormatValidationErrors
	// but we can test the message formatting logic
	v := validator.New()

	type TestStruct struct {
		Email string `validate:"required,email"`
		Name  string `validate:"required,min=5,max=10"`
		ID    string `validate:"uuid"`
	}

	tests := []struct {
		name      string
		data      TestStruct
		expectTag string
	}{
		{
			name:      "required field",
			data:      TestStruct{},
			expectTag: "required",
		},
		{
			name:      "email validation",
			data:      TestStruct{Email: "invalid"},
			expectTag: "email",
		},
		{
			name:      "min length",
			data:      TestStruct{Name: "ab"},
			expectTag: "min",
		},
		{
			name:      "max length",
			data:      TestStruct{Name: "toolongname"},
			expectTag: "max",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Struct(tt.data)
			if err != nil {
				validationErrors := FormatValidationErrors(err)
				assert.NotEmpty(t, validationErrors)
			}
		})
	}
}

func TestRespondWithError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		validationErrors := []ValidationError{
			{Field: "email", Message: "email is required"},
			{Field: "username", Message: "username must be at least 3 characters"},
		}
		RespondWithError(c, http.StatusBadRequest, "validation_error", "Validation failed", validationErrors)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "validation_error")
	assert.Contains(t, w.Body.String(), "Validation failed")
	assert.Contains(t, w.Body.String(), "email")
}

