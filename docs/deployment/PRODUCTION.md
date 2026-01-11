# Production Deployment Guide

**Last Updated**: 2026-01-11  
**Version**: 1.0.0

---

## Prerequisites

- PostgreSQL 14+ (database)
- Redis 6+ (caching & rate limiting) **REQUIRED**
- Docker/Kubernetes (recommended)
- SSL/TLS certificates
- Domain name with DNS configured

---

## Environment Configuration

### Required Environment Variables

Copy `.env.production.example` to `.env` and configure:

```bash
cp .env.production.example .env
```

#### Critical Security Settings

| Variable | Description | Example | Required |
|----------|-------------|---------|----------|
| `ENVIRONMENT` | Deployment environment | `production` | ✅ Yes |
| `DATABASE_URL` | PostgreSQL connection string | `postgresql://user:pass@host:5432/arauth` | ✅ Yes |
| `REDIS_URL` | Redis connection string | `redis://host:6379/0` | ✅ Yes |
| `JWT_SIGNING_KEY` | JWT signing key (32+ bytes) | `<random-32-byte-key>` | ✅ Yes |
| `ENCRYPTION_KEY` | AES-256 encryption key (exactly 32 bytes) | `<random-32-byte-key>` | ✅ Yes |

**⚠️ CRITICAL**: Never use default/example keys in production!

Generate secure keys:
```bash
# JWT Signing Key
openssl rand -base64 32

# Encryption Key (must be exactly 32 bytes)
openssl rand -hex 16
```

---

## Rate Limiting Configuration

### Overview

ARauth implements multi-tier rate limiting to prevent abuse:
- **Per-User**: Limits requests from individual users
- **Per-Client**: Limits requests from OAuth clients
- **Per-IP**: Limits requests from IP addresses (admin endpoints)

### Default Limits (Conservative)

| Category | Default RPM | Burst | Environment Variable |
|----------|-------------|-------|---------------------|
| User (General) | 60 | +10 | `RATE_LIMIT_USER_RPM` |
| OAuth Client | 100 | +20 | `RATE_LIMIT_CLIENT_RPM` |
| Admin IP | 30 | +5 | `RATE_LIMIT_ADMIN_IP_RPM` |

### Endpoint-Specific Limits (Built-in)

These limits are **hardcoded** and cannot be overridden:

| Endpoint Category | RPM | Burst | Examples |
|-------------------|-----|-------|----------|
| **Auth** | 20 | +3 | `/api/v1/auth/login`, `/api/v1/auth/token` |
| **Sensitive** | 10 | +2 | MFA enrollment, password reset, user suspension |

### Tuning Rate Limits

#### When to Increase Limits

- High legitimate traffic volume
- Automated integrations (use OAuth clients)
- Batch operations

#### When to Decrease Limits

- Experiencing abuse/attacks
- Limited infrastructure capacity
- Stricter security requirements

#### How to Tune Safely

1. **Start with defaults** (already conservative)
2. **Monitor metrics** (Phase C.2 will add observability)
3. **Adjust incrementally** (±20% at a time)
4. **Test thoroughly** before deploying

Example:
```bash
# Increase user limit for high-traffic application
RATE_LIMIT_USER_RPM=120

# Decrease admin IP limit for stricter security
RATE_LIMIT_ADMIN_IP_RPM=15
```

### Rate Limit Headers

Clients receive rate limit information in response headers:

```
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 45
X-RateLimit-Reset: 1704988800
Retry-After: 30
```

### Rate Limit Errors

When rate limited, clients receive:

**HTTP 429 Too Many Requests**
```json
{
  "error": "rate_limit_exceeded",
  "message": "Rate limit exceeded. Please retry after 30s.",
  "details": {
    "limit_type": "user",
    "limit": 60,
    "current_count": 71,
    "retry_after": 30
  }
}
```

---

## Startup Validation

### Production Safety Checks

ARauth validates configuration on startup:

#### ✅ Required Checks (Fail-Fast)

- `DATABASE_URL` is set and reachable
- `REDIS_URL` is set and reachable (REQUIRED for rate limiting)
- `JWT_SIGNING_KEY` is set and strong
- `ENCRYPTION_KEY` is exactly 32 bytes
- `ENVIRONMENT=production` when deployed to production

