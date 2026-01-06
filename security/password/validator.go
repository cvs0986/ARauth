package password

import (
	"fmt"
	"strings"
	"unicode"
)

// Validator provides password validation functionality
type Validator struct {
	MinLength      int
	RequireUpper   bool
	RequireLower   bool
	RequireNumber  bool
	RequireSpecial bool
}

// NewValidator creates a new password validator
func NewValidator(minLength int, requireUpper, requireLower, requireNumber, requireSpecial bool) *Validator {
	return &Validator{
		MinLength:      minLength,
		RequireUpper:   requireUpper,
		RequireLower:   requireLower,
		RequireNumber:  requireNumber,
		RequireSpecial: requireSpecial,
	}
}

// Validate validates a password against the policy
func (v *Validator) Validate(password string, username string) error {
	// Check minimum length
	if len(password) < v.MinLength {
		return fmt.Errorf("password must be at least %d characters long", v.MinLength)
	}

	// Check if password contains username
	if username != "" && strings.Contains(strings.ToLower(password), strings.ToLower(username)) {
		return fmt.Errorf("password cannot contain your username")
	}

	// Check complexity requirements
	var hasUpper, hasLower, hasNumber, hasSpecial bool

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	var errors []string

	if v.RequireUpper && !hasUpper {
		errors = append(errors, "at least one uppercase letter")
	}
	if v.RequireLower && !hasLower {
		errors = append(errors, "at least one lowercase letter")
	}
	if v.RequireNumber && !hasNumber {
		errors = append(errors, "at least one number")
	}
	if v.RequireSpecial && !hasSpecial {
		errors = append(errors, "at least one special character")
	}

	if len(errors) > 0 {
		return fmt.Errorf("password must contain: %s", strings.Join(errors, ", "))
	}

	return nil
}

// CheckCommonPasswords checks if password is in common password list
func (v *Validator) CheckCommonPasswords(password string) bool {
	commonPasswords := []string{
		"password", "123456", "12345678", "123456789", "1234567890",
		"qwerty", "abc123", "password1", "Password1", "admin",
		"letmein", "welcome", "monkey", "1234567", "sunshine",
		"master", "123123", "dragon", "passw0rd", "trustno1",
	}

	passwordLower := strings.ToLower(password)
	for _, common := range commonPasswords {
		if passwordLower == common {
			return true
		}
	}

	return false
}

