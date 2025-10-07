# DB Secret Operator

A Kubernetes operator for automatically generating database credentials and connection URIs for CloudNativePG (CNPG) clusters.

## Features

- **Automatic Password Generation**: Creates cryptographically secure passwords for database users
- **Connection URI Generation**: Automatically generates PostgreSQL connection strings from user credentials
- **Secret Watching**: Only reconciles secrets with the operator's annotations (efficient)
- **URL Encoding**: Properly encodes special characters in passwords for URIs
- **CloudNativePG Integration**: Works seamlessly with CNPG naming conventions

## How It Works

The operator watches for Kubernetes Secrets with specific annotations and performs two types of operations:

### 1. DB User Secrets (`db-user`)

Creates database user credentials with auto-generated passwords.

**Annotations Required:**
- `cnpg-secret-generator.io/type: "db-user"`
- `cnpg-secret-generator.io/username: "<username>"`
- `cnpg-secret-generator.io/cluster: "<cluster-name>"`

**Generated Data:**
- `username`: The specified username
- `password`: A cryptographically secure 32-character password

### 2. DB URI Secrets (`db-uri`)

Generates PostgreSQL connection URIs from user credentials.

**Annotations Required:**
- `cnpg-secret-generator.io/type: "db-uri"`
- `cnpg-secret-generator.io/source-secret: "<source-secret-name>"`
- `cnpg-secret-generator.io/cluster: "<cluster-name>"`
- `cnpg-secret-generator.io/dbname: "<database-name>"`

**Generated Data:**
- `DATABASE_URI`: Full PostgreSQL connection string (format: `postgresql://user:pass@host:5432/dbname`)

## Installation

### Prerequisites

- Kubernetes cluster (v1.24+)
- kubectl configured to access your cluster
- Docker (for building custom images)

### Deploy the Operator

```bash
# Clone the repository
git clone https://github.com/yourusername/db-secret-operator
cd db-secret-operator

# Download dependencies
make mod-download

# Build and push the image (optional - customize registry)
make docker-build IMG=db-secret-operator:latest
# make docker-push REGISTRY=your-registry.com

# Deploy to cluster
make deploy
```

### Verify Installation

```bash
# Check operator is running
kubectl get pods -n db-secret-operator-system

# View logs
make logs
```

## Usage

### Example 1: Basic DB User Secret

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: myapp-db-user
  namespace: default
  annotations:
    cnpg-secret-generator.io/type: "db-user"
    cnpg-secret-generator.io/username: "myapp_user"
    cnpg-secret-generator.io/cluster: "postgres-cluster"
type: Opaque
data: {}
```

Apply this secret:
```bash
kubectl apply -f examples/db-user-secret.yaml
```

The operator will populate `username` and `password` fields automatically.

### Example 2: DB URI Secret

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: myapp-db-uri
  namespace: default
  annotations:
    cnpg-secret-generator.io/type: "db-uri"
    cnpg-secret-generator.io/source-secret: "myapp-db-user"
    cnpg-secret-generator.io/cluster: "postgres-cluster"
    cnpg-secret-generator.io/dbname: "myapp_db"
type: Opaque
data: {}
```

Apply this secret:
```bash
kubectl apply -f examples/db-uri-secret.yaml
```

The operator will generate a connection URI using credentials from `myapp-db-user`.

### Example 3: Using Secrets in Deployments

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
spec:
  template:
    spec:
      containers:
      - name: app
        image: myapp:latest
        env:
        # Use the connection URI
        - name: DATABASE_URI
          valueFrom:
            secretKeyRef:
              name: myapp-db-uri
              key: DATABASE_URI

        # Or use individual credentials
        - name: DB_USERNAME
          valueFrom:
            secretKeyRef:
              name: myapp-db-user
              key: username
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: myapp-db-user
              key: password
```

See [examples/complete-example.yaml](examples/complete-example.yaml) for a complete working example.

## Development

### Build Locally

```bash
# Format code
make fmt

# Run linters
make vet

# Run tests
make test

# Build binary
make build

# Run locally (requires kubeconfig)
make run
```

### Project Structure

```
db-secret-operator/
├── controllers/
│   └── secret_controller.go  # Main reconciliation logic
├── deploy/
│   ├── namespace.yaml        # Operator namespace
│   ├── rbac.yaml            # RBAC permissions
│   └── deployment.yaml      # Operator deployment
├── examples/
│   ├── db-user-secret.yaml
│   ├── db-uri-secret.yaml
│   └── complete-example.yaml
├── main.go                   # Operator entry point
├── Dockerfile               # Container image
├── Makefile                # Build automation
└── README.md
```

## Architecture

The operator uses the [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime) framework to:

1. Watch for Secret resources with specific annotations
2. Generate passwords using `crypto/rand` for security
3. Update secrets with generated data
4. Requeue if dependent secrets are not ready

### Security Features

- **Crypto-secure passwords**: Uses `crypto/rand` instead of `math/rand`
- **URL encoding**: Properly escapes special characters in connection URIs
- **Non-root containers**: Runs as user 65532 with dropped capabilities
- **Read-only filesystem**: Uses distroless base image
- **Minimal permissions**: Only requires get/list/watch/update on secrets

## Troubleshooting

### Secret not being updated

1. Check the secret has the required annotations:
   ```bash
   kubectl get secret myapp-db-user -o yaml
   ```

2. Verify the operator is running:
   ```bash
   kubectl get pods -n db-secret-operator-system
   ```

3. Check operator logs:
   ```bash
   make logs
   ```

### URI secret waiting for source secret

The `db-uri` secret will requeue every 5 seconds until the source secret exists and has data. Check:

```bash
# Verify source secret exists
kubectl get secret myapp-db-user

# Verify it has data
kubectl get secret myapp-db-user -o jsonpath='{.data}'
```

## Limitations

- Only supports PostgreSQL (port 5432 hardcoded)
- Uses CNPG naming convention (`<cluster>-rw` for read-write service)
- Does not create actual database users (credentials only)
- Password rotation requires manual secret deletion/recreation

## Roadmap

- [ ] Add support for creating actual PostgreSQL users via SQL
- [ ] Support custom port configuration
- [ ] Add password rotation scheduling
- [ ] Support for additional database types (MySQL, MongoDB)
- [ ] Metrics and monitoring integration
- [ ] Webhook for validation

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details

## Acknowledgments

Inspired by [kubernetes-secret-generator](https://github.com/mittwald/kubernetes-secret-generator)
