# Production Deployment Guide

This guide covers deploying ARauth Identity IAM Platform to production environments.

## Prerequisites

- Kubernetes cluster (1.21+)
- kubectl configured
- Helm 3.0+ (for Helm deployment)
- PostgreSQL 12+ database
- Redis 6+ (for caching and rate limiting)
- ORY Hydra (OAuth2/OIDC provider)
- Domain name and SSL certificates
- Monitoring stack (Prometheus, Grafana)

## Architecture Overview

```
┌─────────────┐
│   Ingress   │
└──────┬──────┘
       │
┌──────▼─────────────────┐
│   IAM API (3-10 pods)  │
└──────┬─────────────────┘
       │
   ┌───┴───┬──────────┬──────────┐
   │       │          │          │
┌──▼──┐ ┌─▼───┐  ┌───▼──┐  ┌───▼──┐
│PostgreSQL│ │Redis│  │Hydra│  │Prometheus│
└─────────┘ └────┘  └─────┘  └─────────┘
```

## Deployment Options

### Option 1: Helm Chart (Recommended)

1. **Prepare values file**:

```yaml
# production-values.yaml
replicaCount: 5

image:
  repository: your-registry/arauth-identity/iam-api
  tag: "v0.1.0"

database:
  host: postgres.production.svc.cluster.local
  user: iam_user

redis:
  host: redis.production.svc.cluster.local

ingress:
  enabled: true
  hosts:
    - host: iam.yourdomain.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: iam-tls
      hosts:
        - iam.yourdomain.com

autoscaling:
  enabled: true
  minReplicas: 5
  maxReplicas: 20

resources:
  limits:
    cpu: 1000m
    memory: 1Gi
  requests:
    cpu: 200m
    memory: 512Mi
```

2. **Create secrets** (use external secret management):

```bash
# Using kubectl
kubectl create secret generic arauth-identity-secrets \
  --from-literal=database-password='secure-password' \
  --from-literal=encryption-key='32-byte-encryption-key-here' \
  --from-literal=redis-password='redis-password' \
  -n arauth-identity

# Or use sealed-secrets, vault, etc.
```

3. **Deploy**:

```bash
helm install arauth-identity ./helm/arauth-identity \
  -f production-values.yaml \
  -n arauth-identity \
  --create-namespace
```

### Option 2: Kubernetes Manifests

1. **Create namespace**:

```bash
kubectl apply -f k8s/namespace.yaml
```

2. **Create secrets**:

```bash
cp k8s/secret.yaml.example k8s/secret.yaml
# Edit secret.yaml with your values
kubectl apply -f k8s/secret.yaml
```

3. **Deploy**:

```bash
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
kubectl apply -f k8s/ingress.yaml
kubectl apply -f k8s/hpa.yaml
```

## Database Setup

### PostgreSQL Configuration

1. **Create database**:

```sql
CREATE DATABASE iam_db;
CREATE USER iam_user WITH PASSWORD 'secure-password';
GRANT ALL PRIVILEGES ON DATABASE iam_db TO iam_user;
```

2. **Run migrations**:

```bash
# Using migration tool
migrate -path migrations -database "postgres://iam_user:password@host:5432/iam_db?sslmode=require" up

# Or using kubectl
kubectl run migrate --image=your-registry/arauth-identity/migrate:latest \
  --env="DATABASE_URL=postgres://iam_user:password@postgres:5432/iam_db" \
  --command -- migrate up
```

3. **Configure connection pooling**:

- Set appropriate `max_open_conns` and `max_idle_conns`
- Use connection pooler (PgBouncer) for high traffic

## Redis Setup

1. **Deploy Redis** (if not already deployed):

```bash
# Using Helm
helm install redis bitnami/redis \
  --set auth.password=redis-password \
  --set persistence.enabled=true
```

2. **Configure Redis**:

- Enable persistence for production
- Set appropriate memory limits
- Configure eviction policy

## ORY Hydra Setup

1. **Deploy Hydra** (see Hydra documentation)

2. **Configure Hydra**:

- Set up database for Hydra
- Configure OAuth2 clients
- Set up token signing keys

3. **Update IAM API configuration**:

```yaml
hydra:
  adminURL: "http://hydra:4445"
```

## Security Configuration

### Encryption Key

Generate a secure 32-byte encryption key:

```bash
# Generate random 32-byte key
openssl rand -base64 32
```

