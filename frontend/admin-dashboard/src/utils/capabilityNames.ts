/**
 * Utility functions for formatting capability names and descriptions
 */

/**
 * Converts a capability key (e.g., "allowed_grant_types") to a display name (e.g., "Allowed Grant Types")
 */
export function formatCapabilityName(key: string): string {
  return key
    .split('_')
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ');
}

/**
 * Gets a user-friendly description for a capability key that explains what it does and how it works
 */
export function getCapabilityDescription(key: string, description?: string): string {
  if (description) {
    return description;
  }
  
  // Detailed descriptions that explain what the feature does and how it works
  const descriptions: Record<string, string> = {
    allowed_grant_types: 'Controls which OAuth2 grant types are permitted for authentication flows. Common types include authorization_code, client_credentials, and refresh_token. This determines how applications can obtain access tokens.',
    allowed_scope_namespaces: 'Defines the scope namespaces that can be used in OAuth2/OIDC requests. Scopes like "openid", "profile", "users", and "clients" control what information and permissions are accessible through tokens.',
    max_token_ttl: 'Sets the maximum time-to-live (TTL) for access tokens. Tokens expire after this duration to enhance security. Default is 15 minutes. Shorter TTLs improve security but may require more frequent token refreshes.',
    mfa: 'Enables multi-factor authentication (MFA) for enhanced security. When enabled, users must provide a second authentication factor (like a code from an authenticator app) in addition to their password.',
    oauth2: 'Enables OAuth2 protocol support, allowing third-party applications to securely access user resources without sharing passwords. OAuth2 is the industry-standard authorization framework.',
    oidc: 'Enables OpenID Connect (OIDC) support, which extends OAuth2 to provide identity information. OIDC allows applications to verify user identity and obtain basic profile information.',
    pkce_mandatory: 'Makes Proof Key for Code Exchange (PKCE) mandatory for OAuth flows. PKCE adds an extra layer of security for public clients by preventing authorization code interception attacks.',
    totp: 'Enables Time-based One-Time Password (TOTP) support for MFA. Users can use authenticator apps (like Google Authenticator) to generate time-based codes for two-factor authentication.',
    ldap: 'Enables LDAP/Active Directory integration, allowing users to authenticate using their existing corporate credentials. This connects your identity system with enterprise directory services.',
    saml: 'Enables SAML (Security Assertion Markup Language) federation support. Allows single sign-on (SSO) with external identity providers and enterprise systems using SAML protocol.',
    passwordless: 'Enables passwordless authentication methods. Users can authenticate using alternative methods like magic links, biometrics, or hardware keys instead of traditional passwords.',
  };
  
  return descriptions[key] || 'No description available';
}

