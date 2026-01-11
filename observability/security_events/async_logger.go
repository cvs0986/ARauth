package security_events

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

// AsyncLogger implements the Logger interface with async, batched writes
type AsyncLogger struct {
	repo          Repository
	logger        *zap.Logger
	eventChan     chan *SecurityEvent
	batchSize     int
	flushInterval time.Duration
	wg            sync.WaitGroup
	stopChan      chan struct{}
	once          sync.Once
}

// NewAsyncLogger creates a new async security event logger
func NewAsyncLogger(repo Repository, logger *zap.Logger, batchSize int, flushInterval time.Duration) *AsyncLogger {
	if batchSize <= 0 {
		batchSize = 100 // default batch size
	}
	if flushInterval <= 0 {
		flushInterval = 5 * time.Second // default flush interval
	}

	l := &AsyncLogger{
		repo:          repo,
		logger:        logger,
		eventChan:     make(chan *SecurityEvent, batchSize*2), // buffer for 2 batches
		batchSize:     batchSize,
		flushInterval: flushInterval,
		stopChan:      make(chan struct{}),
	}

	// Start background worker
	l.wg.Add(1)
	go l.worker()

	return l
}

// LogEvent logs a security event asynchronously
func (l *AsyncLogger) LogEvent(ctx context.Context, event *SecurityEvent) error {
	select {
	case l.eventChan <- event:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		// Channel full, log warning and drop event
		l.logger.Warn("Security event channel full, dropping event",
			zap.String("event_type", string(event.EventType)),
			zap.String("severity", string(event.Severity)),
		)
		return nil
	}
}

// GetEvents retrieves security events with filters
func (l *AsyncLogger) GetEvents(ctx context.Context, filters EventFilters) ([]*SecurityEvent, error) {
	return l.repo.Find(ctx, filters)
}

// GetEventsByType retrieves events of a specific type
func (l *AsyncLogger) GetEventsByType(ctx context.Context, eventType EventType, limit int) ([]*SecurityEvent, error) {
	filters := EventFilters{
		EventType: &eventType,
		Limit:     limit,
	}
	return l.repo.Find(ctx, filters)
}

// GetEventsBySeverity retrieves events of a specific severity
func (l *AsyncLogger) GetEventsBySeverity(ctx context.Context, severity Severity, limit int) ([]*SecurityEvent, error) {
	filters := EventFilters{
		Severity: &severity,
		Limit:    limit,
	}
	return l.repo.Find(ctx, filters)
}

// GetRecentEvents retrieves events since a specific time
func (l *AsyncLogger) GetRecentEvents(ctx context.Context, since time.Time, limit int) ([]*SecurityEvent, error) {
	filters := EventFilters{
		Since: &since,
		Limit: limit,
	}
	return l.repo.Find(ctx, filters)
}

// Close gracefully shuts down the logger
func (l *AsyncLogger) Close() error {
	l.once.Do(func() {
		close(l.stopChan)
		l.wg.Wait()
		close(l.eventChan)
	})
	return nil
}

// worker processes events in batches
func (l *AsyncLogger) worker() {
	defer l.wg.Done()

	ticker := time.NewTicker(l.flushInterval)
	defer ticker.Stop()

	batch := make([]*SecurityEvent, 0, l.batchSize)

	flush := func() {
		if len(batch) == 0 {
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := l.repo.CreateBatch(ctx, batch); err != nil {
			l.logger.Error("Failed to write security event batch",
				zap.Error(err),
				zap.Int("batch_size", len(batch)),
			)
		} else {
			l.logger.Debug("Wrote security event batch",
				zap.Int("batch_size", len(batch)),
			)
		}

		// Clear batch
		batch = batch[:0]
	}

	for {
		select {
		case event, ok := <-l.eventChan:
			if !ok {
				// Channel closed, flush remaining events
				flush()
				return
			}

			batch = append(batch, event)

			// Flush if batch is full
			if len(batch) >= l.batchSize {
				flush()
			}

		case <-ticker.C:
			// Periodic flush
			flush()

		case <-l.stopChan:
			// Drain remaining events
			for len(l.eventChan) > 0 {
				event := <-l.eventChan
				batch = append(batch, event)
			}
			flush()
			return
		}
	}
}
