# Nuage Identity Helm Chart

This Helm chart deploys the Nuage Identity IAM Platform on a Kubernetes cluster.

## Prerequisites

- Kubernetes 1.21+
- Helm 3.0+
- PostgreSQL database (can be deployed separately)
- Redis (can be deployed separately)
- ORY Hydra (can be deployed separately)

## Installation

### Quick Start

```bash
# Add the repository (if using a chart repository)
helm repo add nuage-identity https://charts.nuage-identity.com
helm repo update

# Install with default values
helm install nuage-identity nuage-identity/nuage-identity

# Or install from local chart
helm install nuage-identity ./helm/nuage-identity
```

### Custom Installation

1. **Create a values file** with your configuration:

```yaml
# my-values.yaml
replicaCount: 5

database:
  host: postgres.example.com
  user: iam_user

redis:
  host: redis.example.com

secrets:
  databasePassword: "your-secure-password"
  encryptionKey: "your-32-byte-encryption-key"
```

2. **Install with custom values**:

```bash
helm install nuage-identity ./helm/nuage-identity -f my-values.yaml
```

### Using Secrets from External Secret Management

For production, use external secret management (e.g., Vault, Sealed Secrets):

```bash
# Create secrets externally, then reference them
helm install nuage-identity ./helm/nuage-identity \
  --set secrets.databasePassword=$(vault kv get -field=password secret/iam/db) \
  --set secrets.encryptionKey=$(vault kv get -field=key secret/iam/encryption)
```

## Configuration

### Required Values

- `secrets.databasePassword`: Database password
- `secrets.encryptionKey`: 32-byte encryption key for AES-256

### Important Configuration Options

| Parameter | Description | Default |
|-----------|-------------|---------|
| `replicaCount` | Number of replicas | `3` |
| `image.repository` | Container image repository | `nuage-identity/iam-api` |
| `image.tag` | Container image tag | `latest` |
| `database.host` | PostgreSQL host | `postgres-iam` |
| `database.port` | PostgreSQL port | `5432` |
| `redis.host` | Redis host | `redis` |
| `redis.port` | Redis port | `6379` |
| `hydra.adminURL` | ORY Hydra admin URL | `http://hydra:4445` |
| `autoscaling.enabled` | Enable HPA | `true` |
| `autoscaling.minReplicas` | Minimum replicas | `3` |
| `autoscaling.maxReplicas` | Maximum replicas | `10` |
| `ingress.enabled` | Enable ingress | `false` |

See `values.yaml` for all available options.

## Upgrading

```bash
# Upgrade with new values
helm upgrade nuage-identity ./helm/nuage-identity -f my-values.yaml

# Upgrade to a specific version
helm upgrade nuage-identity ./helm/nuage-identity --version 0.1.0
```

## Uninstalling

```bash
helm uninstall nuage-identity
```

## Health Checks

The chart includes:
- **Liveness probe**: `/health/live`
- **Readiness probe**: `/health/ready`
- **Health check**: `/health`

## Monitoring

The application exposes Prometheus metrics at `/metrics`. Configure Prometheus to scrape this endpoint.

## Troubleshooting

### Check Pod Status

```bash
kubectl get pods -l app.kubernetes.io/name=nuage-identity
```

### View Logs

```bash
kubectl logs -l app.kubernetes.io/name=nuage-identity
```

### Check Configuration

```bash
kubectl get configmap nuage-identity-config -o yaml
kubectl get secret nuage-identity-secrets -o yaml
```

### Test Health Endpoints

```bash
kubectl port-forward svc/nuage-identity 8080:80
curl http://localhost:8080/health
```

## Production Considerations

1. **Secrets Management**: Use external secret management (Vault, Sealed Secrets, etc.)
2. **Image Security**: Use private registry and scan images
3. **Resource Limits**: Adjust based on workload
4. **Network Policies**: Implement pod-to-pod communication policies
5. **Backup**: Set up database backups
6. **Monitoring**: Configure Prometheus and Grafana
7. **Logging**: Set up centralized logging (ELK stack)
8. **SSL/TLS**: Use proper certificates for ingress

## Support

For issues and questions, please open an issue on GitHub:
https://github.com/nuage-identity/iam/issues

