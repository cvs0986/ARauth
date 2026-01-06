package audit

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/nuage-identity/iam/storage/interfaces"
)

// Logger provides audit logging functionality
type Logger struct {
	repo interfaces.AuditRepository
}

// NewLogger creates a new audit logger
func NewLogger(repo interfaces.AuditRepository) *Logger {
	return &Logger{repo: repo}
}

// Log creates an audit log entry
func (l *Logger) Log(ctx context.Context, log *interfaces.AuditLog) error {
	return l.repo.Create(ctx, log)
}

// LogAction creates an audit log for an action
func (l *Logger) LogAction(ctx context.Context, tenantID uuid.UUID, userID *uuid.UUID, action, resource string, req *http.Request, status string, message string, metadata map[string]interface{}) error {
	ipAddress := getClientIP(req)
	userAgent := req.UserAgent()

	log := &interfaces.AuditLog{
		ID:         uuid.New(),
		TenantID:   tenantID,
		UserID:     userID,
		Action:     action,
		Resource:   resource,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Status:     status,
		Message:    message,
		Metadata:   metadata,
		CreatedAt: time.Now(),
	}

	return l.Log(ctx, log)
}

// LogMFAAction logs an MFA-related action
func (l *Logger) LogMFAAction(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, action string, req *http.Request, status string, message string) error {
	return l.LogAction(ctx, tenantID, &userID, action, "mfa", req, status, message, nil)
}

// LogLoginAction logs a login action
func (l *Logger) LogLoginAction(ctx context.Context, tenantID uuid.UUID, userID *uuid.UUID, req *http.Request, status string, message string) error {
	return l.LogAction(ctx, tenantID, userID, "login", "auth", req, status, message, nil)
}

// LogUserAction logs a user management action
func (l *Logger) LogUserAction(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, action string, resourceID string, req *http.Request, status string, message string) error {
	log := &interfaces.AuditLog{
		ID:         uuid.New(),
		TenantID:   tenantID,
		UserID:     &userID,
		Action:     action,
		Resource:   "user",
		ResourceID: &resourceID,
		IPAddress:  getClientIP(req),
		UserAgent:  req.UserAgent(),
		Status:     status,
		Message:    message,
		CreatedAt:  time.Now(),
	}

	return l.Log(ctx, log)
}

// getClientIP extracts the client IP from the request
func getClientIP(req *http.Request) string {
	// Check X-Forwarded-For header
	if ip := req.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}

	// Check X-Real-IP header
	if ip := req.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}

	// Fall back to RemoteAddr
	return req.RemoteAddr
}

