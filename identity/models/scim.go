package models

import (
	"time"

	"github.com/google/uuid"
)

// SCIMToken represents a SCIM API authentication token
type SCIMToken struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	TenantID   uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	Name       string     `json:"name" db:"name"`
	TokenHash  string     `json:"-" db:"token_hash"` // Never expose hash (bcrypt)
	LookupHash string     `json:"-" db:"lookup_hash"` // Never expose hash (SHA256 for lookup)
	Scopes     []string   `json:"scopes" db:"scopes"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty" db:"last_used_at"`
	CreatedBy  *uuid.UUID `json:"created_by,omitempty" db:"created_by"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// IsExpired returns true if the token has expired
func (t *SCIMToken) IsExpired() bool {
	if t.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*t.ExpiresAt)
}

// IsDeleted returns true if the token is soft-deleted
func (t *SCIMToken) IsDeleted() bool {
	return t.DeletedAt != nil
}

// SCIMUser represents a SCIM 2.0 User resource
type SCIMUser struct {
	Schemas    []string              `json:"schemas"`
	ID         string                `json:"id"`
	ExternalID string                `json:"externalId,omitempty"`
	UserName   string                `json:"userName"`
	Name       SCIMName              `json:"name"`
	DisplayName string               `json:"displayName,omitempty"`
	NickName   string                `json:"nickName,omitempty"`
	ProfileURL string                `json:"profileUrl,omitempty"`
	Title      string                `json:"title,omitempty"`
	UserType   string                `json:"userType,omitempty"`
	PreferredLanguage string         `json:"preferredLanguage,omitempty"`
	Locale     string                `json:"locale,omitempty"`
	Timezone   string                `json:"timezone,omitempty"`
	Active     bool                  `json:"active"`
	Password   string                `json:"password,omitempty"` // Only for create/update
	Emails     []SCIMEmail            `json:"emails"`
	PhoneNumbers []SCIMPhoneNumber    `json:"phoneNumbers,omitempty"`
	IMS        []SCIMIMS              `json:"ims,omitempty"`
	Photos     []SCIMPhoto            `json:"photos,omitempty"`
	Addresses  []SCIMAddress          `json:"addresses,omitempty"`
	Groups     []SCIMGroupReference   `json:"groups,omitempty"`
	Entitlements []SCIMEntitlement    `json:"entitlements,omitempty"`
	Roles      []SCIMRole             `json:"roles,omitempty"`
	X509Certificates []SCIMX509Certificate `json:"x509Certificates,omitempty"`
	Meta       SCIMMeta               `json:"meta"`
}

// SCIMName represents a SCIM name structure
type SCIMName struct {
	Formatted  string `json:"formatted,omitempty"`
	FamilyName string `json:"familyName,omitempty"`
	GivenName  string `json:"givenName,omitempty"`
	MiddleName string `json:"middleName,omitempty"`
	HonorificPrefix string `json:"honorificPrefix,omitempty"`
	HonorificSuffix string `json:"honorificSuffix,omitempty"`
}

// SCIMEmail represents a SCIM email
type SCIMEmail struct {
	Value   string `json:"value"`
	Type    string `json:"type,omitempty"`
	Primary bool   `json:"primary,omitempty"`
	Display string `json:"display,omitempty"`
}

// SCIMPhoneNumber represents a SCIM phone number
type SCIMPhoneNumber struct {
	Value   string `json:"value"`
	Type    string `json:"type,omitempty"`
	Primary bool   `json:"primary,omitempty"`
}

// SCIMIMS represents an Instant Messaging address
type SCIMIMS struct {
	Value   string `json:"value"`
	Type    string `json:"type,omitempty"`
	Primary bool   `json:"primary,omitempty"`
}

// SCIMPhoto represents a photo URL
type SCIMPhoto struct {
	Value   string `json:"value"`
	Type    string `json:"type,omitempty"`
	Primary bool   `json:"primary,omitempty"`
}

// SCIMAddress represents a physical address
type SCIMAddress struct {
	Formatted     string `json:"formatted,omitempty"`
	StreetAddress string `json:"streetAddress,omitempty"`
	Locality      string `json:"locality,omitempty"`
	Region        string `json:"region,omitempty"`
	PostalCode    string `json:"postalCode,omitempty"`
	Country       string `json:"country,omitempty"`
	Type          string `json:"type,omitempty"`
	Primary       bool   `json:"primary,omitempty"`
}

// SCIMGroupReference represents a group reference
type SCIMGroupReference struct {
	Value   string `json:"value"`
	Ref     string `json:"$ref,omitempty"`
	Display string `json:"display,omitempty"`
	Type    string `json:"type,omitempty"`
}

// SCIMEntitlement represents an entitlement
type SCIMEntitlement struct {
	Value   string `json:"value"`
	Display string `json:"display,omitempty"`
	Type    string `json:"type,omitempty"`
	Primary bool   `json:"primary,omitempty"`
}

// SCIMRole represents a role
type SCIMRole struct {
	Value   string `json:"value"`
	Display string `json:"display,omitempty"`
	Type    string `json:"type,omitempty"`
	Primary bool   `json:"primary,omitempty"`
}

// SCIMX509Certificate represents an X.509 certificate
type SCIMX509Certificate struct {
	Value   string `json:"value"`
	Display string `json:"display,omitempty"`
	Type    string `json:"type,omitempty"`
	Primary bool   `json:"primary,omitempty"`
}

// SCIMGroup represents a SCIM 2.0 Group resource
type SCIMGroup struct {
	Schemas     []string            `json:"schemas"`
	ID          string              `json:"id"`
	ExternalID  string              `json:"externalId,omitempty"`
	DisplayName string              `json:"displayName"`
	Description string              `json:"description,omitempty"`
	Members     []SCIMGroupMember   `json:"members,omitempty"`
	Meta        SCIMMeta            `json:"meta"`
}

// SCIMGroupMember represents a group member
type SCIMGroupMember struct {
	Value   string `json:"value"`
	Ref     string `json:"$ref,omitempty"`
	Display string `json:"display,omitempty"`
	Type    string `json:"type,omitempty"`
}

// SCIMMeta represents SCIM metadata
type SCIMMeta struct {
	ResourceType string    `json:"resourceType"`
	Created      time.Time `json:"created"`
	LastModified time.Time `json:"lastModified"`
	Location     string    `json:"location,omitempty"`
	Version      string    `json:"version,omitempty"`
}

// SCIMListResponse represents a SCIM list response
type SCIMListResponse struct {
	Schemas      []string      `json:"schemas"`
	TotalResults int           `json:"totalResults"`
	ItemsPerPage int           `json:"itemsPerPage"`
	StartIndex   int           `json:"startIndex"`
	Resources    []interface{} `json:"Resources"`
}

// SCIMError represents a SCIM error response
type SCIMError struct {
	Schemas  []string `json:"schemas"`
	Detail   string   `json:"detail"`
	Status   string   `json:"status"`
	SCIMType string   `json:"scimType,omitempty"`
}

