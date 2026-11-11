# Deploying

Production is a Raspberry Pi (arm64) running Docker Compose. Kubernetes is an alternate path.

## Docker Compose on the Raspberry Pi (primary)

`compose.yaml` is the portable base: one service, port `9990` (override with `PORT`), `.env` as `env_file`, restart
policy `unless-stopped`, no devices, not privileged. `compose.pi.yaml` adds the Pi-only parts: privileged, `/run/udev`,
`/dev/media0-3`, `/dev/video0`, and `user: "0:0"` so the camera devices are readable. The image itself runs as a
non-root user; the Pi overlay overrides that because the container is privileged anyway.

### One-time setup

```bash
git clone https://github.com/pascalallen/pascalallen.com.git && cd pascalallen.com
cp .env.example .env
```

Fill in `.env` (`APP_BASE_URL`, `GITHUB_TOKEN`, `SLACK_DM_URL`, `GIN_MODE=release`), then start it once:

```bash
bin/up prod
```

Add the redeploy job with `crontab -e`, using the absolute path to the clone:

```cron
* * * * * /home/pi/pascalallen.com/scripts/cron-deploy >> /home/pi/pascalallen-deploy.log 2>&1
```

### What `scripts/cron-deploy` does

- takes a lock so runs never overlap (a stale lock from a dead process is reclaimed)
- runs `git fetch origin main` and exits quietly when `HEAD` already equals `origin/main`
- refuses to touch a working tree with local changes: it logs, posts one Slack line per new commit, and leaves the
  tree alone
- otherwise checks out `main`, resets to `origin/main`, rebuilds and restarts with `compose.yaml` plus `compose.pi.yaml`,
  and polls `http://localhost:$PORT/` until it answers
- on a failed build or health check it resets to the previous commit, rebuilds, and posts a rollback line; the failed
  commit is remembered in `.git/cron-deploy-failed` so it is not retried every minute (push a new commit to retry)
- posts to `SLACK_DM_URL` from `.env` on deploy, rollback, and failure; the URL is passed to curl on stdin, never echoed

Tests for the decision logic run without Docker: `scripts/cron-deploy.test.sh`.

### Day to day

```bash
bin/up        # build, start, follow logs (Pi overlay is picked automatically on the Pi)
bin/up base   # force the portable base file
bin/up pi     # force the Pi overlay
bin/up prod   # Pi overlay, quiet, detached
bin/down      # same selectors as bin/up
```

## Kubernetes (alternate)

The manifests under `internal/pascalallen/infrastructure/etc/k8s/` run only the app (deployment, service, ingress with a
cert-manager issuer). There is no database or message broker. The camera needs a Pi, so a cluster deployment serves the
site without a working `/api/v1/camera/stream`.

### Prerequisites

- [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl)
- [A provisioned Kubernetes cluster that you can connect to](https://kubernetes.io/docs/home/#set-up-a-k8s-cluster)
- The image `pascalallen/pascalallen-go:latest` pushed to Docker Hub: `docker login && bin/deploy`

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
