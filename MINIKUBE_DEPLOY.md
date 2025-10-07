# Deploy to Minikube

This guide walks you through deploying the db-secret-operator to minikube.

## Prerequisites

1. Install minikube: https://minikube.sigs.k8s.io/docs/start/
2. Start minikube:
   ```bash
   minikube start
   ```

## Method 1: Build Inside Minikube (Recommended)

This method builds the Docker image directly inside minikube's Docker daemon, avoiding the need to push to a registry.

### Step 1: Point Docker to Minikube's Docker Daemon

**On Linux/macOS:**
```bash
eval $(minikube docker-env)
```

**On Windows (PowerShell):**
```powershell
& minikube -p minikube docker-env --shell powershell | Invoke-Expression
```

**On Windows (CMD):**
```cmd
@FOR /f "tokens=*" %i IN ('minikube -p minikube docker-env --shell cmd') DO @%i
```

### Step 2: Build the Docker Image

```bash
docker build -t db-secret-operator:latest .
```

### Step 3: Deploy to Kubernetes

```bash
# Apply all manifests
kubectl apply -f deploy/namespace.yaml
kubectl apply -f deploy/rbac.yaml
kubectl apply -f deploy/deployment.yaml
```

### Step 4: Verify Deployment

```bash
# Check if the operator pod is running
kubectl get pods -n db-secret-operator-system

# View logs
kubectl logs -n db-secret-operator-system -l app=db-secret-operator -f
```

## Method 2: Load Pre-built Image

If you've already built the image outside minikube:

```bash
# Build the image (if not already built)
docker build -t db-secret-operator:latest .

# Load into minikube
minikube image load db-secret-operator:latest

# Deploy
kubectl apply -f deploy/namespace.yaml
kubectl apply -f deploy/rbac.yaml
kubectl apply -f deploy/deployment.yaml
```

## Quick Deploy Script

Create a file `deploy-minikube.sh` (or `deploy-minikube.ps1` on Windows):

**Linux/macOS (deploy-minikube.sh):**
```bash
#!/bin/bash
set -e

echo "Setting up minikube Docker environment..."
eval $(minikube docker-env)

echo "Building operator image..."
docker build -t db-secret-operator:latest .

echo "Deploying to Kubernetes..."
kubectl apply -f deploy/namespace.yaml
kubectl apply -f deploy/rbac.yaml
kubectl apply -f deploy/deployment.yaml

echo "Waiting for operator to be ready..."
kubectl wait --for=condition=available --timeout=60s \
  deployment/db-secret-operator -n db-secret-operator-system

echo "Operator deployed successfully!"
kubectl get pods -n db-secret-operator-system
```

**Windows PowerShell (deploy-minikube.ps1):**
```powershell
Write-Host "Setting up minikube Docker environment..."
& minikube -p minikube docker-env --shell powershell | Invoke-Expression

Write-Host "Building operator image..."
docker build -t db-secret-operator:latest .

Write-Host "Deploying to Kubernetes..."
kubectl apply -f deploy/namespace.yaml
kubectl apply -f deploy/rbac.yaml
kubectl apply -f deploy/deployment.yaml

Write-Host "Waiting for operator to be ready..."
kubectl wait --for=condition=available --timeout=60s `
  deployment/db-secret-operator -n db-secret-operator-system

Write-Host "Operator deployed successfully!"
kubectl get pods -n db-secret-operator-system
```

Make the script executable (Linux/macOS):
```bash
chmod +x deploy-minikube.sh
./deploy-minikube.sh
```

Run on Windows:
```powershell
.\deploy-minikube.ps1
```

## Testing the Operator

### 1. Deploy Example Secrets

```bash
kubectl apply -f examples/db-user-secret.yaml
kubectl apply -f examples/db-uri-secret.yaml
```

### 2. Check Secrets Were Generated

```bash
# Check db-user secret
kubectl get secret myapp-db-user -o yaml

# View generated password
kubectl get secret myapp-db-user -o jsonpath='{.data.password}' | base64 -d
echo

# Check db-uri secret
kubectl get secret myapp-db-uri -o yaml

# View generated URI
kubectl get secret myapp-db-uri -o jsonpath='{.data.DATABASE_URI}' | base64 -d
echo
```

### 3. Monitor Operator Logs

```bash
kubectl logs -n db-secret-operator-system -l app=db-secret-operator -f
```

## Troubleshooting

### Pod Not Starting

```bash
# Describe the pod to see errors
kubectl describe pod -n db-secret-operator-system -l app=db-secret-operator

# Check events
kubectl get events -n db-secret-operator-system --sort-by='.lastTimestamp'
```

### Image Pull Errors

If you see `ImagePullBackOff` errors:

1. Ensure you're using minikube's Docker daemon:
   ```bash
   eval $(minikube docker-env)
   ```

2. Verify the image exists:
   ```bash
   docker images | grep db-secret-operator
   ```

3. Make sure `imagePullPolicy` is set to `IfNotPresent` in [deploy/deployment.yaml](deploy/deployment.yaml) (already configured)

### Secrets Not Being Updated

1. Check operator logs for errors:
   ```bash
   kubectl logs -n db-secret-operator-system -l app=db-secret-operator
   ```

2. Verify RBAC permissions:
   ```bash
   kubectl get clusterrole db-secret-operator-role -o yaml
   ```

3. Ensure secrets have correct annotations:
   ```bash
   kubectl get secret myapp-db-user -o jsonpath='{.metadata.annotations}'
   ```

## Uninstall

```bash
# Remove example secrets
kubectl delete -f examples/db-user-secret.yaml
kubectl delete -f examples/db-uri-secret.yaml

# Remove operator
kubectl delete -f deploy/deployment.yaml
kubectl delete -f deploy/rbac.yaml
kubectl delete -f deploy/namespace.yaml
```

Or use the Makefile:
```bash
make undeploy
```

## Makefile Commands

The project includes a Makefile with helpful commands:

```bash
# Build and deploy in one command (after setting minikube docker-env)
make docker-build deploy

# View logs
make logs

# Undeploy
make undeploy

# Deploy examples
make deploy-examples
```

## Next Steps

- Set up a PostgreSQL cluster using CloudNativePG
- Create database users with the generated credentials
- Connect your applications using the generated URIs
