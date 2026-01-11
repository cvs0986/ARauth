package scim

import (
	"context"

	"github.com/arauth-identity/iam/identity/models"
	"github.com/google/uuid"
)

// ProvisioningServiceInterface defines the interface for SCIM provisioning
type ProvisioningServiceInterface interface {
	// User provisioning
	CreateUser(ctx context.Context, tenantID uuid.UUID, scimUser *models.SCIMUser) (*models.SCIMUser, error)
	GetUser(ctx context.Context, tenantID uuid.UUID, userID string) (*models.SCIMUser, error)
	GetUserByExternalID(ctx context.Context, tenantID uuid.UUID, externalID string) (*models.SCIMUser, error)
	GetUserByUserName(ctx context.Context, tenantID uuid.UUID, userName string) (*models.SCIMUser, error)
	ListUsers(ctx context.Context, tenantID uuid.UUID, filters *UserFilters) ([]*models.SCIMUser, int, error)
	UpdateUser(ctx context.Context, tenantID uuid.UUID, userID string, scimUser *models.SCIMUser) (*models.SCIMUser, error)
	DeleteUser(ctx context.Context, tenantID uuid.UUID, userID string) error

	// Group provisioning
	CreateGroup(ctx context.Context, tenantID uuid.UUID, scimGroup *models.SCIMGroup) (*models.SCIMGroup, error)
	GetGroup(ctx context.Context, tenantID uuid.UUID, groupID string) (*models.SCIMGroup, error)
	GetGroupByExternalID(ctx context.Context, tenantID uuid.UUID, externalID string) (*models.SCIMGroup, error)
	ListGroups(ctx context.Context, tenantID uuid.UUID, filters *GroupFilters) ([]*models.SCIMGroup, int, error)
	UpdateGroup(ctx context.Context, tenantID uuid.UUID, groupID string, scimGroup *models.SCIMGroup) (*models.SCIMGroup, error)
	DeleteGroup(ctx context.Context, tenantID uuid.UUID, groupID string) error

	// Bulk operations
	BulkCreate(ctx context.Context, tenantID uuid.UUID, operations []BulkOperation) (*BulkResponse, error)
}

// UserFilters defines filters for listing users
type UserFilters struct {
	Filter     string // SCIM filter expression
	StartIndex int
	Count      int
}

// GroupFilters defines filters for listing groups
type GroupFilters struct {
	Filter     string // SCIM filter expression
	StartIndex int
	Count      int
}

// BulkOperation represents a bulk operation
type BulkOperation struct {
	Method string      `json:"method"` // POST, PUT, PATCH, DELETE
	Path   string      `json:"path"`
	BulkID string      `json:"bulkId,omitempty"`
	Data   interface{} `json:"data,omitempty"`
}

// BulkResponse represents a bulk operation response
type BulkResponse struct {
	Schemas      []string           `json:"schemas"`
	Operations   []BulkOperationResult `json:"Operations"`
}

// BulkOperationResult represents the result of a bulk operation
type BulkOperationResult struct {
	Location string      `json:"location,omitempty"`
	Method   string      `json:"method"`
	BulkID   string      `json:"bulkId,omitempty"`
	Status   string      `json:"status"`
	Response interface{} `json:"response,omitempty"`
}


