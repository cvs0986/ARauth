package federation

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/arauth-identity/iam/auth/claims"
	"github.com/arauth-identity/iam/auth/token"
	"github.com/arauth-identity/iam/identity/federation"
	"github.com/arauth-identity/iam/identity/models"
	"github.com/arauth-identity/iam/storage/interfaces"
	oidcclient "github.com/arauth-identity/iam/auth/federation/oidc"
	samlclient "github.com/arauth-identity/iam/auth/federation/saml"
)

// Service provides federation functionality
type Service struct {
	idpRepo      interfaces.IdentityProviderRepository
	fedIdRepo    interfaces.FederatedIdentityRepository
	userRepo     interfaces.UserRepository
	credentialRepo interfaces.CredentialRepository
	claimsBuilder *claims.Builder
	tokenService  token.ServiceInterface
	stateStore   map[string]*State // In-memory state store (should be Redis in production)
}

// State represents OAuth state for federation
type State struct {
	ProviderID  uuid.UUID
	TenantID    uuid.UUID
	RedirectURI string
	CreatedAt   time.Time
}

// NewService creates a new federation service
func NewService(
	idpRepo interfaces.IdentityProviderRepository,
	fedIdRepo interfaces.FederatedIdentityRepository,
	userRepo interfaces.UserRepository,
	credentialRepo interfaces.CredentialRepository,
	claimsBuilder *claims.Builder,
	tokenService token.ServiceInterface,
) ServiceInterface {
	return &Service{
		idpRepo:        idpRepo,
		fedIdRepo:      fedIdRepo,
		userRepo:       userRepo,
		credentialRepo: credentialRepo,
		claimsBuilder:  claimsBuilder,
		tokenService:   tokenService,
		stateStore:     make(map[string]*State),
	}
}

// CreateIdentityProvider creates a new identity provider
func (s *Service) CreateIdentityProvider(ctx context.Context, tenantID uuid.UUID, req *CreateIdPRequest) (*federation.IdentityProvider, error) {
	// Validate provider type
	if req.Type != federation.IdentityProviderTypeOIDC && req.Type != federation.IdentityProviderTypeSAML {
		return nil, fmt.Errorf("invalid provider type: %s", req.Type)
	}

	// Validate configuration based on type
	if err := s.validateConfiguration(req.Type, req.Configuration); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	provider := &federation.IdentityProvider{
		ID:              uuid.New(),
		TenantID:        tenantID,
		Name:            req.Name,
		Type:            req.Type,
		Enabled:         req.Enabled,
		Configuration:   req.Configuration,
		AttributeMapping: req.AttributeMapping,
	}

	if err := s.idpRepo.Create(ctx, provider); err != nil {
		return nil, fmt.Errorf("failed to create identity provider: %w", err)
	}

	return provider, nil
}

// validateConfiguration validates provider configuration
func (s *Service) validateConfiguration(providerType federation.IdentityProviderType, config map[string]interface{}) error {
	if providerType == federation.IdentityProviderTypeOIDC {
		// Validate OIDC configuration
		if _, ok := config["client_id"]; !ok {
			return fmt.Errorf("client_id is required for OIDC provider")
		}
		if _, ok := config["client_secret"]; !ok {
			return fmt.Errorf("client_secret is required for OIDC provider")
		}
		if _, ok := config["issuer_url"]; !ok {
			return fmt.Errorf("issuer_url is required for OIDC provider")
		}
	} else if providerType == federation.IdentityProviderTypeSAML {
		// Validate SAML configuration
		if _, ok := config["entity_id"]; !ok {
			return fmt.Errorf("entity_id is required for SAML provider")
		}
		if _, ok := config["sso_url"]; !ok {
			return fmt.Errorf("sso_url is required for SAML provider")
		}
		if _, ok := config["x509_certificate"]; !ok {
			return fmt.Errorf("x509_certificate is required for SAML provider")
		}
	}

	return nil
}