#### ⚠️ Warning Checks

- Debug mode enabled in production
- CORS allows all origins
- Default/weak JWT signing key
- TLS disabled

### Startup Failure Examples

```
FATAL: Redis is required for rate limiting in production
```
**Solution**: Ensure Redis is running and `REDIS_URL` is correct.

```
FATAL: Encryption key must be exactly 32 bytes (AES-256)
```
**Solution**: Generate a 32-byte encryption key.

---

## Health Checks

### Endpoints

| Endpoint | Purpose | Use Case |
|----------|---------|----------|
| `GET /health` | Overall health | Load balancer health check |
| `GET /health/live` | Liveness probe | Kubernetes liveness |
| `GET /health/ready` | Readiness probe | Kubernetes readiness |

### Health Check Bypass

Health checks are **exempt from rate limiting** to prevent false negatives.

---

## Deployment Checklist

### Pre-Deployment

- [ ] All environment variables configured
- [ ] Secure keys generated (not defaults)
- [ ] Database migrations applied
- [ ] Redis accessible and persistent
- [ ] SSL/TLS certificates valid
- [ ] CORS origins configured correctly

### Deployment

- [ ] Deploy during low-traffic window
- [ ] Monitor startup logs for errors
- [ ] Verify health checks pass
- [ ] Test authentication flow
- [ ] Verify rate limiting active

### Post-Deployment

- [ ] Monitor error rates
- [ ] Check rate limit violations
- [ ] Verify audit logs
- [ ] Test failover scenarios

---

## Monitoring

### Key Metrics (Phase C.2 will add)

- Request rate per endpoint
- Rate limit violations
- Authentication success/failure rate
- Token validation errors
- Database connection pool usage
- Redis connection health

### Logs to Monitor

```
INFO  Rate limiter initialized  user_rpm=60 client_rpm=100 admin_ip_rpm=30
WARN  Rate limiting disabled - Redis not available (NOT SAFE FOR PRODUCTION)
FATAL Redis is required for rate limiting in production
```

---

## Troubleshooting

### Rate Limiting Not Working

**Symptom**: No rate limiting applied

**Causes**:
1. Redis not connected
2. `ENVIRONMENT != production` (falls back to legacy limiter)
3. Middleware not applied

**Solution**:
```bash
# Check logs for:
INFO  Rate limiter initialized

# If you see:
WARN  Rate limiting disabled

# Then Redis is not available
```

### Too Many Rate Limit Errors

**Symptom**: Legitimate users getting 429 errors

**Causes**:
1. Limits too strict for traffic volume
2. Burst allowance too low
3. Shared IP (NAT/proxy)

**Solution**:
1. Increase `RATE_LIMIT_USER_RPM`
2. Monitor and tune incrementally
3. Consider IP allowlisting for trusted sources

---

## Security Best Practices

### Rate Limiting

✅ **DO**:
- Keep Redis persistent (AOF or RDB)
- Monitor rate limit violations
- Tune limits based on actual traffic
- Use OAuth clients for integrations

❌ **DON'T**:
- Disable rate limiting in production
- Set limits too high (defeats purpose)
- Ignore rate limit violation spikes
- Use same limits for all endpoint types

### General

✅ **DO**:
- Use strong, random keys
- Enable TLS/SSL
- Restrict CORS origins
- Monitor audit logs
- Keep dependencies updated

❌ **DON'T**:
- Use default/example keys
- Expose admin endpoints publicly
- Allow all CORS origins
- Ignore security warnings

---

## Scaling Considerations

### Horizontal Scaling

ARauth is stateless and can scale horizontally:

- **Load Balancer**: Distribute traffic across instances
- **Shared Redis**: All instances share rate limit state
- **Shared Database**: PostgreSQL with connection pooling

### Redis Scaling

For high-traffic deployments:

- **Redis Cluster**: Distribute rate limit data
- **Redis Sentinel**: High availability
- **Separate Redis**: Dedicated instance for rate limiting

---

## Support

For issues or questions:
- GitHub Issues: [arauth-identity/iam](https://github.com/arauth-identity/iam)
- Documentation: `/docs`
- Security Issues: security@your-domain.com

---

**Last Updated**: 2026-01-11  
**Next Review**: Phase C.2 (Observability)
