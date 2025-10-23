# Deploying to Kubernetes (k3s, microk8s, k0s)

This document provides production-ready manifests and step-by-step instructions to run pascalallen.com on a Kubernetes cluster using Postgres and RabbitMQ. You can use k3s, microk8s, k0s, or any CNCF conformant distro.

## Prerequisites
- kubectl installed and pointing at your cluster
- A container registry you can push to (e.g., Docker Hub, GHCR)
- Ingress controller (this guide assumes ingress-nginx)

## 1) Prepare the cluster

### k3s
- Option A (recommended here): Install ingress-nginx
  ```bash
  kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.10.0/deploy/static/provider/baremetal/deploy.yaml
  ```
- Option B: Use Traefik (default in k3s). If you prefer Traefik, change the annotation in internal/pascalallen/infrastructure/etc/k8s/ingress/ingress.yml to
  ```yaml
  metadata:
    annotations:
      kubernetes.io/ingress.class: traefik
  ```

### microk8s
```bash
microk8s enable dns storage ingress
```

### k0s
- Install ingress-nginx as shown above for k3s Option A.

### RabbitMQ Cluster Operator (required)
Install the operator (cluster-wide CRDs and controller):
```bash
kubectl apply -f https://github.com/rabbitmq/cluster-operator/releases/latest/download/cluster-operator.yml
```

## 2) Build and push the application image
Update the image reference in internal/pascalallen/infrastructure/etc/k8s/go/deployment.yml or use kubectl set image later.

You can use the helper script:
```bash
bin/deploy
```

Example (Docker Hub):
```bash
docker buildx build --platform linux/amd64,linux/arm64 -t <DOCKERHUB_USER>/pascalallen-go:prod --push .
# then update the Deployment image:
kubectl -n pascalallen set image deployment/go pascalallen-go=<DOCKERHUB_USER>/pascalallen-go:prod
```

Example (GHCR):
```bash
echo $GHCR_TOKEN | docker login ghcr.io -u <GITHUB_USER> --password-stdin
docker buildx build --platform linux/amd64,linux/arm64 -t ghcr.io/<GITHUB_USER>/pascalallen-go:prod --push .
kubectl -n pascalallen set image deployment/go pascalallen-go=ghcr.io/<GITHUB_USER>/pascalallen-go:prod
```

## 3) Provide configuration and secrets
Create an .env file in the repo root with at least:
```
DB_NAME=<your_db_name>
DB_USER=<your_db_user>
DB_PASSWORD=<your_db_password>
# Optional (defaults in manifests):
# DB_HOST=pascalallen-postgres
# DB_PORT=5432
# RABBITMQ_HOST=rabbitmq
# RABBITMQ_PORT=5672
# GIN_MODE=release
```
Then create/update the env-vars Secret from .env in your target namespace:
```bash
kubectl create ns pascalallen || true
kubectl -n pascalallen create secret generic env-vars --from-env-file=.env --dry-run=client -o yaml | kubectl apply -f -
```
Note: RabbitMQ credentials are provided automatically by the operator via a Secret named rabbitmq-default-user and are injected into the app Deployment.

## 4) Apply the manifests
You can use the helper script:
```bash
bin/k8s-apply
```
This applies, in order:
- Postgres (StatefulSet + Service)
- RabbitMQ (RabbitmqCluster via the operator)
- Go app (Deployment + Service)
- Ingress (nginx, host: pascalallen.local by default)

