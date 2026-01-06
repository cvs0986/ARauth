package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationMiddleware returns a validation middleware
func ValidationMiddleware() gin.HandlerFunc {
	// This middleware can be used to validate request bodies
	// Individual handlers will use ShouldBindJSON which includes validation
	return func(c *gin.Context) {
		c.Next()
	}
}

// FormatValidationErrors formats validation errors for response
func FormatValidationErrors(err error) []ValidationError {
	var errors []ValidationError

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validationErrors {
			errors = append(errors, ValidationError{
				Field:   fieldError.Field(),
				Message: getValidationMessage(fieldError),
			})
		}
	} else {
		errors = append(errors, ValidationError{
			Field:   "general",
			Message: err.Error(),
		})
	}

	return errors
}

// getValidationMessage returns a human-readable validation message
func getValidationMessage(fieldError validator.FieldError) string {
	switch fieldError.Tag() {
	case "required":
		return fieldError.Field() + " is required"
	case "email":
		return fieldError.Field() + " must be a valid email"
	case "min":
		return fieldError.Field() + " must be at least " + fieldError.Param() + " characters"
	case "max":
		return fieldError.Field() + " must be at most " + fieldError.Param() + " characters"
	case "uuid":
		return fieldError.Field() + " must be a valid UUID"
	default:
		return fieldError.Field() + " is invalid"
	}
}

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Error   string            `json:"error"`
	Message string            `json:"message"`
	Details []ValidationError  `json:"details,omitempty"`
}

// RespondWithError sends a standardized error response
func RespondWithError(c *gin.Context, statusCode int, errorCode string, message string, details []ValidationError) {
	c.JSON(statusCode, ErrorResponse{
		Error:   errorCode,
		Message: message,
		Details: details,
	})
}