Store in Kubernetes secret or external secret management.

### SSL/TLS

1. **Using cert-manager**:

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: iam-tls
spec:
  secretName: iam-tls
  issuerRef:
    name: letsencrypt-prod
    kind: ClusterIssuer
  dnsNames:
    - iam.yourdomain.com
```

2. **Manual certificate**:

```bash
kubectl create secret tls iam-tls \
  --cert=cert.pem \
  --key=key.pem \
  -n arauth-identity
```

## Monitoring Setup

### Prometheus

1. **Configure ServiceMonitor**:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: arauth-identity
spec:
  selector:
    matchLabels:
      app: iam-api
  endpoints:
  - port: http
    path: /metrics
```

2. **Set up alerts** (example):

```yaml
groups:
- name: arauth-identity
  rules:
  - alert: HighErrorRate
    expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.1
    for: 5m
  - alert: DatabaseConnectionFailure
    expr: database_connections_active == 0
    for: 1m
```

### Grafana Dashboards

Import dashboards for:
- HTTP request metrics
- Database performance
- Cache hit rates
- Authentication metrics

## Logging

### Centralized Logging

1. **Configure log aggregation** (ELK, Loki, etc.)

2. **Update logging configuration**:

```yaml
logging:
  level: info
  format: json
  output: stdout
```

3. **Set up log forwarding**:

- Use Fluentd/Fluent Bit
- Configure log shipping to central system

## Backup Strategy

### Database Backups

1. **Automated backups**:

```bash
# Cron job for daily backups
0 2 * * * pg_dump -h postgres -U iam_user iam_db | gzip > backup-$(date +%Y%m%d).sql.gz
```

2. **Point-in-time recovery**:

- Enable WAL archiving
- Configure backup retention policy

### Configuration Backups

- Backup Kubernetes secrets
- Version control configuration files
- Document all configuration changes

## High Availability

### Multi-Region Deployment

1. **Deploy to multiple regions**

2. **Configure database replication**:

- Primary in region 1
- Read replicas in other regions

3. **Load balancing**:

- Use global load balancer
- Route traffic to nearest region

### Disaster Recovery

1. **Backup strategy**:

- Daily database backups
- Off-site backup storage
- Test restore procedures

2. **Recovery procedures**:

- Document recovery steps
- Test recovery regularly
- Maintain runbooks

## Performance Tuning

### Resource Limits

Adjust based on load:

```yaml
resources:
  limits:
    cpu: 2000m
    memory: 2Gi
  requests:
    cpu: 500m
    memory: 1Gi
```

### Auto-scaling

Configure HPA thresholds:

```yaml
autoscaling:
  targetCPUUtilizationPercentage: 70
  targetMemoryUtilizationPercentage: 80
```

### Database Optimization

- Monitor query performance
- Add indexes as needed
- Use connection pooling
- Consider read replicas

## Security Best Practices

1. **Secrets Management**:

- Use external secret management (Vault, Sealed Secrets)
- Rotate secrets regularly
- Never commit secrets to git

2. **Network Policies**:

- Restrict pod-to-pod communication
- Use network policies for isolation

3. **Image Security**:

- Scan container images
- Use private registry
- Sign images

4. **RBAC**:

- Use least privilege principle
- Regular access reviews

## Troubleshooting

### Common Issues

1. **Pods not starting**:

```bash
kubectl describe pod <pod-name> -n arauth-identity
kubectl logs <pod-name> -n arauth-identity
```

2. **Database connection issues**:

- Check network policies
- Verify credentials
- Check database availability

3. **High latency**:

- Check resource limits
- Monitor database performance
- Review cache hit rates

### Health Checks

```bash
# Check health
curl https://iam.yourdomain.com/health

# Check metrics
curl https://iam.yourdomain.com/metrics
```

## Maintenance

### Updates

1. **Rolling updates**:

```bash
helm upgrade arauth-identity ./helm/arauth-identity \
  -f production-values.yaml \
  --set image.tag=v0.2.0
```

2. **Database migrations**:

- Run migrations before deployment
- Test migrations in staging first

### Monitoring

- Monitor error rates
- Track performance metrics
- Review logs regularly
- Set up alerts

## Support

For issues and questions:
- GitHub Issues: https://github.com/arauth-identity/iam/issues
- Documentation: https://docs.arauth-identity.com

