package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/arauth-identity/iam/identity/models"
	"github.com/arauth-identity/iam/storage/interfaces"
)

// webhookRepository implements WebhookRepository
type webhookRepository struct {
	db *sql.DB
}

// NewWebhookRepository creates a new webhook repository
func NewWebhookRepository(db *sql.DB) interfaces.WebhookRepository {
	return &webhookRepository{db: db}
}

// Create creates a new webhook
func (r *webhookRepository) Create(ctx context.Context, w *models.Webhook) error {
	query := `
		INSERT INTO webhooks (
			id, tenant_id, name, url, secret, enabled, events,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)
	`

	now := time.Now()
	w.CreatedAt = now
	w.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		w.ID,
		w.TenantID,
		w.Name,
		w.URL,
		w.Secret,
		w.Enabled,
		pq.Array(w.Events),
		w.CreatedAt,
		w.UpdatedAt,
	)

	return err
}

// GetByID retrieves a webhook by ID
func (r *webhookRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Webhook, error) {
	query := `
		SELECT id, tenant_id, name, url, secret, enabled, events,
		       created_at, updated_at, deleted_at
		FROM webhooks
		WHERE id = $1 AND deleted_at IS NULL
	`

	var w models.Webhook
	var events pq.StringArray
	var deletedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&w.ID,
		&w.TenantID,
		&w.Name,
		&w.URL,
		&w.Secret,
		&w.Enabled,
		&events,
		&w.CreatedAt,
		&w.UpdatedAt,
		&deletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("webhook not found: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get webhook: %w", err)
	}

	w.Events = []string(events)
	if deletedAt.Valid {
		w.DeletedAt = &deletedAt.Time
	}

	return &w, nil
}

// GetByTenantID retrieves all webhooks for a tenant
func (r *webhookRepository) GetByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*webhook.Webhook, error) {
	query := `
		SELECT id, tenant_id, name, url, secret, enabled, events,
		       created_at, updated_at, deleted_at
		FROM webhooks
		WHERE tenant_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query webhooks: %w", err)
	}
	defer rows.Close()

	var webhooks []*models.Webhook
	for rows.Next() {
		var w models.Webhook
		var events pq.StringArray
		var deletedAt sql.NullTime

		err := rows.Scan(
			&w.ID,
			&w.TenantID,
			&w.Name,
			&w.URL,
			&w.Secret,
			&w.Enabled,
			&events,
			&w.CreatedAt,
			&w.UpdatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan webhook: %w", err)
		}

		w.Events = []string(events)
		if deletedAt.Valid {
			w.DeletedAt = &deletedAt.Time
		}

		webhooks = append(webhooks, &w)
	}

	return webhooks, rows.Err()
}

// GetEnabledByTenantID retrieves all enabled webhooks for a tenant
func (r *webhookRepository) GetEnabledByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*webhook.Webhook, error) {
	query := `
		SELECT id, tenant_id, name, url, secret, enabled, events,
		       created_at, updated_at, deleted_at
		FROM webhooks
		WHERE tenant_id = $1 AND enabled = true AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query webhooks: %w", err)
	}
	defer rows.Close()

	var webhooks []*models.Webhook
	for rows.Next() {
		var w models.Webhook
		var events pq.StringArray
		var deletedAt sql.NullTime

		err := rows.Scan(
			&w.ID,
			&w.TenantID,
			&w.Name,
			&w.URL,
			&w.Secret,
			&w.Enabled,
			&events,
			&w.CreatedAt,
			&w.UpdatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan webhook: %w", err)
		}

		w.Events = []string(events)
		if deletedAt.Valid {
			w.DeletedAt = &deletedAt.Time
		}

		webhooks = append(webhooks, &w)
	}

	return webhooks, rows.Err()
}

