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
kubectl apply -f internal/pascalallen/etc/k8s/postgres \
    && kubectl apply -f internal/pascalallen/etc/k8s/rabbitmq \ # TODO
    && kubectl apply -f internal/pascalallen/etc/k8s/go
```
