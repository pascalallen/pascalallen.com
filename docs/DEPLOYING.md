# Deploying

This document describes the steps necessary for deploying to a Kubernetes cluster.

## Prerequisites

- [Docker](https://www.docker.com/)
- [Docker Compose](https://docs.docker.com/compose/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl)
- [A provisioned Kubernetes cluster that you can connect to](https://kubernetes.io/docs/home/#set-up-a-k8s-cluster)

### Create New Docker Images

```bash
docker compose -f compose.yaml build
```

### Tag Docker Images

```bash
docker tag <image-id> <dockerhub-username>/<repository-name>
```

I.e. `docker tag f3375693ddb5 pascalallen/pascalallen-postgres && docker tag bc24df454cb8 pascalallen/pascalallen-go`

### Log in to Docker Hub

```bash
docker login
```

### Push New Docker Images to Docker Hub

```bash
docker push pascalallen/pascalallen-postgres && docker push pascalallen/pascalallen-go
``` 

### Create Kubernetes Secret from `.env` file

```bash
kubectl create secret generic env-vars --from-env-file=.env
```

### Apply Services to Kubernetes Cluster

```bash
kubectl apply -f internal/pascalallen/etc/k8s/postgres && kubectl apply -f internal/pascalallen/etc/k8s/go
```