// GetIdentityProvider retrieves an identity provider by ID
func (s *Service) GetIdentityProvider(ctx context.Context, id uuid.UUID) (*federation.IdentityProvider, error) {
	return s.idpRepo.GetByID(ctx, id)
}

// GetIdentityProvidersByTenant retrieves all identity providers for a tenant
func (s *Service) GetIdentityProvidersByTenant(ctx context.Context, tenantID uuid.UUID) ([]*federation.IdentityProvider, error) {
	return s.idpRepo.GetByTenantID(ctx, tenantID)
}

// UpdateIdentityProvider updates an identity provider
func (s *Service) UpdateIdentityProvider(ctx context.Context, id uuid.UUID, req *UpdateIdPRequest) (*federation.IdentityProvider, error) {
	provider, err := s.idpRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("identity provider not found: %w", err)
	}

	if req.Name != nil {
		provider.Name = *req.Name
	}
	if req.Enabled != nil {
		provider.Enabled = *req.Enabled
	}
	if req.Configuration != nil {
		// Validate configuration if provided
		if err := s.validateConfiguration(provider.Type, req.Configuration); err != nil {
			return nil, fmt.Errorf("invalid configuration: %w", err)
		}
		provider.Configuration = req.Configuration
	}
	if req.AttributeMapping != nil {
		provider.AttributeMapping = req.AttributeMapping
	}

	if err := s.idpRepo.Update(ctx, provider); err != nil {
		return nil, fmt.Errorf("failed to update identity provider: %w", err)
	}

	return provider, nil
}

// DeleteIdentityProvider deletes an identity provider
func (s *Service) DeleteIdentityProvider(ctx context.Context, id uuid.UUID) error {
	return s.idpRepo.Delete(ctx, id)
}

// InitiateOIDCLogin initiates an OIDC login flow
func (s *Service) InitiateOIDCLogin(ctx context.Context, tenantID uuid.UUID, providerID uuid.UUID, redirectURI string) (string, string, error) {
	// Get identity provider
	provider, err := s.idpRepo.GetByID(ctx, providerID)
	if err != nil {
		return "", "", fmt.Errorf("identity provider not found: %w", err)
	}

	if provider.Type != federation.IdentityProviderTypeOIDC {
		return "", "", fmt.Errorf("provider is not an OIDC provider")
	}

	if !provider.Enabled {
		return "", "", fmt.Errorf("identity provider is disabled")
	}

	// Verify tenant matches
	if provider.TenantID != tenantID {
		return "", "", fmt.Errorf("identity provider does not belong to tenant")
	}

	// Build OIDC configuration
	oidcConfig := s.buildOIDCConfig(provider.Configuration)

	// Create OIDC client
	client := oidcclient.NewClient(oidcConfig)

	// Perform discovery if needed
	if oidcConfig.AuthURL == "" || oidcConfig.TokenURL == "" {
		if _, err := client.Discover(ctx); err != nil {
			return "", "", fmt.Errorf("failed to discover OIDC provider: %w", err)
		}
	}

	// Generate state
	state, err := oidcclient.GenerateState()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate state: %w", err)
	}

	// Store state
	s.stateStore[state] = &State{
		ProviderID:  providerID,
		TenantID:    tenantID,
		RedirectURI: redirectURI,
		CreatedAt:   time.Now(),
	}

	// Generate authorization URL
	authURL, err := client.GenerateAuthorizationURL(redirectURI, state)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate authorization URL: %w", err)
	}

	return authURL, state, nil
}

