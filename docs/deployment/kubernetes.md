# Kubernetes Deployment Guide

This document describes how to deploy Nuage Identity on Kubernetes.

## 🎯 Prerequisites

- Kubernetes cluster (v1.25+)
- Helm 3.0+
- kubectl configured
- PostgreSQL database
- Redis instance
- ORY Hydra instance

## 📦 Helm Chart Structure

```
helm/
├── Chart.yaml
├── values.yaml
├── values.prod.yaml
├── values.dev.yaml
├── templates/
│   ├── deployment.yaml
│   ├── service.yaml
│   ├── configmap.yaml
│   ├── secret.yaml
│   ├── hpa.yaml
│   └── ingress.yaml
└── README.md
```

## 🔧 Configuration

### values.yaml

```yaml
replicaCount: 2

image:
  repository: nuage-identity
  tag: latest
  pullPolicy: IfNotPresent

service:
  type: ClusterIP
  port: 8080

ingress:
  enabled: true
  className: nginx
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
  hosts:
    - host: iam.example.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: iam-tls
      hosts:
        - iam.example.com

config:
  server:
    port: 8080
    host: "0.0.0.0"
  database:
    host: postgres-service
    port: 5432
    name: iam
    user: iam_user
  redis:
    host: redis-service
    port: 6379
  hydra:
    adminURL: http://hydra-admin:4445
    publicURL: http://hydra-public:4444

resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 500m
    memory: 256Mi

autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70
  targetMemoryUtilizationPercentage: 80

securityContext:
  runAsNonRoot: true
  runAsUser: 1000
  fsGroup: 1000
```

## 🚀 Deployment Steps

### 1. Create Namespace

```bash
kubectl create namespace iam
```

### 2. Create Secrets

```bash
kubectl create secret generic iam-secrets \
  --from-literal=database-password='<password>' \
  --from-literal=redis-password='<password>' \
  --from-literal=jwt-secret='<secret>' \
  --namespace=iam
```

### 3. Install Helm Chart

```bash
helm install iam ./helm \
  --namespace iam \
  --values helm/values.prod.yaml
```

### 4. Verify Deployment

```bash
kubectl get pods -n iam
kubectl get services -n iam
kubectl logs -f deployment/iam-api -n iam
```

## 📊 Horizontal Pod Autoscaling

### HPA Configuration

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: iam-api-hpa
  namespace: iam
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

## 🔐 Security

### Pod Security

```yaml
securityContext:
  runAsNonRoot: true
  runAsUser: 1000
  fsGroup: 1000
  capabilities:
    drop:
      - ALL
```

### Network Policies

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: iam-api-policy
  namespace: iam
spec:
  podSelector:
    matchLabels:
      app: iam-api
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
    ports:
    - protocol: TCP
      port: 8080
  egress:
  - to:
    - podSelector:
        matchLabels:
          app: postgres
    ports:
    - protocol: TCP
      port: 5432
  - to:
    - podSelector:
        matchLabels:
          app: redis
    ports:
    - protocol: TCP
      port: 6379
```

## 📈 Monitoring

### ServiceMonitor (Prometheus)

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: iam-api
  namespace: iam
spec:
  selector:
    matchLabels:
      app: iam-api
  endpoints:
  - port: http
    path: /metrics
    interval: 30s
```

## 🔄 Updates and Rollbacks

### Update Deployment

```bash
helm upgrade iam ./helm \
  --namespace iam \
  --values helm/values.prod.yaml
```

### Rollback

```bash
helm rollback iam --namespace iam
```

## 📚 Related Documentation

- [Docker Compose](./docker-compose.md) - Local development
- [Configuration](./configuration.md) - Configuration management
- [Monitoring](./monitoring.md) - Monitoring setup

