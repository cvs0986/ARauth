package saml

import (
	"context"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"encoding/xml"
	"fmt"
	"net/url"
	"time"

	"github.com/arauth-identity/iam/identity/federation"
)

// Client handles SAML authentication flows
type Client struct {
	config *federation.SAMLConfiguration
}

// NewClient creates a new SAML client
func NewClient(config *federation.SAMLConfiguration) *Client {
	return &Client{
		config: config,
	}
}

// AuthnRequest represents a SAML authentication request
type AuthnRequest struct {
	XMLName                xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:protocol AuthnRequest"`
	ID                     string   `xml:"ID,attr"`
	Version                string   `xml:"Version,attr"`
	IssueInstant           string   `xml:"IssueInstant,attr"`
	Destination            string   `xml:"Destination,attr"`
	AssertionConsumerServiceURL string `xml:"AssertionConsumerServiceURL,attr"`
	Issuer                 Issuer   `xml:"Issuer"`
}

// Issuer represents a SAML issuer
type Issuer struct {
	XMLName xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:assertion Issuer"`
	Value   string   `xml:",chardata"`
}

// GenerateAuthnRequest generates a SAML AuthnRequest
func (c *Client) GenerateAuthnRequest(entityID, acsURL string) (string, error) {
	requestID, err := generateRequestID()
	if err != nil {
		return "", fmt.Errorf("failed to generate request ID: %w", err)
	}

	authnRequest := AuthnRequest{
		ID:                     requestID,
		Version:                "2.0",
		IssueInstant:           time.Now().UTC().Format(time.RFC3339),
		Destination:            c.config.SSOURL,
		AssertionConsumerServiceURL: acsURL,
		Issuer: Issuer{
			Value: entityID,
		},
	}

	xmlData, err := xml.MarshalIndent(authnRequest, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal AuthnRequest: %w", err)
	}

	// Base64 encode
	encoded := base64.StdEncoding.EncodeToString(xmlData)

	// Create redirect URL
	redirectURL := fmt.Sprintf("%s?SAMLRequest=%s", c.config.SSOURL, urlEncode(encoded))

	return redirectURL, nil
}

// generateRequestID generates a unique request ID
func generateRequestID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate request ID: %w", err)
	}
	return fmt.Sprintf("_%x", b), nil
}

// urlEncode URL encodes a string
func urlEncode(s string) string {
	return url.QueryEscape(s)
}

// Response represents a SAML response
type Response struct {
	XMLName      xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:protocol Response"`
	ID           string   `xml:"ID,attr"`
	Version      string   `xml:"Version,attr"`
	IssueInstant string   `xml:"IssueInstant,attr"`
	Destination  string   `xml:"Destination,attr"`
	Issuer       Issuer   `xml:"Issuer"`
	Status       Status   `xml:"Status"`
	Assertion    Assertion `xml:"Assertion"`
}

// Status represents SAML status
type Status struct {
	XMLName             xml.Name      `xml:"urn:oasis:names:tc:SAML:2.0:protocol Status"`
	StatusCode          StatusCode     `xml:"StatusCode"`
}

// StatusCode represents SAML status code
type StatusCode struct {
	XMLName xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:protocol StatusCode"`
	Value   string   `xml:"Value,attr"`
}

// Assertion represents a SAML assertion
type Assertion struct {
	XMLName      xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:assertion Assertion"`
	ID           string   `xml:"ID,attr"`
	Version      string   `xml:"Version,attr"`
	IssueInstant string   `xml:"IssueInstant,attr"`
	Issuer       Issuer   `xml:"Issuer"`
	Subject      Subject  `xml:"Subject"`
	AttributeStatement AttributeStatement `xml:"AttributeStatement"`
}

// Subject represents a SAML subject
type Subject struct {
	XMLName             xml.Name      `xml:"urn:oasis:names:tc:SAML:2.0:assertion Subject"`
	NameID              NameID        `xml:"NameID"`
}

// NameID represents a SAML NameID
type NameID struct {
	XMLName xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:assertion NameID"`
	Format  string   `xml:"Format,attr"`
	Value   string   `xml:",chardata"`
}

// AttributeStatement represents SAML attribute statement
type AttributeStatement struct {
	XMLName    xml.Name    `xml:"urn:oasis:names:tc:SAML:2.0:assertion AttributeStatement"`
	Attributes []Attribute `xml:"Attribute"`
}

// Attribute represents a SAML attribute
type Attribute struct {
	XMLName xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:assertion Attribute"`
	Name    string   `xml:"Name,attr"`
	Values  []string `xml:"AttributeValue"`
}

// ValidateResponse validates a SAML response
func (c *Client) ValidateResponse(ctx context.Context, samlResponse string) (*Response, error) {
	// Decode base64
	decoded, err := base64.StdEncoding.DecodeString(samlResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to decode SAML response: %w", err)
	}

	// Parse XML
	var response Response
	if err := xml.Unmarshal(decoded, &response); err != nil {
		return nil, fmt.Errorf("failed to parse SAML response: %w", err)
	}

	// Basic validation
	if response.Status.StatusCode.Value != "urn:oasis:names:tc:SAML:2.0:status:Success" {
		return nil, fmt.Errorf("SAML response status is not success: %s", response.Status.StatusCode.Value)
	}

	// Verify signature if configured
	if c.config.WantAssertionsSigned {
		if err := c.verifySignature(decoded); err != nil {
			return nil, fmt.Errorf("failed to verify SAML signature: %w", err)
		}
	}

	return &response, nil
}

// verifySignature verifies the SAML response signature
func (c *Client) verifySignature(xmlData []byte) error {
	// In production, implement proper XML signature verification
	// This requires parsing the XML, extracting the signature,
	// and verifying it against the IdP's certificate

	if c.config.X509Certificate == "" {
		return fmt.Errorf("X509 certificate not configured")
	}

	// Parse certificate
	cert, err := parseCertificate(c.config.X509Certificate)
	if err != nil {
		return fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Verify certificate is valid
	if time.Now().After(cert.NotAfter) || time.Now().Before(cert.NotBefore) {
		return fmt.Errorf("certificate is expired or not yet valid")
	}

	// TODO: Implement actual XML signature verification
	// This is a placeholder - proper implementation requires:
	// 1. Extract signature from XML
	// 2. Verify signature using certificate
	// 3. Verify signed elements

	return nil
}

// parseCertificate parses an X509 certificate
func parseCertificate(certPEM string) (*x509.Certificate, error) {
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to parse certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	return cert, nil
}

// ExtractAttributes extracts user attributes from a SAML response
func (c *Client) ExtractAttributes(response *Response) map[string]interface{} {
	attributes := make(map[string]interface{})

	// Extract NameID
	if response.Assertion.Subject.NameID.Value != "" {
		attributes["name_id"] = response.Assertion.Subject.NameID.Value
		attributes["name_id_format"] = response.Assertion.Subject.NameID.Format
	}

	// Extract attributes from AttributeStatement
	for _, attr := range response.Assertion.AttributeStatement.Attributes {
		if len(attr.Values) > 0 {
			attributes[attr.Name] = attr.Values[0] // Use first value
		}
	}

	return attributes
}

