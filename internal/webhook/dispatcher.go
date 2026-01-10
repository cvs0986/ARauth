package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/arauth-identity/iam/identity/models"
	"github.com/arauth-identity/iam/storage/interfaces"
	"go.uber.org/zap"
)

// Dispatcher handles webhook delivery with retry logic
type Dispatcher struct {
	deliveryRepo interfaces.WebhookDeliveryRepository
	httpClient   *http.Client
	logger       *zap.Logger
	maxRetries   int
	retryBackoff time.Duration
}

// NewDispatcher creates a new webhook dispatcher
func NewDispatcher(
	deliveryRepo interfaces.WebhookDeliveryRepository,
	logger *zap.Logger,
) *Dispatcher {
	return &Dispatcher{
		deliveryRepo: deliveryRepo,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger:       logger,
		maxRetries:   5,
		retryBackoff: 1 * time.Minute,
	}
}

// Deliver sends a webhook payload to the specified URL
func (d *Dispatcher) Deliver(ctx context.Context, w *models.Webhook, eventType string, payload map[string]interface{}, eventID *uuid.UUID) error {
	// Create webhook payload
	webhookPayload := models.WebhookPayload{
		ID:        uuid.New().String(),
		EventType: eventType,
		Timestamp: time.Now(),
		Data:      payload,
	}

	// Marshal payload
	payloadJSON, err := json.Marshal(webhookPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Sign payload with HMAC-SHA256
	signature := d.signPayload(payloadJSON, w.Secret)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", w.URL, bytes.NewReader(payloadJSON))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Signature", signature)
	req.Header.Set("X-Webhook-Event", eventType)
	req.Header.Set("X-Webhook-ID", webhookPayload.ID)

	// Send request
	resp, err := d.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	responseBody, _ := io.ReadAll(resp.Body)

	// Create delivery record
	delivery := &models.WebhookDelivery{
		ID:            uuid.New(),
		WebhookID:     w.ID,
		EventID:       eventID,
		EventType:     eventType,
		Payload:       payload,
		Status:        models.DeliveryStatusPending,
		HTTPStatusCode: &resp.StatusCode,
		AttemptNumber:  1,
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// Success
		delivery.Status = models.DeliveryStatusSuccess
		now := time.Now()
		delivery.DeliveredAt = &now
	} else {
		// Failed
		delivery.Status = models.DeliveryStatusFailed
		responseBodyStr := string(responseBody)
		delivery.ResponseBody = &responseBodyStr

		// Schedule retry if not exceeded max retries
		if delivery.AttemptNumber < d.maxRetries {
			delivery.Status = models.DeliveryStatusRetrying
			nextRetry := time.Now().Add(d.calculateBackoff(delivery.AttemptNumber))
			delivery.NextRetryAt = &nextRetry
		}
	}

	// Save delivery record
	if err := d.deliveryRepo.Create(ctx, delivery); err != nil {
		d.logger.Error("Failed to create delivery record", zap.Error(err))
	}

	return nil
}

// RetryFailedDeliveries retries failed webhook deliveries
func (d *Dispatcher) RetryFailedDeliveries(ctx context.Context, webhookRepo interfaces.WebhookRepository) error {
	// Get pending retries
	before := time.Now()
	deliveries, err := d.deliveryRepo.GetPendingRetries(ctx, before)
	if err != nil {
		return fmt.Errorf("failed to get pending retries: %w", err)
	}

	for _, delivery := range deliveries {
		// Get webhook
		w, err := webhookRepo.GetByID(ctx, delivery.WebhookID)
		if err != nil {
			d.logger.Error("Failed to get webhook for retry", zap.Error(err))
			continue
		}

		// Skip if webhook is disabled
		if !w.Enabled {
			delivery.Status = models.DeliveryStatusFailed
			delivery.NextRetryAt = nil
			if err := d.deliveryRepo.Update(ctx, delivery); err != nil {
				d.logger.Error("Failed to update delivery", zap.Error(err))
			}
			continue
		}

		// Retry delivery
		delivery.AttemptNumber++
		delivery.Status = models.DeliveryStatusRetrying

		// Marshal payload
		webhookPayload := models.WebhookPayload{
			ID:        uuid.New().String(),
			EventType: delivery.EventType,
			Timestamp: time.Now(),
			Data:      delivery.Payload,
		}

		payloadJSON, err := json.Marshal(webhookPayload)
		if err != nil {
			delivery.Status = models.DeliveryStatusFailed
			delivery.NextRetryAt = nil
			if err := d.deliveryRepo.Update(ctx, delivery); err != nil {
				d.logger.Error("Failed to update delivery", zap.Error(err))
			}
			continue
		}

		// Sign payload
		signature := d.signPayload(payloadJSON, w.Secret)

		// Create HTTP request
		req, err := http.NewRequestWithContext(ctx, "POST", w.URL, bytes.NewReader(payloadJSON))
		if err != nil {
			delivery.Status = models.DeliveryStatusFailed
			delivery.NextRetryAt = nil
			if err := d.deliveryRepo.Update(ctx, delivery); err != nil {
				d.logger.Error("Failed to update delivery", zap.Error(err))
			}
			continue
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Webhook-Signature", signature)
		req.Header.Set("X-Webhook-Event", delivery.EventType)
		req.Header.Set("X-Webhook-ID", webhookPayload.ID)

		// Send request
		resp, err := d.httpClient.Do(req)
		if err != nil {
			// Schedule next retry
			if delivery.AttemptNumber < d.maxRetries {
				nextRetry := time.Now().Add(d.calculateBackoff(delivery.AttemptNumber))
				delivery.NextRetryAt = &nextRetry
			} else {
				delivery.Status = models.DeliveryStatusFailed
				delivery.NextRetryAt = nil
			}
			if err := d.deliveryRepo.Update(ctx, delivery); err != nil {
				d.logger.Error("Failed to update delivery", zap.Error(err))
			}
			continue
		}
		defer resp.Body.Close()

		// Read response
		responseBody, _ := io.ReadAll(resp.Body)
		responseBodyStr := string(responseBody)
		delivery.HTTPStatusCode = &resp.StatusCode
		delivery.ResponseBody = &responseBodyStr

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			// Success
			delivery.Status = models.DeliveryStatusSuccess
			now := time.Now()
			delivery.DeliveredAt = &now
			delivery.NextRetryAt = nil
		} else {
			// Failed - schedule next retry
			if delivery.AttemptNumber < d.maxRetries {
				delivery.Status = models.DeliveryStatusRetrying
				nextRetry := time.Now().Add(d.calculateBackoff(delivery.AttemptNumber))
				delivery.NextRetryAt = &nextRetry
			} else {
				delivery.Status = models.DeliveryStatusFailed
				delivery.NextRetryAt = nil
			}
		}

		if err := d.deliveryRepo.Update(ctx, delivery); err != nil {
			d.logger.Error("Failed to update delivery", zap.Error(err))
		}
	}

	return nil
}

// signPayload signs a payload with HMAC-SHA256
func (d *Dispatcher) signPayload(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

// calculateBackoff calculates exponential backoff delay
func (d *Dispatcher) calculateBackoff(attemptNumber int) time.Duration {
	// Exponential backoff: 1min, 2min, 4min, 8min, 16min
	backoff := d.retryBackoff * time.Duration(1<<uint(attemptNumber-1))
	if backoff > 16*time.Minute {
		backoff = 16 * time.Minute
	}
	return backoff
}

