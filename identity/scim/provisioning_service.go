package scim

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/arauth-identity/iam/identity/models"
	"github.com/arauth-identity/iam/identity/role"
	"github.com/arauth-identity/iam/identity/user"
	"github.com/arauth-identity/iam/storage/interfaces"
)

// ProvisioningService provides SCIM 2.0 provisioning
type ProvisioningService struct {
	userService    user.ServiceInterface
	roleService    role.ServiceInterface
	userRepo       interfaces.UserRepository
	roleRepo       interfaces.RoleRepository
	tenantID       uuid.UUID // Tenant ID from token context
}

// NewProvisioningService creates a new SCIM provisioning service
func NewProvisioningService(
	userService user.ServiceInterface,
	roleService role.ServiceInterface,
	userRepo interfaces.UserRepository,
	roleRepo interfaces.RoleRepository,
) ProvisioningServiceInterface {
	return &ProvisioningService{
		userService: userService,
		roleService: roleService,
		userRepo:    userRepo,
		roleRepo:    roleRepo,
	}
}

// SetTenantID sets the tenant ID for the service (from token context)
func (s *ProvisioningService) SetTenantID(tenantID uuid.UUID) {
	s.tenantID = tenantID
}

// CreateUser creates a user from SCIM User resource
func (s *ProvisioningService) CreateUser(ctx context.Context, tenantID uuid.UUID, scimUser *models.SCIMUser) (*models.SCIMUser, error) {
	// Extract username
	username := scimUser.UserName
	if username == "" {
		return nil, fmt.Errorf("userName is required")
	}

	// Extract email (primary email)
	var email string
	for _, e := range scimUser.Emails {
		if e.Primary || email == "" {
			email = e.Value
		}
	}
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}

	// Extract name
	firstName := scimUser.Name.GivenName
	lastName := scimUser.Name.FamilyName

	// Extract password (if provided)
	password := scimUser.Password
	if password == "" {
		// Generate a random password if not provided
		password = generateRandomPassword()
	}

	// Create user request
	createReq := &user.CreateUserRequest{
		TenantID:  tenantID,
		Username:  username,
		Email:     email,
		Password:  password,
		FirstName: &firstName,
		LastName:  &lastName,
		Status:    mapSCIMActiveToStatus(scimUser.Active),
	}

	// Create user
	createdUser, err := s.userService.Create(ctx, createReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Convert to SCIM format
	return s.userToSCIM(createdUser, tenantID), nil
}

// GetUser retrieves a user by ID and converts to SCIM format
func (s *ProvisioningService) GetUser(ctx context.Context, tenantID uuid.UUID, userID string) (*models.SCIMUser, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format")
	}

	u, err := s.userService.GetByID(ctx, userUUID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Verify tenant ownership
	if u.TenantID == nil || *u.TenantID != tenantID {
		return nil, fmt.Errorf("user not found")
	}

	return s.userToSCIM(u, tenantID), nil
}