// buildOIDCConfig builds OIDC configuration from provider config
func (s *Service) buildOIDCConfig(config map[string]interface{}) *federation.OIDCConfiguration {
	oidcConfig := &federation.OIDCConfiguration{}

	if clientID, ok := config["client_id"].(string); ok {
		oidcConfig.ClientID = clientID
	}
	if clientSecret, ok := config["client_secret"].(string); ok {
		oidcConfig.ClientSecret = clientSecret
	}
	if issuerURL, ok := config["issuer_url"].(string); ok {
		oidcConfig.IssuerURL = issuerURL
	}
	if authURL, ok := config["auth_url"].(string); ok {
		oidcConfig.AuthURL = authURL
	}
	if tokenURL, ok := config["token_url"].(string); ok {
		oidcConfig.TokenURL = tokenURL
	}
	if userInfoURL, ok := config["userinfo_url"].(string); ok {
		oidcConfig.UserInfoURL = userInfoURL
	}
	if scopes, ok := config["scopes"].([]interface{}); ok {
		oidcConfig.Scopes = make([]string, len(scopes))
		for i, scope := range scopes {
			if s, ok := scope.(string); ok {
				oidcConfig.Scopes[i] = s
			}
		}
	}

	return oidcConfig
}

// HandleOIDCCallback handles the OIDC callback
func (s *Service) HandleOIDCCallback(ctx context.Context, providerID uuid.UUID, code, state, redirectURI string) (*LoginResponse, error) {
	// Verify state
	storedState, ok := s.stateStore[state]
	if !ok {
		return nil, fmt.Errorf("invalid state")
	}

	// Clean up state (one-time use)
	delete(s.stateStore, state)

	// Verify state hasn't expired (5 minutes)
	if time.Since(storedState.CreatedAt) > 5*time.Minute {
		return nil, fmt.Errorf("state has expired")
	}

	// Verify provider ID matches
	if storedState.ProviderID != providerID {
		return nil, fmt.Errorf("provider ID mismatch")
	}

	// Get identity provider
	provider, err := s.idpRepo.GetByID(ctx, providerID)
	if err != nil {
		return nil, fmt.Errorf("identity provider not found: %w", err)
	}

	// Build OIDC configuration
	oidcConfig := s.buildOIDCConfig(provider.Configuration)
	client := oidcclient.NewClient(oidcConfig)

	// Exchange code for tokens
	tokenResp, err := client.ExchangeCode(ctx, code, redirectURI)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	// Validate ID token
	idTokenClaims, err := client.ValidateIDToken(ctx, tokenResp.IDToken)
	if err != nil {
		return nil, fmt.Errorf("failed to validate ID token: %w", err)
	}

	// Get user info (prefer UserInfo endpoint, fallback to ID token claims)
	var userInfo *oidcclient.UserInfo
	if oidcConfig.UserInfoURL != "" {
		userInfo, err = client.GetUserInfo(ctx, tokenResp.AccessToken)
		if err != nil {
			// Fallback to ID token claims
			userInfo = &oidcclient.UserInfo{
				Sub:               idTokenClaims.Sub,
				Email:             idTokenClaims.Email,
				EmailVerified:     idTokenClaims.EmailVerified,
				Name:              idTokenClaims.Name,
				GivenName:         idTokenClaims.GivenName,
				FamilyName:        idTokenClaims.FamilyName,
				PreferredUsername: idTokenClaims.PreferredUsername,
			}
		}
	} else {
		userInfo = &oidcclient.UserInfo{
			Sub:               idTokenClaims.Sub,
			Email:             idTokenClaims.Email,
			EmailVerified:     idTokenClaims.EmailVerified,
			Name:              idTokenClaims.Name,
			GivenName:         idTokenClaims.GivenName,
			FamilyName:        idTokenClaims.FamilyName,
			PreferredUsername: idTokenClaims.PreferredUsername,
		}
	}

	// Find or create user
	user, isNewUser, err := s.findOrCreateUser(ctx, provider, userInfo.Sub, userInfo, storedState.TenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to find or create user: %w", err)
	}

	// Link federated identity if new
	if isNewUser {
		fedIdentity := &federation.FederatedIdentity{
			ID:         uuid.New(),
			UserID:     user.ID,
			ProviderID: providerID,
			ExternalID: userInfo.Sub,
			Attributes: map[string]interface{}{
				"email":             userInfo.Email,
				"email_verified":    userInfo.EmailVerified,
				"name":              userInfo.Name,
				"preferred_username": userInfo.PreferredUsername,
			},
			IsPrimary: true,
			Verified:  true,
		}
		now := time.Now()
		fedIdentity.VerifiedAt = &now

		if err := s.fedIdRepo.Create(ctx, fedIdentity); err != nil {
			return nil, fmt.Errorf("failed to create federated identity: %w", err)
		}
	}

	// Build claims and generate tokens
	claimsObj, err := s.claimsBuilder.BuildClaims(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to build claims: %w", err)
	}

	// Generate tokens
	accessToken, err := s.tokenService.GenerateAccessToken(claimsObj, 15*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	idToken, err := s.tokenService.GenerateAccessToken(claimsObj, 60*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ID token: %w", err)
	}

	firstName := ""
	lastName := ""
	if user.FirstName != nil {
		firstName = *user.FirstName
	}
	if user.LastName != nil {
		lastName = *user.LastName
	}

	return &LoginResponse{
		UserID:      user.ID,
		Username:    user.Username,
		Email:       user.Email,
		FirstName:   firstName,
		LastName:    lastName,
		TenantID:    storedState.TenantID,
		IsNewUser:   isNewUser,
		AccessToken: accessToken,
		IDToken:     idToken,
	}, nil
}

// findOrCreateUser finds an existing user or creates a new one
func (s *Service) findOrCreateUser(ctx context.Context, provider *federation.IdentityProvider, externalID string, userInfo *oidcclient.UserInfo, tenantID uuid.UUID) (*models.User, bool, error) {
	// Try to find existing federated identity
	fedIdentity, err := s.fedIdRepo.GetByProviderAndExternalID(ctx, provider.ID, externalID)
	if err == nil {
		// User exists, return it
		user, err := s.userRepo.GetByID(ctx, fedIdentity.UserID)
		if err != nil {
			return nil, false, fmt.Errorf("failed to get user: %w", err)
		}
		return user, false, nil
	}

	// User doesn't exist, create new user
	// Map attributes
	email := userInfo.Email
	if email == "" {
		email = userInfo.PreferredUsername
	}
	if email == "" {
		email = fmt.Sprintf("%s@%s.local", externalID, provider.Name)
	}

	username := userInfo.PreferredUsername
	if username == "" {
		username = email
	}

	// Check if user with this email already exists (need tenant ID for tenant users)
	existingUser, _ := s.userRepo.GetByEmail(ctx, email, tenantID)
	if existingUser != nil {
		// Link to existing user
		fedIdentity := &federation.FederatedIdentity{
			ID:         uuid.New(),
			UserID:     existingUser.ID,
			ProviderID: provider.ID,
			ExternalID: externalID,
			Attributes: map[string]interface{}{
				"email": userInfo.Email,
			},
			IsPrimary: false,
			Verified:  true,
		}
		now := time.Now()
		fedIdentity.VerifiedAt = &now

		if err := s.fedIdRepo.Create(ctx, fedIdentity); err != nil {
			return nil, false, fmt.Errorf("failed to create federated identity: %w", err)
		}

		return existingUser, false, nil
	}

	// Create new user
	var firstName, lastName *string
	if userInfo.GivenName != "" {
		fn := userInfo.GivenName
		firstName = &fn
	}
	if userInfo.FamilyName != "" {
		ln := userInfo.FamilyName
		lastName = &ln
	}

	user := &models.User{
		ID:            uuid.New(),
		TenantID:      &tenantID,
		PrincipalType: models.PrincipalTypeTenant,
		Username:      username,
		Email:         email,
		FirstName:     firstName,
		LastName:      lastName,
		Status:        models.UserStatusActive,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, false, fmt.Errorf("failed to create user: %w", err)
	}

	return user, true, nil
}

// InitiateSAMLLogin initiates a SAML login flow
func (s *Service) InitiateSAMLLogin(ctx context.Context, tenantID uuid.UUID, providerID uuid.UUID, acsURL string) (string, error) {
	// Get identity provider
	provider, err := s.idpRepo.GetByID(ctx, providerID)
	if err != nil {
		return "", fmt.Errorf("identity provider not found: %w", err)
	}

	if provider.Type != federation.IdentityProviderTypeSAML {
		return "", fmt.Errorf("provider is not a SAML provider")
	}

	if !provider.Enabled {
		return "", fmt.Errorf("identity provider is disabled")
	}

	// Verify tenant matches
	if provider.TenantID != tenantID {
		return "", fmt.Errorf("identity provider does not belong to tenant")
	}

	// Build SAML configuration
	samlConfig := s.buildSAMLConfig(provider.Configuration)

	// Create SAML client
	client := samlclient.NewClient(samlConfig)

	// Get entity ID from config
	entityID, ok := provider.Configuration["entity_id"].(string)
	if !ok {
		return "", fmt.Errorf("entity_id not found in configuration")
	}

	// Generate AuthnRequest
	redirectURL, err := client.GenerateAuthnRequest(entityID, acsURL)
	if err != nil {
		return "", fmt.Errorf("failed to generate AuthnRequest: %w", err)
	}

	return redirectURL, nil
}

// buildSAMLConfig builds SAML configuration from provider config
func (s *Service) buildSAMLConfig(config map[string]interface{}) *federation.SAMLConfiguration {
	samlConfig := &federation.SAMLConfiguration{}

	if entityID, ok := config["entity_id"].(string); ok {
		samlConfig.EntityID = entityID
	}
	if ssoURL, ok := config["sso_url"].(string); ok {
		samlConfig.SSOURL = ssoURL
	}
	if sloURL, ok := config["slo_url"].(string); ok {
		samlConfig.SLOURL = sloURL
	}
	if cert, ok := config["x509_certificate"].(string); ok {
		samlConfig.X509Certificate = cert
	}
	if signRequests, ok := config["sign_requests"].(bool); ok {
		samlConfig.SignRequests = signRequests
	}
	if signAssertions, ok := config["sign_assertions"].(bool); ok {
		samlConfig.SignAssertions = signAssertions
	}
	if wantAssertionsSigned, ok := config["want_assertions_signed"].(bool); ok {
		samlConfig.WantAssertionsSigned = wantAssertionsSigned
	}

	return samlConfig
}

// HandleSAMLCallback handles the SAML callback
func (s *Service) HandleSAMLCallback(ctx context.Context, providerID uuid.UUID, samlResponse, relayState string) (*LoginResponse, error) {
	// Get identity provider
	provider, err := s.idpRepo.GetByID(ctx, providerID)
	if err != nil {
		return nil, fmt.Errorf("identity provider not found: %w", err)
	}

	// Build SAML configuration
	samlConfig := s.buildSAMLConfig(provider.Configuration)
	client := samlclient.NewClient(samlConfig)

	// Validate SAML response
	response, err := client.ValidateResponse(ctx, samlResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to validate SAML response: %w", err)
	}

	// Extract attributes
	attributes := client.ExtractAttributes(response)

	// Get external ID (NameID)
	externalID, ok := attributes["name_id"].(string)
	if !ok {
		return nil, fmt.Errorf("name_id not found in SAML response")
	}

	// Map attributes using attribute mapping
	email := s.mapAttribute(attributes, provider.AttributeMapping, "email")
	username := s.mapAttribute(attributes, provider.AttributeMapping, "username")
	firstName := s.mapAttribute(attributes, provider.AttributeMapping, "first_name")
	lastName := s.mapAttribute(attributes, provider.AttributeMapping, "last_name")

	if email == "" {
		email = fmt.Sprintf("%s@%s.local", externalID, provider.Name)
	}
	if username == "" {
		username = email
	}

	// Find or create user
	user, isNewUser, err := s.findOrCreateSAMLUser(ctx, provider, externalID, email, username, firstName, lastName, provider.TenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to find or create user: %w", err)
	}

	// Link federated identity if new
	if isNewUser {
		fedIdentity := &federation.FederatedIdentity{
			ID:         uuid.New(),
			UserID:     user.ID,
			ProviderID: providerID,
			ExternalID: externalID,
			Attributes: attributes,
			IsPrimary:  true,
			Verified:   true,
		}
		now := time.Now()
		fedIdentity.VerifiedAt = &now

		if err := s.fedIdRepo.Create(ctx, fedIdentity); err != nil {
			return nil, fmt.Errorf("failed to create federated identity: %w", err)
		}
	}

	// Build claims and generate tokens
	claimsObj, err := s.claimsBuilder.BuildClaims(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to build claims: %w", err)
	}

	// Generate tokens
	accessToken, err := s.tokenService.GenerateAccessToken(claimsObj, 15*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	idToken, err := s.tokenService.GenerateAccessToken(claimsObj, 60*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ID token: %w", err)
	}

	var firstName, lastName string
	if user.FirstName != nil {
		firstName = *user.FirstName
	}
	if user.LastName != nil {
		lastName = *user.LastName
	}

	return &LoginResponse{
		UserID:      user.ID,
		Username:    user.Username,
		Email:       user.Email,
		FirstName:   firstName,
		LastName:    lastName,
		TenantID:    provider.TenantID,
		IsNewUser:   isNewUser,
		AccessToken: accessToken,
		IDToken:     idToken,
	}, nil
}

// mapAttribute maps an attribute using the attribute mapping configuration
func (s *Service) mapAttribute(attributes map[string]interface{}, mapping map[string]interface{}, targetKey string) string {
	if mapping == nil {
		// No mapping, try direct lookup
		if val, ok := attributes[targetKey].(string); ok {
			return val
		}
		return ""
	}

	// Check if there's a mapping for this target key
	if mapConfig, ok := mapping[targetKey].(string); ok {
		if val, ok := attributes[mapConfig].(string); ok {
			return val
		}
	}

	// Fallback to direct lookup
	if val, ok := attributes[targetKey].(string); ok {
		return val
	}

	return ""
}

// findOrCreateSAMLUser finds an existing user or creates a new one for SAML
func (s *Service) findOrCreateSAMLUser(ctx context.Context, provider *federation.IdentityProvider, externalID, email, username, firstName, lastName string, tenantID uuid.UUID) (*models.User, bool, error) {
	// Try to find existing federated identity
	fedIdentity, err := s.fedIdRepo.GetByProviderAndExternalID(ctx, provider.ID, externalID)
	if err == nil {
		// User exists, return it
		user, err := s.userRepo.GetByID(ctx, fedIdentity.UserID)
		if err != nil {
			return nil, false, fmt.Errorf("failed to get user: %w", err)
		}
		return user, false, nil
	}

	// Check if user with this email already exists (need tenant ID for tenant users)
	existingUser, _ := s.userRepo.GetByEmail(ctx, email, tenantID)
	if existingUser != nil {
		// Link to existing user
		fedIdentity := &federation.FederatedIdentity{
			ID:         uuid.New(),
			UserID:     existingUser.ID,
			ProviderID: provider.ID,
			ExternalID: externalID,
			IsPrimary:  false,
			Verified:   true,
		}
		now := time.Now()
		fedIdentity.VerifiedAt = &now

		if err := s.fedIdRepo.Create(ctx, fedIdentity); err != nil {
			return nil, false, fmt.Errorf("failed to create federated identity: %w", err)
		}

		return existingUser, false, nil
	}

	// Create new user
	var firstNamePtr, lastNamePtr *string
	if firstName != "" {
		firstNamePtr = &firstName
	}
	if lastName != "" {
		lastNamePtr = &lastName
	}

	user := &models.User{
		ID:            uuid.New(),
		TenantID:      &tenantID,
		PrincipalType: models.PrincipalTypeTenant,
		Username:      username,
		Email:         email,
		FirstName:     firstNamePtr,
		LastName:      lastNamePtr,
		Status:        models.UserStatusActive,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, false, fmt.Errorf("failed to create user: %w", err)
	}

	return user, true, nil
}

