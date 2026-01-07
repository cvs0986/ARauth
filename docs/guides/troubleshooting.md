# Troubleshooting Guide

This guide helps troubleshoot common issues with ARauth Identity.

## 🔍 Common Issues

### 1. Database Connection Failed

**Symptoms**:
- Error: "connection refused"
- Error: "authentication failed"

**Solutions**:
```bash
# Check database is running
docker-compose ps postgres-iam

# Check connection string
echo $DATABASE_URL

# Test connection
psql -h localhost -U iam_user -d iam
```

### 2. Redis Connection Failed

**Symptoms**:
- Error: "connection refused"
- Cache not working

**Solutions**:
```bash
# Check Redis is running
docker-compose ps redis

# Test connection
redis-cli -h localhost -p 6379 ping
```

### 3. Hydra Integration Issues

**Symptoms**:
- Error: "login challenge not found"
- Tokens not issued

**Solutions**:
```bash
# Check Hydra is running
docker-compose ps hydra

# Check Hydra admin API
curl http://localhost:4445/health/ready

# Check Hydra logs
docker-compose logs hydra
```

### 4. Performance Issues

**Symptoms**:
- Slow login (> 50ms)
- Slow token issuance (> 10ms)

**Solutions**:
```bash
# Check database performance
EXPLAIN ANALYZE SELECT * FROM users WHERE username = '...';

# Check Redis performance
redis-cli --latency

# Profile application
go tool pprof http://localhost:8080/debug/pprof/profile
```

### 5. Authentication Failures

**Symptoms**:
- Login fails
- Invalid credentials error

**Solutions**:
```bash
# Check user exists
SELECT * FROM users WHERE username = '...';

# Check password hash
SELECT password_hash FROM credentials WHERE user_id = '...';

# Check account status
SELECT status FROM users WHERE id = '...';
```

## 🔧 Debug Mode

### Enable Debug Logging

```bash
export LOG_LEVEL=debug
go run cmd/server/main.go
```

### Enable SQL Logging

```go
// In database config
db.SetLogger(logger)
```

## 📊 Health Checks

### Check All Services

```bash
# IAM API
curl http://localhost:8080/health

# Database
psql -h localhost -U iam_user -d iam -c "SELECT 1"

# Redis
redis-cli ping

# Hydra
curl http://localhost:4445/health/ready
```

## 🐛 Debugging Tips

### 1. Check Logs

```bash
# Application logs
tail -f logs/api.log

# Docker logs
docker-compose logs -f iam-api
```

### 2. Check Metrics

```bash
# Prometheus metrics
curl http://localhost:8080/metrics
```

### 3. Database Queries

```sql
-- Check recent logins
SELECT * FROM audit_logs WHERE action = 'login' ORDER BY created_at DESC LIMIT 10;

-- Check active users
SELECT COUNT(*) FROM users WHERE status = 'active';

-- Check token issues
SELECT COUNT(*) FROM oauth2_access WHERE expires_at > NOW();
```

## 📚 Related Documentation

- [Configuration](../deployment/configuration.md) - Config issues
- [Monitoring](../deployment/monitoring.md) - Monitoring setup

