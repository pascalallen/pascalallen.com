# Deploying

This document describes the steps necessary for deploying to a Kubernetes cluster.

## Prerequisites

- [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl)
- [A provisioned Kubernetes cluster that you can connect to](https://kubernetes.io/docs/home/#set-up-a-k8s-cluster)

### Create Kubernetes Secret from `.env` file

```bash
kubectl create secret generic env-vars --from-env-file=.env
```

### Apply Services to Kubernetes Cluster

```bash
# Install Nginx Ingress Controller
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.10.0/deploy/static/provider/baremetal/deploy.yaml

# Install Cert-Manager
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.14.4/cert-manager.yaml

# Wait for Cert-Manager to be ready
kubectl wait --for=condition=Ready pods --all -n cert-manager --timeout=300s

# Apply Infrastructure and Application
kubectl apply -f internal/pascalallen/infrastructure/etc/k8s/cert-manager \
    && kubectl apply -f internal/pascalallen/infrastructure/etc/k8s/go
```

### Retrieve `EXTERNAL-IP`

```bash
kubectl get service ingress-nginx-controller --namespace=ingress-nginx
```

### Set up a DNS record pointing to the `EXTERNAL-IP`

## Trusted proxies

The visit ping rate limit is per client address. Behind a reverse proxy or ingress, set `TRUSTED_PROXIES` to a comma-separated list of the proxy CIDRs or IPs (for example `10.0.0.0/8`) so the forwarded client address is used. Left empty, no proxy is trusted: every visitor appears as the proxy address and the per-address limit collapses to one ping per hour for the whole site. A global cap of 30 notifications per hour applies regardless.
