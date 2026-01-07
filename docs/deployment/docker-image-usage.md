# Using ARauth Identity Docker Image

This guide shows how to use ARauth Identity IAM as a Docker image, similar to using Keycloak or other IAM solutions.

## 🐳 Quick Start

### Pull and Run

```bash
# Pull the image
docker pull arauth-identity/iam-api:latest

# Run with minimal configuration
docker run -d \
  --name nuage-iam \
  -p 8080:8080 \
  -e DATABASE_HOST=postgres \
  -e DATABASE_PORT=5432 \
  -e DATABASE_NAME=iam \
  -e DATABASE_USER=iam_user \
  -e DATABASE_PASSWORD=your_password \
  -e JWT_SECRET=your-jwt-secret \
  -e ENCRYPTION_KEY=your-32-byte-encryption-key \
  arauth-identity/iam-api:latest
```

## 📋 Complete Example

### Step 1: Prepare Environment

```bash
# Create .env file
cat > .env << EOF
DATABASE_PASSWORD=secure_password_123
JWT_SECRET=your-super-secret-jwt-key-change-in-production
ENCRYPTION_KEY=01234567890123456789012345678901
REDIS_PASSWORD=redis_password_123
EOF
```

### Step 2: Run with Docker Compose

```yaml
# docker-compose.yml
version: '3.8'

services:
  postgres:
    image: postgres:14-alpine
    environment:
      POSTGRES_DB: iam
      POSTGRES_USER: iam_user
      POSTGRES_PASSWORD: ${DATABASE_PASSWORD}
    volumes:
      - postgres-data:/var/lib/postgresql/data
    ports:
      - "5432:5432"

  iam-api:
    image: arauth-identity/iam-api:latest
    ports:
      - "8080:8080"
    environment:
      DATABASE_HOST: postgres
      DATABASE_PORT: 5432
      DATABASE_NAME: iam
      DATABASE_USER: iam_user
      DATABASE_PASSWORD: ${DATABASE_PASSWORD}
      JWT_SECRET: ${JWT_SECRET}
      ENCRYPTION_KEY: ${ENCRYPTION_KEY}
      REDIS_HOST: redis
      REDIS_PORT: 6379
      REDIS_PASSWORD: ${REDIS_PASSWORD}
    depends_on:
      - postgres
      - redis
    volumes:
      - ./migrations:/migrations

  redis:
    image: redis:7-alpine
    command: redis-server --requirepass ${REDIS_PASSWORD}
    volumes:
      - redis-data:/data

volumes:
  postgres-data:
  redis-data:
```

### Step 3: Start Services

```bash
# Start all services
docker-compose up -d

# Check logs
docker-compose logs -f iam-api

# Run migrations (if needed)
docker-compose exec iam-api migrate -path /migrations \
  -database "postgres://iam_user:${DATABASE_PASSWORD}@postgres:5432/iam?sslmode=disable" up
```

### Step 4: Verify

```bash
# Health check
curl http://localhost:8080/health

# Expected response:
# {"status":"ok","database":"connected","redis":"connected"}
```

## 🔧 Configuration Options

### Environment Variables

All configuration can be done via environment variables:

```bash
# Server Configuration
SERVER_PORT=8080
SERVER_HOST=0.0.0.0

# Database Configuration
DATABASE_HOST=postgres
DATABASE_PORT=5432
DATABASE_NAME=iam
DATABASE_USER=iam_user
DATABASE_PASSWORD=your_password
DATABASE_SSL_MODE=disable

# Security
JWT_SECRET=your-jwt-secret-key
ENCRYPTION_KEY=your-32-byte-encryption-key

# Redis (Optional)
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=your-redis-password

# Hydra (OAuth2 - Optional)
HYDRA_ADMIN_URL=http://hydra:4445
HYDRA_PUBLIC_URL=http://hydra:4444

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
```

### Using Config File

You can also mount a config file:

```bash
docker run -d \
  --name nuage-iam \
  -p 8080:8080 \
  -v $(pwd)/config.yaml:/config/config.yaml \
  -e CONFIG_PATH=/config/config.yaml \
  arauth-identity/iam-api:latest
```

