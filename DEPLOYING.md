# Deploying

This document describes the steps necessary for deploying to a Kubernetes cluster.

## Prerequisites

- [Docker](https://www.docker.com/)
- [Docker Compose](https://docs.docker.com/compose/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl)
- [A provisioned Kubernetes cluster that you can connect to](https://kubernetes.io/docs/home/#set-up-a-k8s-cluster)

### Create New Docker Images

```bash
docker compose -f docker-compose.yml build
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

TLS work in progress
https://kubernetes.io/docs/concepts/services-networking/ingress/#tls
https://kubernetes.io/docs/concepts/configuration/secret/#tls-secrets

### Generate CA private key using OpenSSl

```bash
openssl genrsa -out ca.key 2048
```

### Generate a self-signed certificate from the private key using OpenSSL

```bash
openssl req -x509 \
  -new -nodes  \
  -days 365 \
  -key ca.key \
  -out ca.crt \
  -subj "/CN=localhost"
```

### Create Kubernetes Secret for TLS (uses DER key and cert)

```bash
kubectl create secret tls go-tls --key=ca.key --cert=ca.crt
```

### Apply Services to Kubernetes Cluster

```bash
kubectl apply -f etc/k8s/postgres && kubectl apply -f etc/k8s/go
```
