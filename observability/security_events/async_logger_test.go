package security_events

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// MockRepository implements Repository for testing
type MockRepository struct {
	events      []*SecurityEvent
	createCalls int
	batchCalls  int
	findCalls   int
	countCalls  int
	deleteCalls int
	shouldFail  bool
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		events: make([]*SecurityEvent, 0),
	}
}

func (m *MockRepository) Create(ctx context.Context, event *SecurityEvent) error {
	m.createCalls++
	if m.shouldFail {
		return assert.AnError
	}
	m.events = append(m.events, event)
	return nil
}

func (m *MockRepository) CreateBatch(ctx context.Context, events []*SecurityEvent) error {
	m.batchCalls++
	if m.shouldFail {
		return assert.AnError
	}
	m.events = append(m.events, events...)
	return nil
}

func (m *MockRepository) Find(ctx context.Context, filters EventFilters) ([]*SecurityEvent, error) {
	m.findCalls++
	if m.shouldFail {
		return nil, assert.AnError
	}

	result := make([]*SecurityEvent, 0)
	for _, event := range m.events {
		if filters.EventType != nil && event.EventType != *filters.EventType {
			continue
		}
		if filters.Severity != nil && event.Severity != *filters.Severity {
			continue
		}
		if filters.TenantID != nil && (event.TenantID == nil || *event.TenantID != *filters.TenantID) {
			continue
		}
		if filters.UserID != nil && (event.UserID == nil || *event.UserID != *filters.UserID) {
			continue
		}
		if filters.Since != nil && event.CreatedAt.Before(*filters.Since) {
			continue
		}
		if filters.Until != nil && event.CreatedAt.After(*filters.Until) {
			continue
		}
		result = append(result, event)
	}

	if filters.Limit > 0 && len(result) > filters.Limit {
		result = result[:filters.Limit]
	}

	return result, nil
}

func (m *MockRepository) Count(ctx context.Context, filters EventFilters) (int, error) {
	m.countCalls++
	if m.shouldFail {
		return 0, assert.AnError
	}
	events, _ := m.Find(ctx, filters)
	return len(events), nil
}

func (m *MockRepository) DeleteOlderThan(ctx context.Context, olderThan time.Time) (int, error) {
	m.deleteCalls++
	if m.shouldFail {
		return 0, assert.AnError
	}
	count := 0
	newEvents := make([]*SecurityEvent, 0)
	for _, event := range m.events {
		if event.CreatedAt.Before(olderThan) {
			count++
		} else {
			newEvents = append(newEvents, event)
		}
	}
	m.events = newEvents
	return count, nil
}

func TestNewSecurityEvent(t *testing.T) {
	event := NewSecurityEvent(EventAuthFailure, SeverityCritical)

	assert.NotEqual(t, uuid.Nil, event.ID)
	assert.Equal(t, EventAuthFailure, event.EventType)
	assert.Equal(t, SeverityCritical, event.Severity)
	assert.NotNil(t, event.Details)
	assert.False(t, event.CreatedAt.IsZero())
}

func TestSecurityEvent_BuilderPattern(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()

	event := NewSecurityEvent(EventPermissionDenied, SeverityWarning).
		WithTenant(tenantID).
		WithUser(userID).
		WithIP("192.168.1.100").
		WithResource("/api/v1/users").
		WithAction("DELETE").
		WithResult("denied").
		WithDetail("reason", "insufficient_permissions")

	assert.Equal(t, &tenantID, event.TenantID)
	assert.Equal(t, &userID, event.UserID)
	assert.Equal(t, "192.168.1.100", event.IP)
	assert.Equal(t, "/api/v1/users", event.Resource)
	assert.Equal(t, "DELETE", event.Action)
	assert.Equal(t, "denied", event.Result)
	assert.Equal(t, "insufficient_permissions", event.Details["reason"])
}

func TestAsyncLogger_LogEvent(t *testing.T) {
	repo := NewMockRepository()
	logger, _ := zap.NewDevelopment()
	asyncLogger := NewAsyncLogger(repo, logger, 10, 100*time.Millisecond)
	defer asyncLogger.Close()

	ctx := context.Background()
	event := NewSecurityEvent(EventAuthFailure, SeverityWarning)

	err := asyncLogger.LogEvent(ctx, event)
	require.NoError(t, err)

	// Wait for async processing
	time.Sleep(200 * time.Millisecond)

	assert.Equal(t, 1, repo.batchCalls)
	assert.Equal(t, 1, len(repo.events))
	assert.Equal(t, event.ID, repo.events[0].ID)
}

func TestAsyncLogger_BatchProcessing(t *testing.T) {
	repo := NewMockRepository()
	logger, _ := zap.NewDevelopment()
	asyncLogger := NewAsyncLogger(repo, logger, 5, 1*time.Second)
	defer asyncLogger.Close()

	ctx := context.Background()

	// Log 10 events (should trigger 2 batches of 5)
	for i := 0; i < 10; i++ {
		event := NewSecurityEvent(EventAuthFailure, SeverityWarning)
		err := asyncLogger.LogEvent(ctx, event)
		require.NoError(t, err)
	}

	// Wait for async processing
	time.Sleep(200 * time.Millisecond)

	assert.Equal(t, 2, repo.batchCalls)
	assert.Equal(t, 10, len(repo.events))
}

