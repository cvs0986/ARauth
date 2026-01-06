# Scalability Design

This document describes the scalability architecture and design decisions for Nuage Identity.

## 🎯 Scalability Goals

| Metric | Target | Measurement |
|--------|--------|-------------|
| Startup Time | < 300ms | Time from container start to ready |
| Login Latency | < 50ms | P95 latency for login endpoint |
| Token Issuance | < 10ms | P95 latency for token issuance |
| Memory Usage | < 150MB | Per instance memory footprint |
| Concurrent Logins | 10k+ | Simultaneous login requests |
| Throughput | 10k+ req/s | Requests per second per instance |

## 🏗️ Scalability Architecture

### Horizontal Scaling

```
                    ┌─────────────┐
                    │ Load Balancer│
                    └──────┬───────┘
                           │
        ┌──────────────────┼──────────────────┐
        │                  │                  │
   ┌────▼────┐        ┌────▼────┐       ┌────▼────┐
   │ IAM API │        │ IAM API │       │ IAM API │
   │ Pod 1   │        │ Pod 2   │       │ Pod N   │
   └────┬────┘        └────┬────┘       └────┬────┘
        │                  │                  │
        └──────────────────┼──────────────────┘
                           │
        ┌──────────────────┼──────────────────┐
        │                  │                  │
   ┌────▼────┐        ┌────▼────┐       ┌────▼────┐
   │PostgreSQL│       │  Redis  │       │  Hydra  │
   │ (IAM DB) │       │ Cluster │       │         │
   └──────────┘       └─────────┘       └─────────┘
```

### Stateless Design

**Key Principle**: All IAM API instances are identical and stateless.

- No in-memory state
- No server-side sessions
- All state in external storage (DB, Redis)

**Benefits**:
- Easy horizontal scaling
- No session affinity required
- Simple load balancing

## 💾 Database Scalability

### Connection Pooling

```go
// PostgreSQL connection pool
db, err := sql.Open("postgres", dsn)
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)
```

**Configuration**:
- Max open connections: 25 per instance
- Max idle connections: 5 per instance
- Connection lifetime: 5 minutes

### Read Replicas

```
Write → Primary PostgreSQL
Read  → Read Replica (optional)
```

**Use Cases**:
- User lookups (read-heavy)
- Role/permission queries (read-heavy)
- Tenant queries (read-heavy)

**Implementation**:

```go
type Repository struct {
    writeDB *sql.DB  // Primary
    readDB  *sql.DB  // Read replica (optional)
}

func (r *Repository) GetUser(ctx context.Context, id string) (*User, error) {
    db := r.readDB
    if db == nil {
        db = r.writeDB
    }
    // Query read replica
}
```

### Database Indexing

**Critical Indexes**:

```sql
-- Users
CREATE INDEX idx_users_username_tenant ON users(username, tenant_id);
CREATE INDEX idx_users_email_tenant ON users(email, tenant_id);
CREATE INDEX idx_users_tenant_id ON users(tenant_id);

-- Credentials
CREATE INDEX idx_credentials_user_id ON credentials(user_id);

-- Roles
CREATE INDEX idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX idx_role_permissions_role_id ON role_permissions(role_id);
```

## 🚀 Caching Strategy

### Multi-Level Caching

```
Request
  ↓
L1: In-Memory Cache (5 min TTL)
  ↓ (miss)
L2: Redis Cache (10 min TTL)
  ↓ (miss)
Database
```

### Cache Layers

#### 1. In-Memory Cache (L1)

**Use Cases**:
- Tenant data (rarely changes)
- Role definitions (rarely changes)
- Permission mappings (rarely changes)

**Implementation**:

```go
type InMemoryCache struct {
    cache *sync.Map
    ttl   time.Duration
}

func (c *InMemoryCache) Get(key string) (interface{}, bool) {
    item, ok := c.cache.Load(key)
    if !ok {
        return nil, false
    }
    
    cached := item.(*CacheItem)
    if time.Since(cached.Timestamp) > c.ttl {
        c.cache.Delete(key)
        return nil, false
    }
    
    return cached.Value, true
}
```

**TTL**: 5-10 minutes

#### 2. Redis Cache (L2)

**Use Cases**:
- User data (5 min TTL)
- MFA sessions (5 min TTL)
- Rate limiting counters (1 min TTL)
- Refresh token blacklist (token expiry TTL)

**Implementation**:

```go
type RedisCache struct {
    client *redis.Client
}

func (c *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
    return c.client.Get(ctx, key).Bytes()
}

func (c *RedisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
    return c.client.Set(ctx, key, value, ttl).Err()
}
```

**TTL Strategy**:
- User data: 5 minutes
- Tenant data: 10 minutes
- Role/permissions: 15 minutes
- MFA sessions: 5 minutes

### Cache Invalidation

**Strategies**:

1. **TTL-based**: Automatic expiration
2. **Event-based**: Invalidate on updates
3. **Version-based**: Cache version numbers