// GetByEventType retrieves all enabled webhooks subscribed to an event type
func (r *webhookRepository) GetByEventType(ctx context.Context, tenantID uuid.UUID, eventType string) ([]*webhook.Webhook, error) {
	query := `
		SELECT id, tenant_id, name, url, secret, enabled, events,
		       created_at, updated_at, deleted_at
		FROM webhooks
		WHERE tenant_id = $1 
		  AND enabled = true 
		  AND deleted_at IS NULL
		  AND $2 = ANY(events)
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID, eventType)
	if err != nil {
		return nil, fmt.Errorf("failed to query webhooks: %w", err)
	}
	defer rows.Close()

	var webhooks []*models.Webhook
	for rows.Next() {
		var w models.Webhook
		var events pq.StringArray
		var deletedAt sql.NullTime

		err := rows.Scan(
			&w.ID,
			&w.TenantID,
			&w.Name,
			&w.URL,
			&w.Secret,
			&w.Enabled,
			&events,
			&w.CreatedAt,
			&w.UpdatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan webhook: %w", err)
		}

		w.Events = []string(events)
		if deletedAt.Valid {
			w.DeletedAt = &deletedAt.Time
		}

		webhooks = append(webhooks, &w)
	}

	return webhooks, rows.Err()
}

// Update updates an existing webhook
func (r *webhookRepository) Update(ctx context.Context, w *models.Webhook) error {
	query := `
		UPDATE webhooks
		SET name = $2, url = $3, secret = $4, enabled = $5, events = $6,
		    updated_at = $7
		WHERE id = $1 AND deleted_at IS NULL
	`

	w.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		w.ID,
		w.Name,
		w.URL,
		w.Secret,
		w.Enabled,
		pq.Array(w.Events),
		w.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update webhook: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("webhook not found")
	}

	return nil
}

// Delete soft deletes a webhook
func (r *webhookRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE webhooks
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete webhook: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("webhook not found")
	}

	return nil
}

// webhookDeliveryRepository implements WebhookDeliveryRepository
type webhookDeliveryRepository struct {
	db *sql.DB
}

// NewWebhookDeliveryRepository creates a new webhook delivery repository
func NewWebhookDeliveryRepository(db *sql.DB) interfaces.WebhookDeliveryRepository {
	return &webhookDeliveryRepository{db: db}
}

// Create creates a new webhook delivery record
func (r *webhookDeliveryRepository) Create(ctx context.Context, delivery *models.WebhookDelivery) error {
	query := `
		INSERT INTO webhook_deliveries (
			id, webhook_id, event_id, event_type, payload, status,
			http_status_code, response_body, attempt_number, next_retry_at,
			delivered_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)
	`

	now := time.Now()
	delivery.CreatedAt = now
	delivery.UpdatedAt = now

	payloadJSON, err := json.Marshal(delivery.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query,
		delivery.ID,
		delivery.WebhookID,
		delivery.EventID,
		delivery.EventType,
		payloadJSON,
		delivery.Status,
		delivery.HTTPStatusCode,
		delivery.ResponseBody,
		delivery.AttemptNumber,
		delivery.NextRetryAt,
		delivery.DeliveredAt,
		delivery.CreatedAt,
		delivery.UpdatedAt,
	)

	return err
}

// GetByID retrieves a webhook delivery by ID
func (r *webhookDeliveryRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.WebhookDelivery, error) {
	query := `
		SELECT id, webhook_id, event_id, event_type, payload, status,
		       http_status_code, response_body, attempt_number, next_retry_at,
		       delivered_at, created_at, updated_at
		FROM webhook_deliveries
		WHERE id = $1
	`

	var delivery models.WebhookDelivery
	var payloadJSON []byte
	var eventID sql.NullString
	var httpStatusCode sql.NullInt64
	var responseBody sql.NullString
	var nextRetryAt sql.NullTime
	var deliveredAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&delivery.ID,
		&delivery.WebhookID,
		&eventID,
		&delivery.EventType,
		&payloadJSON,
		&delivery.Status,
		&httpStatusCode,
		&responseBody,
		&delivery.AttemptNumber,
		&nextRetryAt,
		&deliveredAt,
		&delivery.CreatedAt,
		&delivery.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("webhook delivery not found: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get webhook delivery: %w", err)
	}

	if eventID.Valid {
		parsedID, err := uuid.Parse(eventID.String)
		if err == nil {
			delivery.EventID = &parsedID
		}
	}

	if err := json.Unmarshal(payloadJSON, &delivery.Payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	if httpStatusCode.Valid {
		code := int(httpStatusCode.Int64)
		delivery.HTTPStatusCode = &code
	}

	if responseBody.Valid {
		delivery.ResponseBody = &responseBody.String
	}

	if nextRetryAt.Valid {
		delivery.NextRetryAt = &nextRetryAt.Time
	}

	if deliveredAt.Valid {
		delivery.DeliveredAt = &deliveredAt.Time
	}

	return &delivery, nil
}

// GetByWebhookID retrieves all deliveries for a webhook
func (r *webhookDeliveryRepository) GetByWebhookID(ctx context.Context, webhookID uuid.UUID, limit, offset int) ([]*webhook.WebhookDelivery, int, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM webhook_deliveries WHERE webhook_id = $1`
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, webhookID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count deliveries: %w", err)
	}

	// Get deliveries
	query := `
		SELECT id, webhook_id, event_id, event_type, payload, status,
		       http_status_code, response_body, attempt_number, next_retry_at,
		       delivered_at, created_at, updated_at
		FROM webhook_deliveries
		WHERE webhook_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, webhookID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query deliveries: %w", err)
	}
	defer rows.Close()

	var deliveries []*models.WebhookDelivery
	for rows.Next() {
		delivery, err := r.scanDelivery(rows)
		if err != nil {
			return nil, 0, err
		}
		deliveries = append(deliveries, delivery)
	}

	return deliveries, total, rows.Err()
}

// GetPendingRetries retrieves all deliveries that need to be retried
func (r *webhookDeliveryRepository) GetPendingRetries(ctx context.Context, before time.Time) ([]*webhook.WebhookDelivery, error) {
	query := `
		SELECT id, webhook_id, event_id, event_type, payload, status,
		       http_status_code, response_body, attempt_number, next_retry_at,
		       delivered_at, created_at, updated_at
		FROM webhook_deliveries
		WHERE status IN ('failed', 'retrying')
		  AND next_retry_at IS NOT NULL
		  AND next_retry_at <= $1
		ORDER BY next_retry_at ASC
		LIMIT 100
	`

	rows, err := r.db.QueryContext(ctx, query, before)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending retries: %w", err)
	}
	defer rows.Close()

	var deliveries []*models.WebhookDelivery
	for rows.Next() {
		delivery, err := r.scanDelivery(rows)
		if err != nil {
			return nil, err
		}
		deliveries = append(deliveries, delivery)
	}

	return deliveries, rows.Err()
}

// Update updates a webhook delivery record
func (r *webhookDeliveryRepository) Update(ctx context.Context, delivery *models.WebhookDelivery) error {
	query := `
		UPDATE webhook_deliveries
		SET status = $2, http_status_code = $3, response_body = $4,
		    attempt_number = $5, next_retry_at = $6, delivered_at = $7,
		    updated_at = $8
		WHERE id = $1
	`

	delivery.UpdatedAt = time.Now()

	payloadJSON, err := json.Marshal(delivery.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Re-marshal payload to ensure it's stored (though we don't update it)
	_ = payloadJSON

	_, err = r.db.ExecContext(ctx, query,
		delivery.ID,
		delivery.Status,
		delivery.HTTPStatusCode,
		delivery.ResponseBody,
		delivery.AttemptNumber,
		delivery.NextRetryAt,
		delivery.DeliveredAt,
		delivery.UpdatedAt,
	)

	return err
}

// scanDelivery scans a delivery from a row
func (r *webhookDeliveryRepository) scanDelivery(rows *sql.Rows) (*webhook.WebhookDelivery, error) {
	var delivery models.WebhookDelivery
	var payloadJSON []byte
	var eventID sql.NullString
	var httpStatusCode sql.NullInt64
	var responseBody sql.NullString
	var nextRetryAt sql.NullTime
	var deliveredAt sql.NullTime

	err := rows.Scan(
		&delivery.ID,
		&delivery.WebhookID,
		&eventID,
		&delivery.EventType,
		&payloadJSON,
		&delivery.Status,
		&httpStatusCode,
		&responseBody,
		&delivery.AttemptNumber,
		&nextRetryAt,
		&deliveredAt,
		&delivery.CreatedAt,
		&delivery.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan delivery: %w", err)
	}

	if eventID.Valid {
		parsedID, err := uuid.Parse(eventID.String)
		if err == nil {
			delivery.EventID = &parsedID
		}
	}

	if err := json.Unmarshal(payloadJSON, &delivery.Payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	if httpStatusCode.Valid {
		code := int(httpStatusCode.Int64)
		delivery.HTTPStatusCode = &code
	}

	if responseBody.Valid {
		delivery.ResponseBody = &responseBody.String
	}

	if nextRetryAt.Valid {
		delivery.NextRetryAt = &nextRetryAt.Time
	}

	if deliveredAt.Valid {
		delivery.DeliveredAt = &deliveredAt.Time
	}

	return &delivery, nil
}

