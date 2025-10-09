#!/bin/bash

# Apply CRDs
kubectl apply -f deploy/crds/auto-secret.io_autosecretbasics.yaml
kubectl apply -f deploy/crds/auto-secret.io_autosecretdbs.yaml
kubectl apply -f deploy/crds/auto-secret.io_autosecretguids.yaml

# Apply namespace and RBAC
kubectl apply -f deploy/namespace.yaml
kubectl apply -f deploy/rbac.yaml

# Apply operator deployment
kubectl apply -f deploy/deployment.yaml
