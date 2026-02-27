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
    && kubectl apply -f internal/pascalallen/infrastructure/etc/k8s/postgres \
    && kubectl apply -f https://github.com/rabbitmq/cluster-operator/releases/latest/download/cluster-operator.yml \
    && kubectl apply -f internal/pascalallen/infrastructure/etc/k8s/rabbitmq \
    && kubectl apply -f internal/pascalallen/infrastructure/etc/k8s/go
```

### Retrieve `EXTERNAL-IP`

```bash
kubectl get service ingress-nginx-controller --namespace=ingress-nginx
```

### Set up a DNS record pointing to the `EXTERNAL-IP`
