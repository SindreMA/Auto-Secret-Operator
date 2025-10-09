
kubectl apply -f deploy/namespace.yaml
kubectl apply -f deploy/rbac.yaml
kubectl apply -f deploy/crds/auto-secret.io_autosecrets.yaml
kubectl apply -f deploy/deployment.yaml