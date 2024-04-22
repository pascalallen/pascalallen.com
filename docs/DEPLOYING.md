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
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.10.0/deploy/static/provider/baremetal/deploy.yaml \
    && kubectl apply -f internal/pascalallen/etc/k8s/postgres \
# TODO && kubectl apply -f internal/pascalallen/etc/k8s/rabbitmq \
    && kubectl apply -f internal/pascalallen/etc/k8s/go
```

### Retrieve `EXTERNAL-IP`

```bash
kubectl get service ingress-nginx-controller --namespace=ingress-nginx
```

### Set up a DNS record pointing to the `EXTERNAL-IP`

### Create an Ingress Resource

```bash
kubectl create ingress pascalallen --class=nginx --rule="pascalallen.com/*=pascalallen:80"
```
