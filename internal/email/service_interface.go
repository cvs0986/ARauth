package email

import (
	"context"
)

// ServiceInterface defines the interface for email sending
type ServiceInterface interface {
	// SendInvitationEmail sends an invitation email to a user
	SendInvitationEmail(ctx context.Context, to string, invitationToken string, tenantName string, expiresAt string) error

	// SendPasswordResetEmail sends a password reset email
	SendPasswordResetEmail(ctx context.Context, to string, resetToken string) error

	// SendWelcomeEmail sends a welcome email to a new user
	SendWelcomeEmail(ctx context.Context, to string, username string) error
}

// NoOpEmailService is a no-op implementation for development/testing
type NoOpEmailService struct{}

// NewNoOpEmailService creates a new no-op email service
func NewNoOpEmailService() ServiceInterface {
	return &NoOpEmailService{}
}

// SendInvitationEmail logs the invitation email (no-op)
func (s *NoOpEmailService) SendInvitationEmail(ctx context.Context, to string, invitationToken string, tenantName string, expiresAt string) error {
	// No-op: In production, this would send an actual email
	// For now, we'll just log it or return nil
	return nil
}

// SendPasswordResetEmail logs the password reset email (no-op)
func (s *NoOpEmailService) SendPasswordResetEmail(ctx context.Context, to string, resetToken string) error {
	// No-op: In production, this would send an actual email
	return nil
}

// SendWelcomeEmail logs the welcome email (no-op)
func (s *NoOpEmailService) SendWelcomeEmail(ctx context.Context, to string, username string) error {
	// No-op: In production, this would send an actual email
	return nil
}

