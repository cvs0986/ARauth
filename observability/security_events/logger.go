package security_events

import (
	"context"
	"time"
)

// Logger defines the interface for security event logging
type Logger interface {
	// LogEvent logs a security event asynchronously
	LogEvent(ctx context.Context, event *SecurityEvent) error

	// GetEvents retrieves security events with filters
	GetEvents(ctx context.Context, filters EventFilters) ([]*SecurityEvent, error)

	// GetEventsByType retrieves events of a specific type
	GetEventsByType(ctx context.Context, eventType EventType, limit int) ([]*SecurityEvent, error)

	// GetEventsBySeverity retrieves events of a specific severity
	GetEventsBySeverity(ctx context.Context, severity Severity, limit int) ([]*SecurityEvent, error)

	// GetRecentEvents retrieves events since a specific time
	GetRecentEvents(ctx context.Context, since time.Time, limit int) ([]*SecurityEvent, error)

	// Close gracefully shuts down the logger
	Close() error
}

// Repository defines the interface for security event persistence
type Repository interface {
	// Create stores a new security event
	Create(ctx context.Context, event *SecurityEvent) error

	// CreateBatch stores multiple security events in a single transaction
	CreateBatch(ctx context.Context, events []*SecurityEvent) error

	// Find retrieves security events with filters
	Find(ctx context.Context, filters EventFilters) ([]*SecurityEvent, error)

	// Count returns the count of events matching filters
	Count(ctx context.Context, filters EventFilters) (int, error)

	// DeleteOlderThan deletes events older than the specified duration
	DeleteOlderThan(ctx context.Context, olderThan time.Time) (int, error)
}