func TestAsyncLogger_PeriodicFlush(t *testing.T) {
	repo := NewMockRepository()
	logger, _ := zap.NewDevelopment()
	asyncLogger := NewAsyncLogger(repo, logger, 100, 100*time.Millisecond)
	defer asyncLogger.Close()

	ctx := context.Background()

	// Log 3 events (below batch size)
	for i := 0; i < 3; i++ {
		event := NewSecurityEvent(EventAuthFailure, SeverityWarning)
		err := asyncLogger.LogEvent(ctx, event)
		require.NoError(t, err)
	}

	// Wait for periodic flush
	time.Sleep(200 * time.Millisecond)

	assert.Equal(t, 1, repo.batchCalls)
	assert.Equal(t, 3, len(repo.events))
}

func TestAsyncLogger_GetEvents(t *testing.T) {
	repo := NewMockRepository()
	logger, _ := zap.NewDevelopment()
	asyncLogger := NewAsyncLogger(repo, logger, 10, 100*time.Millisecond)
	defer asyncLogger.Close()

	ctx := context.Background()

	// Add events directly to repo
	event1 := NewSecurityEvent(EventAuthFailure, SeverityWarning)
	event2 := NewSecurityEvent(EventPermissionDenied, SeverityCritical)
	repo.events = []*SecurityEvent{event1, event2}

	// Get all events
	events, err := asyncLogger.GetEvents(ctx, EventFilters{})
	require.NoError(t, err)
	assert.Equal(t, 2, len(events))
}

func TestAsyncLogger_GetEventsByType(t *testing.T) {
	repo := NewMockRepository()
	logger, _ := zap.NewDevelopment()
	asyncLogger := NewAsyncLogger(repo, logger, 10, 100*time.Millisecond)
	defer asyncLogger.Close()

	ctx := context.Background()

	// Add events directly to repo
	event1 := NewSecurityEvent(EventAuthFailure, SeverityWarning)
	event2 := NewSecurityEvent(EventPermissionDenied, SeverityCritical)
	event3 := NewSecurityEvent(EventAuthFailure, SeverityInfo)
	repo.events = []*SecurityEvent{event1, event2, event3}

	// Get only auth failures
	events, err := asyncLogger.GetEventsByType(ctx, EventAuthFailure, 10)
	require.NoError(t, err)
	assert.Equal(t, 2, len(events))
	assert.Equal(t, EventAuthFailure, events[0].EventType)
	assert.Equal(t, EventAuthFailure, events[1].EventType)
}

func TestAsyncLogger_GetEventsBySeverity(t *testing.T) {
	repo := NewMockRepository()
	logger, _ := zap.NewDevelopment()
	asyncLogger := NewAsyncLogger(repo, logger, 10, 100*time.Millisecond)
	defer asyncLogger.Close()

	ctx := context.Background()

	// Add events directly to repo
	event1 := NewSecurityEvent(EventAuthFailure, SeverityWarning)
	event2 := NewSecurityEvent(EventPermissionDenied, SeverityCritical)
	event3 := NewSecurityEvent(EventRateLimitExceeded, SeverityCritical)
	repo.events = []*SecurityEvent{event1, event2, event3}

	// Get only critical events
	events, err := asyncLogger.GetEventsBySeverity(ctx, SeverityCritical, 10)
	require.NoError(t, err)
	assert.Equal(t, 2, len(events))
	assert.Equal(t, SeverityCritical, events[0].Severity)
	assert.Equal(t, SeverityCritical, events[1].Severity)
}

func TestAsyncLogger_GetRecentEvents(t *testing.T) {
	repo := NewMockRepository()
	logger, _ := zap.NewDevelopment()
	asyncLogger := NewAsyncLogger(repo, logger, 10, 100*time.Millisecond)
	defer asyncLogger.Close()

	ctx := context.Background()

	// Add events with different timestamps
	now := time.Now()
	event1 := NewSecurityEvent(EventAuthFailure, SeverityWarning)
	event1.CreatedAt = now.Add(-2 * time.Hour)
	event2 := NewSecurityEvent(EventPermissionDenied, SeverityCritical)
	event2.CreatedAt = now.Add(-30 * time.Minute)
	repo.events = []*SecurityEvent{event1, event2}

	// Get events from last hour
	since := now.Add(-1 * time.Hour)
	events, err := asyncLogger.GetRecentEvents(ctx, since, 10)
	require.NoError(t, err)
	assert.Equal(t, 1, len(events))
	assert.Equal(t, event2.ID, events[0].ID)
}

func TestAsyncLogger_Close(t *testing.T) {
	repo := NewMockRepository()
	logger, _ := zap.NewDevelopment()
	asyncLogger := NewAsyncLogger(repo, logger, 10, 100*time.Millisecond)

	ctx := context.Background()

	// Log some events
	for i := 0; i < 3; i++ {
		event := NewSecurityEvent(EventAuthFailure, SeverityWarning)
		asyncLogger.LogEvent(ctx, event)
	}

	// Close should flush remaining events
	err := asyncLogger.Close()
	require.NoError(t, err)

	assert.Equal(t, 3, len(repo.events))
}
