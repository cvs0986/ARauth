# Monitoring & Observability

This document describes monitoring and observability setup for ARauth Identity.

## 🎯 Monitoring Strategy

1. **Metrics**: Prometheus for metrics collection
2. **Logging**: Structured JSON logging
3. **Tracing**: OpenTelemetry for distributed tracing (optional)
4. **Alerting**: Prometheus Alertmanager

## 📊 Metrics

### Application Metrics

**Key Metrics**:

```go
// HTTP metrics
http_requests_total{method, endpoint, status}
http_request_duration_seconds{method, endpoint, status}

// Authentication metrics
auth_login_attempts_total{status}  // success, failure
auth_login_duration_seconds
auth_mfa_attempts_total{status}
auth_token_issued_total{type}  // access, refresh, id

// Database metrics
db_queries_total{table, operation}
db_query_duration_seconds{table, operation}
db_connections_active{pool}
db_connections_idle{pool}

// Cache metrics
cache_operations_total{cache, operation}  // hit, miss
cache_operation_duration_seconds{cache, operation}

// Business metrics
users_total{tenant}
active_users_total{tenant}
roles_total{tenant}
```

### Implementation

```go
import "github.com/prometheus/client_golang/prometheus"

var (
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )
    
    httpRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request duration",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "endpoint", "status"},
    )
)

func init() {
    prometheus.MustRegister(httpRequestsTotal)
    prometheus.MustRegister(httpRequestDuration)
}
```

### Metrics Endpoint

```go
// GET /metrics
http.Handle("/metrics", promhttp.Handler())
```

## 📝 Logging

### Structured Logging

**Format**: JSON

**Fields**:
```json
{
  "timestamp": "2024-01-01T00:00:00Z",
  "level": "info",
  "message": "User logged in",
  "user_id": "user-123",
  "tenant_id": "tenant-123",
  "ip": "192.168.1.1",
  "request_id": "req-123",
  "duration_ms": 45
}
```

### Implementation

```go
import "go.uber.org/zap"

logger, _ := zap.NewProduction()
defer logger.Sync()

logger.Info("user_logged_in",
    zap.String("user_id", userID),
    zap.String("tenant_id", tenantID),
    zap.String("ip", ip),
    zap.String("request_id", requestID),
    zap.Int64("duration_ms", duration),
)
```

### Log Levels

- **DEBUG**: Detailed debugging information
- **INFO**: General information
- **WARN**: Warning messages
- **ERROR**: Error messages

## 🔍 Tracing (Optional)

### OpenTelemetry

```go
import "go.opentelemetry.io/otel"

tracer := otel.Tracer("iam-api")

ctx, span := tracer.Start(ctx, "login")
defer span.End()

span.SetAttributes(
    attribute.String("user_id", userID),
    attribute.String("tenant_id", tenantID),
)
```

## 🚨 Alerting

### Alert Rules

```yaml
groups:
  - name: iam_alerts
    rules:
      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.01
        for: 5m
        annotations:
          summary: "High error rate detected"
          
      - alert: HighLatency
        expr: histogram_quantile(0.95, http_request_duration_seconds) > 0.1
        for: 5m
        annotations:
          summary: "High latency detected"
          
      - alert: DatabaseConnectionPoolExhausted
        expr: db_connections_active / db_connections_max > 0.9
        for: 5m
        annotations:
          summary: "Database connection pool nearly exhausted"
```

## 📈 Dashboards

### Grafana Dashboard

**Key Panels**:

1. **Request Rate**: Requests per second
2. **Error Rate**: Error percentage
3. **Latency**: P50, P95, P99 latency
4. **Authentication**: Login success/failure rate
5. **Database**: Query performance, connection pool
6. **Cache**: Hit/miss ratio
7. **Business**: Active users, tenants

### Example Queries

```promql
# Request rate
rate(http_requests_total[5m])

# Error rate
rate(http_requests_total{status=~"5.."}[5m]) / rate(http_requests_total[5m])

# P95 latency
histogram_quantile(0.95, http_request_duration_seconds)

# Login success rate
rate(auth_login_attempts_total{status="success"}[5m]) / rate(auth_login_attempts_total[5m])
```

## 🔔 Health Checks

### Health Endpoint

```go
// GET /health
{
  "status": "healthy",
  "checks": {
    "database": "up",
    "redis": "up",
    "hydra": "up"
  }
}
```

### Readiness Probe

```go
// GET /ready
// Checks:
// - Database connection
// - Redis connection
// - Hydra connection
```

### Liveness Probe

```go
// GET /live
// Always returns 200 if process is running
```

## 📊 Performance Monitoring

### Key Performance Indicators (KPIs)

| Metric | Target | Alert Threshold |
|--------|--------|-----------------|
| Login P95 Latency | < 50ms | > 100ms |
| Token Issuance P95 | < 10ms | > 20ms |
| Database Query P95 | < 10ms | > 50ms |
| Cache Hit Rate | > 80% | < 70% |
| Error Rate | < 0.1% | > 1% |

## 🔄 Log Aggregation

### Options

1. **ELK Stack**: Elasticsearch, Logstash, Kibana
2. **Loki**: Grafana Loki
3. **Cloud Services**: AWS CloudWatch, Azure Monitor, GCP Logging

### Example: Loki

```yaml
# docker-compose.yml
loki:
  image: grafana/loki:latest
  ports:
    - "3100:3100"
  
promtail:
  image: grafana/promtail:latest
  volumes:
    - ./logs:/var/log/iam
```

## 📚 Related Documentation

- [Kubernetes Deployment](./kubernetes.md) - K8s monitoring
- [Configuration](./configuration.md) - Config for monitoring