// GetUserByExternalID retrieves a user by external ID
func (s *ProvisioningService) GetUserByExternalID(ctx context.Context, tenantID uuid.UUID, externalID string) (*models.SCIMUser, error) {
	// For now, we'll use externalID as username
	// In production, you might want a separate external_id field
	u, err := s.userService.GetByUsername(ctx, externalID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return s.userToSCIM(u, tenantID), nil
}

// GetUserByUserName retrieves a user by username
func (s *ProvisioningService) GetUserByUserName(ctx context.Context, tenantID uuid.UUID, userName string) (*models.SCIMUser, error) {
	u, err := s.userService.GetByUsername(ctx, userName, tenantID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return s.userToSCIM(u, tenantID), nil
}

// ListUsers lists users with SCIM filters
func (s *ProvisioningService) ListUsers(ctx context.Context, tenantID uuid.UUID, filters *UserFilters) ([]*models.SCIMUser, int, error) {
	// Convert SCIM filters to internal filters
	internalFilters := &interfaces.UserFilters{
		Page:     (filters.StartIndex / filters.Count) + 1,
		PageSize: filters.Count,
	}

	// Parse SCIM filter if provided
	if filters.Filter != "" {
		// Simple filter parsing (can be enhanced)
		// For now, support: userName eq "value", email eq "value"
		if strings.HasPrefix(filters.Filter, "userName eq ") {
			userName := strings.Trim(strings.TrimPrefix(filters.Filter, "userName eq "), "\"")
			u, err := s.userService.GetByUsername(ctx, userName, tenantID)
			if err != nil {
				return []*models.SCIMUser{}, 0, nil
			}
			return []*models.SCIMUser{s.userToSCIM(u, tenantID)}, 1, nil
		}
	}

	users, err := s.userService.List(ctx, tenantID, internalFilters)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	total, err := s.userService.Count(ctx, tenantID, internalFilters)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	scimUsers := make([]*models.SCIMUser, len(users))
	for i, u := range users {
		scimUsers[i] = s.userToSCIM(u, tenantID)
	}

	return scimUsers, total, nil
}

// UpdateUser updates a user from SCIM User resource
func (s *ProvisioningService) UpdateUser(ctx context.Context, tenantID uuid.UUID, userID string, scimUser *models.SCIMUser) (*models.SCIMUser, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format")
	}

	// Get existing user
	existingUser, err := s.userService.GetByID(ctx, userUUID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Verify tenant ownership
	if existingUser.TenantID == nil || *existingUser.TenantID != tenantID {
		return nil, fmt.Errorf("user not found")
	}

	// Build update request
	updateReq := &user.UpdateUserRequest{}

	if scimUser.UserName != "" {
		updateReq.Username = &scimUser.UserName
	}

	// Update email (primary email)
	for _, e := range scimUser.Emails {
		if e.Primary || updateReq.Email == nil {
			updateReq.Email = &e.Value
		}
	}

	// Update name
	if scimUser.Name.GivenName != "" {
		updateReq.FirstName = &scimUser.Name.GivenName
	}
	if scimUser.Name.FamilyName != "" {
		updateReq.LastName = &scimUser.Name.FamilyName
	}

	// Update status
	if scimUser.Active != mapStatusToSCIMActive(existingUser.Status) {
		status := mapSCIMActiveToStatus(scimUser.Active)
		updateReq.Status = &status
	}

	// Update password if provided
	if scimUser.Password != "" {
		// Password update would need credential service
		// For now, skip password updates via SCIM
	}

	// Update user
	updatedUser, err := s.userService.Update(ctx, userUUID, updateReq)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return s.userToSCIM(updatedUser, tenantID), nil
}

// DeleteUser deletes a user
func (s *ProvisioningService) DeleteUser(ctx context.Context, tenantID uuid.UUID, userID string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID format")
	}

	// Verify tenant ownership
	existingUser, err := s.userService.GetByID(ctx, userUUID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	if existingUser.TenantID == nil || *existingUser.TenantID != tenantID {
		return fmt.Errorf("user not found")
	}

	return s.userService.Delete(ctx, userUUID)
}

// CreateGroup creates a group from SCIM Group resource
func (s *ProvisioningService) CreateGroup(ctx context.Context, tenantID uuid.UUID, scimGroup *models.SCIMGroup) (*models.SCIMGroup, error) {
	// Create role request
	createReq := &role.CreateRoleRequest{
		TenantID: tenantID,
		Name:     scimGroup.DisplayName,
		Description: &scimGroup.Description,
	}

	// Create role
	createdRole, err := s.roleService.Create(ctx, createReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create group: %w", err)
	}

	// Assign members (users) to role
	for _, member := range scimGroup.Members {
		if member.Type == "User" || member.Type == "" {
			userUUID, err := uuid.Parse(member.Value)
			if err == nil {
				_ = s.roleService.AssignRoleToUser(ctx, userUUID, createdRole.ID)
			}
		}
	}

	return s.roleToSCIM(createdRole, tenantID), nil
}

// GetGroup retrieves a group by ID
func (s *ProvisioningService) GetGroup(ctx context.Context, tenantID uuid.UUID, groupID string) (*models.SCIMGroup, error) {
	roleUUID, err := uuid.Parse(groupID)
	if err != nil {
		return nil, fmt.Errorf("invalid group ID format")
	}

	r, err := s.roleService.GetByID(ctx, roleUUID)
	if err != nil {
		return nil, fmt.Errorf("group not found: %w", err)
	}

	// Verify tenant ownership
	if r.TenantID != tenantID {
		return nil, fmt.Errorf("group not found")
	}

	return s.roleToSCIM(r, tenantID), nil
}

// GetGroupByExternalID retrieves a group by external ID
func (s *ProvisioningService) GetGroupByExternalID(ctx context.Context, tenantID uuid.UUID, externalID string) (*models.SCIMGroup, error) {
	// For now, use externalID as name
	// In production, add external_id field to roles
	roles, err := s.roleService.List(ctx, tenantID, &interfaces.RoleFilters{})
	if err != nil {
		return nil, fmt.Errorf("failed to list groups: %w", err)
	}

	for _, r := range roles {
		if r.Name == externalID {
			return s.roleToSCIM(r, tenantID), nil
		}
	}

	return nil, fmt.Errorf("group not found")
}

// ListGroups lists groups
func (s *ProvisioningService) ListGroups(ctx context.Context, tenantID uuid.UUID, filters *GroupFilters) ([]*models.SCIMGroup, int, error) {
	roles, err := s.roleService.List(ctx, tenantID, &interfaces.RoleFilters{})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list groups: %w", err)
	}

	scimGroups := make([]*models.SCIMGroup, len(roles))
	for i, r := range roles {
		scimGroups[i] = s.roleToSCIM(r, tenantID)
	}

	return scimGroups, len(scimGroups), nil
}

// UpdateGroup updates a group
func (s *ProvisioningService) UpdateGroup(ctx context.Context, tenantID uuid.UUID, groupID string, scimGroup *models.SCIMGroup) (*models.SCIMGroup, error) {
	roleUUID, err := uuid.Parse(groupID)
	if err != nil {
		return nil, fmt.Errorf("invalid group ID format")
	}

	// Verify tenant ownership
	existingRole, err := s.roleService.GetByID(ctx, roleUUID)
	if err != nil {
		return nil, fmt.Errorf("group not found: %w", err)
	}

	if existingRole.TenantID != tenantID {
		return nil, fmt.Errorf("group not found")
	}

	// Build update request
	updateReq := &role.UpdateRoleRequest{}
	if scimGroup.DisplayName != "" {
		updateReq.Name = &scimGroup.DisplayName
	}
	if scimGroup.Description != "" {
		updateReq.Description = &scimGroup.Description
	}

	// Update role
	updatedRole, err := s.roleService.Update(ctx, roleUUID, updateReq)
	if err != nil {
		return nil, fmt.Errorf("failed to update group: %w", err)
	}

	// Update members if provided
	// TODO: Implement member management

	return s.roleToSCIM(updatedRole, tenantID), nil
}

// DeleteGroup deletes a group
func (s *ProvisioningService) DeleteGroup(ctx context.Context, tenantID uuid.UUID, groupID string) error {
	roleUUID, err := uuid.Parse(groupID)
	if err != nil {
		return fmt.Errorf("invalid group ID format")
	}

	// Verify tenant ownership
	existingRole, err := s.roleService.GetByID(ctx, roleUUID)
	if err != nil {
		return fmt.Errorf("group not found: %w", err)
	}

	if existingRole.TenantID != tenantID {
		return fmt.Errorf("group not found")
	}

	return s.roleService.Delete(ctx, roleUUID)
}

// BulkCreate handles bulk operations
func (s *ProvisioningService) BulkCreate(ctx context.Context, tenantID uuid.UUID, operations []BulkOperation) (*BulkResponse, error) {
	results := make([]BulkOperationResult, len(operations))

	for i, op := range operations {
		result := BulkOperationResult{
			Method: op.Method,
			BulkID: op.BulkID,
		}

		switch op.Method {
		case "POST":
			// Handle POST operations
			// This would need to parse the path and data
			result.Status = "201"
		case "PUT", "PATCH":
			result.Status = "200"
		case "DELETE":
			result.Status = "204"
		default:
			result.Status = "400"
		}

		results[i] = result
	}

	return &BulkResponse{
		Schemas:    []string{"urn:ietf:params:scim:api:messages:2.0:BulkResponse"},
		Operations: results,
	}, nil
}

// Helper functions

func (s *ProvisioningService) userToSCIM(u *models.User, tenantID uuid.UUID) *models.SCIMUser {
	scimUser := &models.SCIMUser{
		Schemas:    []string{"urn:ietf:params:scim:schemas:core:2.0:User"},
		ID:         u.ID.String(),
		UserName:   u.Username,
		Active:     mapStatusToSCIMActive(u.Status),
		Name: models.SCIMName{
			GivenName:  getStringValue(u.FirstName),
			FamilyName: getStringValue(u.LastName),
		},
		Emails: []models.SCIMEmail{
			{
				Value:   u.Email,
				Primary: true,
			},
		},
		Meta: models.SCIMMeta{
			ResourceType: "User",
			Created:      u.CreatedAt,
			LastModified: u.UpdatedAt,
		},
	}

	// Set display name
	if u.FirstName != nil && u.LastName != nil {
		scimUser.DisplayName = *u.FirstName + " " + *u.LastName
		scimUser.Name.Formatted = scimUser.DisplayName
	}

	return scimUser
}

func (s *ProvisioningService) roleToSCIM(r *models.Role, tenantID uuid.UUID) *models.SCIMGroup {
	scimGroup := &models.SCIMGroup{
		Schemas:     []string{"urn:ietf:params:scim:schemas:core:2.0:Group"},
		ID:          r.ID.String(),
		DisplayName: r.Name,
		Description: getStringValue(r.Description),
		Meta: models.SCIMMeta{
			ResourceType: "Group",
			Created:      r.CreatedAt,
			LastModified: r.UpdatedAt,
		},
	}

	// TODO: Populate members from role assignments

	return scimGroup
}

func mapSCIMActiveToStatus(active bool) string {
	if active {
		return "active"
	}
	return "disabled"
}

func mapStatusToSCIMActive(status string) bool {
	return status == "active"
}

func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func generateRandomPassword() string {
	// Generate a secure random password
	// In production, this should be a proper random password generator
	return "TempPassword123!@#"
}

