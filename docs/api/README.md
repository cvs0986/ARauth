# API Documentation

This directory contains the OpenAPI/Swagger specification for the ARauth Identity IAM API.

## OpenAPI Specification

The API is documented using OpenAPI 3.0.3 specification:
- `openapi.yaml` - Complete API specification

## Viewing the Documentation

### Using Swagger UI

1. **Online**: Upload `openapi.yaml` to https://editor.swagger.io/

2. **Local Swagger UI**:
```bash
# Install swagger-ui
docker run -p 8080:8080 -e SWAGGER_JSON=/api/openapi.yaml -v $(pwd)/docs/api:/api swaggerapi/swagger-ui

# Or use npx
npx swagger-ui-serve docs/api/openapi.yaml
```

3. **Redoc**:
```bash
npx redoc-cli serve docs/api/openapi.yaml
```

### Generate Documentation

```bash
# Generate HTML from OpenAPI spec
npx @redocly/cli build-docs docs/api/openapi.yaml -o docs/api/index.html
```

## API Endpoints

### Health Checks
- `GET /health` - Health check
- `GET /health/live` - Liveness probe
- `GET /health/ready` - Readiness probe

### Tenants
- `POST /tenants` - Create tenant
- `GET /tenants` - List tenants
- `GET /tenants/{id}` - Get tenant
- `PUT /tenants/{id}` - Update tenant
- `DELETE /tenants/{id}` - Delete tenant

### Users (Tenant-scoped)
- `POST /users` - Create user
- `GET /users` - List users
- `GET /users/{id}` - Get user
- `PUT /users/{id}` - Update user
- `DELETE /users/{id}` - Delete user

### Authentication
- `POST /auth/login` - Login (OAuth2 flow)

### MFA
- `POST /mfa/enroll` - Enroll in MFA
- `POST /mfa/verify` - Verify MFA code
- `POST /mfa/challenge` - Challenge MFA
- `POST /mfa/challenge/verify` - Verify challenge

### Roles (Tenant-scoped)
- `POST /roles` - Create role
- `GET /roles` - List roles
- `GET /roles/{id}` - Get role
- `PUT /roles/{id}` - Update role
- `DELETE /roles/{id}` - Delete role
- `GET /roles/{id}/permissions` - Get role permissions
- `POST /roles/{id}/permissions/{permission_id}` - Assign permission
- `DELETE /roles/{id}/permissions/{permission_id}` - Remove permission

### Permissions (Tenant-scoped)
- `POST /permissions` - Create permission
- `GET /permissions` - List permissions
- `GET /permissions/{id}` - Get permission
- `PUT /permissions/{id}` - Update permission
- `DELETE /permissions/{id}` - Delete permission

### User-Role Assignments
- `GET /users/{user_id}/roles` - Get user roles
- `POST /users/{user_id}/roles/{role_id}` - Assign role to user
- `DELETE /users/{user_id}/roles/{role_id}` - Remove role from user

## Authentication

Most endpoints require tenant context via the `X-Tenant-ID` header.

OAuth2/OIDC authentication is handled through ORY Hydra. See the integration guide for details.

## Multi-Tenancy

All user, role, and permission operations are tenant-scoped. The tenant ID must be provided via:
- `X-Tenant-ID` header (recommended)
- `X-Tenant-Domain` header
- Query parameter `tenant_id`
- Subdomain extraction (if configured)

## Error Responses

All errors follow a consistent format:

```json
{
  "error": "error_code",
  "message": "Human-readable error message",
  "details": {}
}
```

Common HTTP status codes:
- `400` - Bad Request
- `401` - Unauthorized
- `403` - Forbidden
- `404` - Not Found
- `429` - Too Many Requests
- `500` - Internal Server Error
- `503` - Service Unavailable

## Rate Limiting

Rate limiting is applied globally and per-tenant. Check response headers:
- `X-RateLimit-Limit` - Request limit
- `X-RateLimit-Remaining` - Remaining requests
- `X-RateLimit-Reset` - Reset timestamp
- `Retry-After` - Seconds to wait (if rate limited)

## Examples

See the integration guide for complete examples:
- [Integration Guide](../guides/integration-guide.md)
- [Getting Started](../guides/getting-started.md)

