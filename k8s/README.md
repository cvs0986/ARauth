# Kubernetes Deployment

This directory contains Kubernetes manifests for deploying Nuage Identity IAM API.

## Prerequisites

- Kubernetes cluster (1.21+)
- kubectl configured
- PostgreSQL database (can be deployed separately)
- Redis (can be deployed separately)
- ORY Hydra (can be deployed separately)

## Deployment Steps

### 1. Create Namespace

```bash
kubectl apply -f namespace.yaml
```

### 2. Create Secrets

First, copy the example secret file and update it with your actual values:

```bash
cp secret.yaml.example secret.yaml
# Edit secret.yaml with your actual secrets
kubectl apply -f secret.yaml
```

**Important**: The `encryption-key` must be exactly 32 bytes for AES-256 encryption.

### 3. Create ConfigMap

```bash
kubectl apply -f configmap.yaml
```

### 4. Deploy Application

```bash
kubectl apply -f deployment.yaml
kubectl apply -f service.yaml
```

### 5. (Optional) Configure Ingress

If you want external access:

```bash
kubectl apply -f ingress.yaml
```

### 6. (Optional) Configure Horizontal Pod Autoscaler

For automatic scaling:

```bash
kubectl apply -f hpa.yaml
```

## Verify Deployment

```bash
# Check pods
kubectl get pods -n nuage-identity

# Check services
kubectl get svc -n nuage-identity

# Check deployment status
kubectl get deployment -n nuage-identity

# View logs
kubectl logs -f deployment/iam-api -n nuage-identity
```

## Health Checks

The deployment includes liveness and readiness probes that check the `/health` endpoint.

## Scaling

The HPA will automatically scale the deployment based on CPU and memory usage:
- Minimum replicas: 3
- Maximum replicas: 10
- CPU target: 70%
- Memory target: 80%

## Configuration

Configuration is managed through:
- ConfigMap: `iam-config` (non-sensitive settings)
- Secret: `iam-secrets` (sensitive data like passwords)

## Troubleshooting

### Pods not starting

```bash
# Check pod status
kubectl describe pod <pod-name> -n nuage-identity

# Check logs
kubectl logs <pod-name> -n nuage-identity
```

### Database connection issues

Verify that:
1. PostgreSQL is accessible from the cluster
2. Database credentials in secret are correct
3. Network policies allow connection

### Redis connection issues

Verify that:
1. Redis is accessible from the cluster
2. Redis password in secret is correct (if required)

## Production Considerations

1. **Secrets Management**: Use a proper secrets management system (e.g., Vault, Sealed Secrets)
2. **Image Security**: Use private container registry and scan images
3. **Network Policies**: Implement network policies for pod-to-pod communication
4. **Resource Limits**: Adjust resource requests/limits based on your workload
5. **Monitoring**: Set up Prometheus and Grafana for monitoring
6. **Logging**: Configure centralized logging (e.g., ELK stack)
7. **Backup**: Set up database backups
8. **SSL/TLS**: Use proper certificates for ingress