## 🔌 Integration Examples

### Example 1: Use with Existing Database

```bash
# Connect to your existing PostgreSQL
docker run -d \
  --name nuage-iam \
  -p 8080:8080 \
  -e DATABASE_HOST=your-postgres-host \
  -e DATABASE_PORT=5432 \
  -e DATABASE_NAME=your_database \
  -e DATABASE_USER=your_user \
  -e DATABASE_PASSWORD=your_password \
  -e JWT_SECRET=your-secret \
  arauth-identity/iam-api:latest
```

### Example 2: Use in Docker Network

```bash
# Create network
docker network create iam-network

# Run IAM API
docker run -d \
  --name iam-api \
  --network iam-network \
  -e DATABASE_HOST=postgres \
  -e DATABASE_PORT=5432 \
  arauth-identity/iam-api:latest

# Your app can connect via network
docker run -d \
  --name your-app \
  --network iam-network \
  -e IAM_API_URL=http://iam-api:8080 \
  your-app:latest
```

### Example 3: Behind Reverse Proxy

```yaml
# docker-compose.yml
version: '3.8'

services:
  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
    depends_on:
      - iam-api

  iam-api:
    image: arauth-identity/iam-api:latest
    environment:
      DATABASE_HOST: postgres
      # ... other config
    expose:
      - "8080"
```

## 📦 Building Your Own Image

### Build from Source

```bash
# Clone repository
git clone https://github.com/your-org/arauth-identity.git
cd arauth-identity

# Build image
docker build -t arauth-identity/iam-api:latest .

# Tag for your registry
docker tag arauth-identity/iam-api:latest your-registry/arauth-identity:latest

# Push
docker push your-registry/arauth-identity:latest
```

### Custom Build

```dockerfile
# Custom Dockerfile
FROM arauth-identity/iam-api:latest

# Add custom migrations
COPY custom-migrations /migrations/custom

# Add custom config
COPY custom-config.yaml /config/custom-config.yaml

ENV CONFIG_PATH=/config/custom-config.yaml
```

## 🚀 Production Deployment

### Docker Swarm

```yaml
# docker-stack.yml
version: '3.8'

services:
  iam-api:
    image: arauth-identity/iam-api:latest
    deploy:
      replicas: 3
      update_config:
        parallelism: 1
        delay: 10s
    environment:
      DATABASE_HOST: postgres
      # ... config
    networks:
      - iam-network

networks:
  iam-network:
    driver: overlay
```

### Kubernetes

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: iam-api
spec:
  replicas: 3
  selector:
    matchLabels:
      app: iam-api
  template:
    metadata:
      labels:
        app: iam-api
    spec:
      containers:
      - name: iam-api
        image: arauth-identity/iam-api:latest
        ports:
        - containerPort: 8080
        env:
        - name: DATABASE_HOST
          value: postgres
        - name: DATABASE_PASSWORD
          valueFrom:
            secretKeyRef:
              name: iam-secrets
              key: database-password
```

## 🔍 Troubleshooting

### Check Logs

```bash
# Docker logs
docker logs nuage-iam

# Follow logs
docker logs -f nuage-iam

# Docker Compose logs
docker-compose logs -f iam-api
```

### Health Check

```bash
# Check health endpoint
curl http://localhost:8080/health

# Check readiness
curl http://localhost:8080/health/ready

# Check liveness
curl http://localhost:8080/health/live
```

### Common Issues

1. **Database Connection Failed**
   ```bash
   # Check database is running
   docker ps | grep postgres
   
   # Check connection string
   docker exec nuage-iam env | grep DATABASE
   ```

2. **Port Already in Use**
   ```bash
   # Change port
   docker run -p 8081:8080 arauth-identity/iam-api:latest
   ```

3. **Migration Issues**
   ```bash
   # Run migrations manually
   docker exec nuage-iam migrate -path /migrations \
     -database "postgres://..." up
   ```

## 📚 Next Steps

- [Using as IAM Service](../guides/using-as-iam-service.md)
- [Kubernetes Deployment](./kubernetes.md)
- [API Documentation](../api/README.md)
- [Configuration Guide](./configuration.md)

---

**Last Updated**: 2024

