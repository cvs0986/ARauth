# Performance Testing Guide

This document describes how to run performance benchmarks and load tests for the Nuage Identity IAM platform.

## Benchmarks

### Running Benchmarks

```bash
# Run all benchmarks
make benchmark

# Run benchmarks for specific packages
go test -bench=. -benchmem ./security/password/...
go test -bench=. -benchmem ./security/totp/...
go test -bench=. -benchmem ./security/encryption/...

# Run with CPU profiling
go test -bench=. -benchmem -cpuprofile=cpu.prof ./security/password/...
go tool pprof cpu.prof
```

### Benchmark Targets

#### Password Security
- **Hash**: Password hashing with Argon2id
- **Verify**: Password verification
- **HashLongPassword**: Hashing with longer passwords

**Expected Performance**:
- Hash: ~100-200ms per operation (Argon2id is intentionally slow)
- Verify: ~100-200ms per operation

#### TOTP
- **GenerateSecret**: TOTP secret generation
- **Validate**: TOTP code validation
- **GenerateQRCode**: QR code generation for MFA setup

**Expected Performance**:
- GenerateSecret: <1ms
- Validate: <1ms
- GenerateQRCode: ~10-50ms

#### Encryption
- **Encrypt**: AES-GCM encryption
- **Decrypt**: AES-GCM decryption
- **EncryptDecrypt**: Full cycle

**Expected Performance**:
- Encrypt: <1ms
- Decrypt: <1ms
- EncryptDecrypt: <2ms

#### API Handlers
- **HealthCheck**: Health endpoint
- **HealthLive**: Liveness endpoint
- **HealthReady**: Readiness endpoint

**Expected Performance**:
- All health endpoints: <1ms

## Load Testing

### Using the Performance Test Script

```bash
# Set environment variables
export API_URL="http://localhost:8080"
export TENANT_ID="your-tenant-id"
export CONCURRENT_USERS=100
export TOTAL_REQUESTS=10000

# Run performance tests
./scripts/performance-test.sh
```

### Using hey

```bash
# Install hey
go install github.com/rakyll/hey@latest

# Health check endpoint
hey -n 10000 -c 100 -m GET http://localhost:8080/health

# Authenticated endpoint
hey -n 5000 -c 50 -m GET \
  -H "X-Tenant-ID: tenant-id" \
  http://localhost:8080/api/v1/users
```

### Using Apache Bench (ab)

```bash
# Install Apache Bench
sudo apt-get install apache2-utils  # Ubuntu/Debian
sudo yum install httpd-tools         # CentOS/RHEL

# Health check
ab -n 10000 -c 100 http://localhost:8080/health
```

### Load Test Scenarios

#### 1. Health Check Endpoint
```bash
hey -n 100000 -c 1000 -m GET http://localhost:8080/health
```
**Target**: <10ms p95 latency, >1000 req/s

#### 2. User List Endpoint
```bash
hey -n 10000 -c 100 -m GET \
  -H "X-Tenant-ID: tenant-id" \
  http://localhost:8080/api/v1/users
```
**Target**: <50ms p95 latency, >200 req/s

#### 3. Login Endpoint
```bash
hey -n 5000 -c 50 -m POST \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: tenant-id" \
  -d '{"username":"testuser","password":"SecurePass123!","tenant_id":"tenant-id"}' \
  http://localhost:8080/api/v1/auth/login
```
**Target**: <500ms p95 latency (due to Argon2id), >10 req/s

#### 4. MFA Enrollment
```bash
hey -n 1000 -c 10 -m POST \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: tenant-id" \
  -d '{"user_id":"user-id"}' \
  http://localhost:8080/api/v1/mfa/enroll
```
**Target**: <100ms p95 latency, >50 req/s

## Performance Targets

### API Endpoints

| Endpoint | Target p95 Latency | Target Throughput |
|----------|-------------------|-------------------|
| Health Check | <10ms | >1000 req/s |
| User List | <50ms | >200 req/s |
| User Create | <100ms | >100 req/s |
| Login | <500ms | >10 req/s |
| MFA Enroll | <100ms | >50 req/s |
| MFA Verify | <50ms | >100 req/s |

### Security Operations

| Operation | Target Latency |
|-----------|---------------|
| Password Hash | <200ms (Argon2id) |
| Password Verify | <200ms (Argon2id) |
| TOTP Generate | <1ms |
| TOTP Validate | <1ms |
| Encryption | <1ms |
| Decryption | <1ms |

## Profiling

### CPU Profiling

```bash
# Run benchmark with CPU profiling
go test -bench=. -benchmem -cpuprofile=cpu.prof ./security/password/...

# Analyze profile
go tool pprof cpu.prof
(pprof) top
(pprof) web
```

### Memory Profiling

```bash
# Run benchmark with memory profiling
go test -bench=. -benchmem -memprofile=mem.prof ./security/password/...

# Analyze profile
go tool pprof mem.prof
(pprof) top
(pprof) web
```

## Continuous Performance Monitoring

### CI/CD Integration

Add to `.github/workflows/ci.yml`:

```yaml
- name: Run benchmarks
  run: make benchmark

- name: Performance regression check
  run: |
    go test -bench=. -benchmem ./security/password/... > benchmark.txt
    # Compare with baseline
```

## Performance Optimization Tips

1. **Password Hashing**: Argon2id is intentionally slow for security. Don't optimize this.
2. **Database Queries**: Use indexes, connection pooling, and prepared statements.
3. **Caching**: Cache frequently accessed data (users, tenants, roles).
4. **Connection Pooling**: Configure appropriate pool sizes for database and Redis.
5. **Rate Limiting**: Implement rate limiting to prevent abuse.
6. **Async Operations**: Use goroutines for non-blocking operations where possible.

## Troubleshooting

### High Latency

1. Check database connection pool settings
2. Verify indexes are being used
3. Check Redis connection and latency
4. Review application logs for slow queries

### Low Throughput

1. Increase connection pool sizes
2. Enable caching for frequently accessed data
3. Review rate limiting settings
4. Check for bottlenecks in middleware

### Memory Issues

1. Review memory profiles
2. Check for memory leaks
3. Optimize data structures
4. Review cache TTL settings

