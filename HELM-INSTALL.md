# Auto Secret Operator - Helm Chart

This Helm chart deploys the Auto Secret Operator to your Kubernetes cluster.

## Prerequisites

- Kubernetes 1.19+
- Helm 3.0+

## Installation

### Quick Install

```bash
# Run the setup script to create the Helm chart structure (first time only)
./setup-helm-chart.ps1

# Install the operator
helm install auto-secret-operator ./helm/auto-secret-operator
```

### Custom Installation

```bash
# Install with custom values
helm install auto-secret-operator ./helm/auto-secret-operator \
  --set image.tag=0.1.12 \
  --set namespace=my-namespace \
  --set resources.limits.memory=256Mi
```

### Install from specific namespace

```bash
helm install auto-secret-operator ./helm/auto-secret-operator \
  --create-namespace \
  --namespace auto-secret-operator
```

## Configuration

The following table lists the configurable parameters and their default values.

| Parameter | Description | Default |
|-----------|-------------|---------|
| `replicaCount` | Number of replicas | `1` |
| `image.repository` | Image repository | `sindrema/auto-secret-operator` |
| `image.tag` | Image tag | `0.1.12` (from Chart.appVersion) |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `namespace` | Namespace to deploy operator | `auto-secret-operator` |
| `serviceAccount.create` | Create service account | `true` |
| `rbac.create` | Create RBAC resources | `true` |
| `operator.leaderElect` | Enable leader election | `true` |
| `resources.limits.cpu` | CPU limit | `500m` |
| `resources.limits.memory` | Memory limit | `128Mi` |
| `resources.requests.cpu` | CPU request | `10m` |
| `resources.requests.memory` | Memory request | `64Mi` |

## Uninstallation

```bash
helm uninstall auto-secret-operator
```

To also remove the CRDs:

```bash
kubectl delete crd autosecretbasics.auto-secret.io
kubectl delete crd autosecretdbs.auto-secret.io
kubectl delete crd autosecretguids.auto-secret.io
kubectl delete crd autosecretdbsecretredirects.auto-secret.io
```

## Upgrading

```bash
helm upgrade auto-secret-operator ./helm/auto-secret-operator
```

## Usage Examples

See the main [README.md](../README.md) for detailed usage examples.
