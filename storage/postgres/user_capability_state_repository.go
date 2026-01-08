package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/arauth-identity/iam/identity/models"
	"github.com/arauth-identity/iam/storage/interfaces"
)

// userCapabilityStateRepository implements UserCapabilityStateRepository for PostgreSQL
type userCapabilityStateRepository struct {
	db *sql.DB
}

// NewUserCapabilityStateRepository creates a new PostgreSQL user capability state repository
func NewUserCapabilityStateRepository(db *sql.DB) interfaces.UserCapabilityStateRepository {
	return &userCapabilityStateRepository{db: db}
}

// GetByUserIDAndKey retrieves a user capability state by user ID and key
func (r *userCapabilityStateRepository) GetByUserIDAndKey(ctx context.Context, userID uuid.UUID, key string) (*models.UserCapabilityState, error) {
	query := `
		SELECT user_id, capability_key, enrolled, state_data, enrolled_at, last_used_at
		FROM user_capability_state
		WHERE user_id = $1 AND capability_key = $2
	`

	state := &models.UserCapabilityState{}
	var stateData sql.NullString
	var enrolledAt sql.NullTime
	var lastUsedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, userID, key).Scan(
		&state.UserID,
		&state.CapabilityKey,
		&state.Enrolled,
		&stateData,
		&enrolledAt,
		&lastUsedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user capability state not found: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user capability state: %w", err)
	}

	if stateData.Valid {
		state.StateData = json.RawMessage(stateData.String)
	}
	if enrolledAt.Valid {
		state.EnrolledAt = &enrolledAt.Time
	}
	if lastUsedAt.Valid {
		state.LastUsedAt = &lastUsedAt.Time
	}

	return state, nil
}

// GetByUserID retrieves all capability states for a user
func (r *userCapabilityStateRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.UserCapabilityState, error) {
	query := `
		SELECT user_id, capability_key, enrolled, state_data, enrolled_at, last_used_at
		FROM user_capability_state
		WHERE user_id = $1
		ORDER BY capability_key
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user capability states: %w", err)
	}
	defer rows.Close()

	var states []*models.UserCapabilityState
	for rows.Next() {
		state := &models.UserCapabilityState{}
		var stateData sql.NullString
		var enrolledAt sql.NullTime
		var lastUsedAt sql.NullTime

		err := rows.Scan(
			&state.UserID,
			&state.CapabilityKey,
			&state.Enrolled,
			&stateData,
			&enrolledAt,
			&lastUsedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user capability state: %w", err)
		}

		if stateData.Valid {
			state.StateData = json.RawMessage(stateData.String)
		}
		if enrolledAt.Valid {
			state.EnrolledAt = &enrolledAt.Time
		}
		if lastUsedAt.Valid {
			state.LastUsedAt = &lastUsedAt.Time
		}

		states = append(states, state)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user capability states: %w", err)
	}

	return states, nil
}

// GetEnrolledByUserID retrieves all enrolled capabilities for a user
func (r *userCapabilityStateRepository) GetEnrolledByUserID(ctx context.Context, userID uuid.UUID) ([]*models.UserCapabilityState, error) {
	query := `
		SELECT user_id, capability_key, enrolled, state_data, enrolled_at, last_used_at
		FROM user_capability_state
		WHERE user_id = $1 AND enrolled = true
		ORDER BY capability_key
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get enrolled user capability states: %w", err)
	}
	defer rows.Close()

	var states []*models.UserCapabilityState
	for rows.Next() {
		state := &models.UserCapabilityState{}
		var stateData sql.NullString
		var enrolledAt sql.NullTime
		var lastUsedAt sql.NullTime

		err := rows.Scan(
			&state.UserID,
			&state.CapabilityKey,
			&state.Enrolled,
			&stateData,
			&enrolledAt,
			&lastUsedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user capability state: %w", err)
		}

		if stateData.Valid {
			state.StateData = json.RawMessage(stateData.String)
		}
		if enrolledAt.Valid {
			state.EnrolledAt = &enrolledAt.Time
		}
		if lastUsedAt.Valid {
			state.LastUsedAt = &lastUsedAt.Time
		}

		states = append(states, state)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating enrolled user capability states: %w", err)
	}

	return states, nil
}

// Create creates a new user capability state
func (r *userCapabilityStateRepository) Create(ctx context.Context, state *models.UserCapabilityState) error {
	query := `
		INSERT INTO user_capability_state (user_id, capability_key, enrolled, state_data, enrolled_at, last_used_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	stateDataJSON := "{}"
	if len(state.StateData) > 0 {
		stateDataJSON = string(state.StateData)
	}

	now := time.Now()
	if state.Enrolled && state.EnrolledAt == nil {
		state.EnrolledAt = &now
	}

	_, err := r.db.ExecContext(ctx, query,
		state.UserID,
		state.CapabilityKey,
		state.Enrolled,
		stateDataJSON,
		state.EnrolledAt,
		state.LastUsedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create user capability state: %w", err)
	}

	return nil
}

// Update updates an existing user capability state
func (r *userCapabilityStateRepository) Update(ctx context.Context, state *models.UserCapabilityState) error {
	query := `
		UPDATE user_capability_state
		SET enrolled = $3, state_data = $4, enrolled_at = $5, last_used_at = $6
		WHERE user_id = $1 AND capability_key = $2
	`

	stateDataJSON := "{}"
	if len(state.StateData) > 0 {
		stateDataJSON = string(state.StateData)
	}

	now := time.Now()
	if state.Enrolled && state.EnrolledAt == nil {
		state.EnrolledAt = &now
	}

	_, err := r.db.ExecContext(ctx, query,
		state.UserID,
		state.CapabilityKey,
		state.Enrolled,
		stateDataJSON,
		state.EnrolledAt,
		state.LastUsedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update user capability state: %w", err)
	}

	return nil
}

// Delete deletes a user capability state
func (r *userCapabilityStateRepository) Delete(ctx context.Context, userID uuid.UUID, key string) error {
	query := `DELETE FROM user_capability_state WHERE user_id = $1 AND capability_key = $2`

	_, err := r.db.ExecContext(ctx, query, userID, key)
	if err != nil {
		return fmt.Errorf("failed to delete user capability state: %w", err)
	}

	return nil
}

// DeleteByUserID deletes all capability states for a user
func (r *userCapabilityStateRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM user_capability_state WHERE user_id = $1`

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user capability states: %w", err)
	}

	return nil
}