To apply manually:
```bash
# Postgres (apply only the StatefulSet and Service, not the legacy Deployment/PVC)
kubectl -n pascalallen delete deployment/postgres --ignore-not-found
kubectl -n pascalallen delete pvc/postgres --ignore-not-found
kubectl -n pascalallen apply -f internal/pascalallen/infrastructure/etc/k8s/postgres/statefulset.yml
kubectl -n pascalallen apply -f internal/pascalallen/infrastructure/etc/k8s/postgres/service.yml

# RabbitMQ
data=$(kubectl get crd rabbitmqclusters.rabbitmq.com -o name 2>/dev/null || true); if [ -z "$data" ]; then kubectl apply -f https://github.com/rabbitmq/cluster-operator/releases/latest/download/cluster-operator.yml; fi
kubectl -n pascalallen apply -f internal/pascalallen/infrastructure/etc/k8s/rabbitmq

# App and Ingress
kubectl -n pascalallen apply -f internal/pascalallen/infrastructure/etc/k8s/go
kubectl -n pascalallen apply -f internal/pascalallen/infrastructure/etc/k8s/ingress
```

## 5) Verify and access
```bash
kubectl -n pascalallen get pods,svc,ingress
```

### k3s: Access locally via Ingress
1. Ensure you have an ingress controller:
   - Recommended (matches these manifests): install ingress-nginx on k3s bare-metal
     ```bash
     kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.10.0/deploy/static/provider/baremetal/deploy.yaml
     ```
   - Or use k3s default Traefik: change the annotation in internal/pascalallen/infrastructure/etc/k8s/ingress/ingress.yml to
     ```yaml
     metadata:
       annotations:
         kubernetes.io/ingress.class: traefik
     ```
     then re-apply the ingress:
     ```bash
     kubectl -n pascalallen apply -f internal/pascalallen/infrastructure/etc/k8s/ingress
     ```
2. Get the ingress address (prefer LoadBalancer IP; otherwise use the node InternalIP):
   ```bash
   kubectl get svc -n ingress-nginx ingress-nginx-controller || true
   kubectl get svc -n kube-system traefik || true
   kubectl get nodes -o wide
   ```
3. Map the host to that IP (example using the node InternalIP 192.168.64.10):
   ```
   192.168.64.10 pascalallen.local
   ```
   - Edit /etc/hosts on your workstation or configure local DNS accordingly.
4. Test directly against the IP (helpful before updating hosts):
   ```bash
   curl -H 'Host: pascalallen.local' http://<INGRESS_IP>/ -i
   ```
5. Open the site:
   - http://pascalallen.local/

### Docker Desktop (macOS/Windows): Access via Ingress
- Docker Desktop runs a single-node cluster VM. If you installed ingress-nginx using the baremetal manifest, its Service type will be NodePort and the Ingress will advertise the node InternalIP (e.g., 192.168.65.x). This is expected.
- Steps:
  1) Check the controller and node IP:
     ```bash
     kubectl get svc -n ingress-nginx ingress-nginx-controller || true
     kubectl get nodes -o wide
     kubectl -n pascalallen get ingress pascalallen -o wide
     ```
  2) Add an /etc/hosts entry pointing your Ingress host to the node IP (replace with your IP, e.g., 192.168.65.3):
     ```
     192.168.65.3 pascalallen.local
     ```
  3) Test:
     ```bash
     curl -H 'Host: pascalallen.local' http://192.168.65.3/ -i
     ```
  4) Browse: http://pascalallen.local/
- Tip: Use the helper script to print the exact hosts line (works for Docker Desktop, k3s, microk8s, etc.). You can override the host via INGRESS_HOST.

Tip: use the helper to print the exact /etc/hosts line (it auto-detects the host from your live Ingress). You can override the host by exporting INGRESS_HOST=<your-host>:
```bash
# optionally override
# export INGRESS_HOST=pascalallen.local
bin/k8s-access
```

## Notes on production hardening
- Resource requests/limits and probes are configured for the app and Postgres.
- Persistent storage is provisioned via the cluster’s default StorageClass (change size in the StatefulSet if needed).
- Consider enabling TLS on the Ingress (see commented tls section) using cert-manager and a DNS-validated issuer in production.
- Consider NetworkPolicies to restrict traffic to/from the namespace according to your cluster’s CNI capabilities.