```go
func (s *UserService) UpdateUser(ctx context.Context, user *User) error {
    // Update database
    err := s.repo.Update(ctx, user)
    if err != nil {
        return err
    }
    
    // Invalidate cache
    cacheKey := fmt.Sprintf("user:%s", user.ID)
    s.cache.Delete(ctx, cacheKey)
    
    return nil
}
```

## ⚡ Performance Optimizations

### 1. Database Query Optimization

**Batch Operations**:

```go
// Bad: N+1 queries
for _, userID := range userIDs {
    roles, _ := repo.GetUserRoles(userID)
}

// Good: Single query
roles, _ := repo.GetUserRolesBatch(userIDs)
```

**Eager Loading**:

```go
// Load user with roles and permissions in one query
user, err := repo.GetUserWithRelations(ctx, userID)
```

### 2. Parallel Processing

```go
func (s *AuthService) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
    // Parallel execution
    var user *User
    var tenant *Tenant
    var err error
    
    var wg sync.WaitGroup
    wg.Add(2)
    
    go func() {
        defer wg.Done()
        user, err = s.userRepo.GetByUsername(ctx, req.Username, req.TenantID)
    }()
    
    go func() {
        defer wg.Done()
        tenant, err = s.tenantRepo.GetByID(ctx, req.TenantID)
    }()
    
    wg.Wait()
    // Continue with login...
}
```

### 3. Connection Pooling

**HTTP Client Pooling**:

```go
httpClient := &http.Client{
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
    },
    Timeout: 10 * time.Second,
}
```

### 4. Async Operations

**Non-blocking Operations**:

```go
// Async audit logging
go func() {
    s.auditLogger.Log(ctx, &AuditEvent{
        Action: "login",
        UserID: user.ID,
    })
}()

// Continue with response
return loginResponse
```

## 📊 Rate Limiting

### Distributed Rate Limiting

**Redis-based Sliding Window**:

```go
func (rl *RateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
    now := time.Now()
    windowStart := now.Add(-window)
    
    // Remove old entries
    rl.redis.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStart.Unix(), 10))
    
    // Count current requests
    count, err := rl.redis.ZCard(ctx, key).Result()
    if err != nil {
        return false, err
    }
    
    if count >= int64(limit) {
        return false, nil
    }
    
    // Add current request
    rl.redis.ZAdd(ctx, key, &redis.Z{
        Score:  float64(now.Unix()),
        Member: strconv.FormatInt(now.UnixNano(), 10),
    })
    rl.redis.Expire(ctx, key, window)
    
    return true, nil
}
```

**Rate Limits**:
- Login: 5 attempts per minute per IP
- Token refresh: 10 requests per minute per token
- API calls: 100 requests per minute per client

## 🔄 Load Balancing

### Strategies

1. **Round Robin**: Default, equal distribution
2. **Least Connections**: Better for long-lived connections
3. **IP Hash**: Session affinity (not needed for stateless)

### Health Checks

```go
func (h *HealthHandler) Check(ctx *gin.Context) {
    health := &HealthStatus{
        Status: "healthy",
        Checks: map[string]string{},
    }
    
    // Check database
    if err := h.db.Ping(); err != nil {
        health.Status = "unhealthy"
        health.Checks["database"] = "down"
    } else {
        health.Checks["database"] = "up"
    }
    
    // Check Redis
    if err := h.redis.Ping(ctx).Err(); err != nil {
        health.Status = "unhealthy"
        health.Checks["redis"] = "down"
    } else {
        health.Checks["redis"] = "up"
    }
    
    ctx.JSON(200, health)
}
```

## 📈 Monitoring & Metrics

### Key Metrics

```go
// Request latency
metrics.Histogram("http_request_duration_seconds",
    []string{"method", "endpoint", "status"})

// Request rate
metrics.Counter("http_requests_total",
    []string{"method", "endpoint", "status"})

// Database query latency
metrics.Histogram("db_query_duration_seconds",
    []string{"query", "table"})

// Cache hit rate
metrics.Counter("cache_hits_total", []string{"cache", "key"})
metrics.Counter("cache_misses_total", []string{"cache", "key"})

// Active connections
metrics.Gauge("db_connections_active", []string{"pool"})
```

### Performance Targets

| Metric | Target | Alert Threshold |
|--------|--------|-----------------|
| Login P95 Latency | < 50ms | > 100ms |
| Token Issuance P95 | < 10ms | > 20ms |
| Database Query P95 | < 10ms | > 50ms |
| Cache Hit Rate | > 80% | < 70% |
| Error Rate | < 0.1% | > 1% |

## 🚀 Kubernetes Scaling

### Horizontal Pod Autoscaling (HPA)

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: iam-api-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: iam-api
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

### Resource Limits

```yaml
resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 500m
    memory: 256Mi
```

## 📚 Related Documentation

- [Architecture Overview](./overview.md)
- [Components](./components.md)
- [Deployment Guide](../deployment/kubernetes.md)

