# Deploying

Production is a Raspberry Pi (arm64) running Docker Compose. Kubernetes is an alternate path.

## Docker Compose on the Raspberry Pi (primary)

Requires Docker Engine 24+ with Compose v2.20+ (`--progress quiet` is a global Compose flag from that line) and a
64-bit (arm64) userland; `flock` from util-linux (present on Raspberry Pi OS).

`compose.yaml` is the portable base: one service, port `9990` (override with `PORT`), `.env` as a required `env_file`,
restart policy `unless-stopped`, no devices, not privileged. `compose.pi.yaml` adds the Pi-only parts: `platform:
linux/arm64`, the build arg `REQUIRE_CAMERA=1` (the image build fails if the camera packages cannot be installed),
privileged, `/run/udev`, `/dev/media0-3`, `/dev/video0`, and `user: "0:0"` so the camera devices are readable. The image
itself runs as a non-root user; the Pi overlay overrides that because the container is privileged anyway.

### One-time setup (new Pi)

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

### Migrating an existing Pi

The Pi currently runs the old `scripts/cron-deploy`, which never fetches, so merging to `main` changes nothing there.
Do this once, on the Pi, from the clone:

```bash
cd ~/pascalallen.com
git status
```

If `git status` lists changes, look at them with `git diff`; keep what you need elsewhere (`git stash push -u -m pre-migration`
is safe) because the next step discards them. When the tree is clean:

```bash
git fetch origin && git checkout main && git reset --hard origin/main
bin/down prod
bin/up pi
curl -fsS -o /dev/null -w '%{http_code}\n' http://localhost:9990/
docker compose ps
crontab -l | grep cron-deploy
```

The curl must print `200` and the container must show `(healthy)`. The crontab line must still point at
`scripts/cron-deploy` in this clone (the path did not change). Cron needs nothing else; the new script reads `.env`
itself.

Optional cleanup of leftovers from the old setup (postgres, rabbitmq and the old named network). Find them first:

```bash
docker volume ls
docker network ls
docker ps -a
```

If an earlier compose project created volumes, they are named `<project>_<name>` after the clone directory (for example
`pascalallencom_postgres_data`) and the network is `pascalallen`. Once the new stack is healthy:

```bash
docker network rm pascalallen
docker volume rm <the old database volume name from the list above>
```

Removing the volume is irreversible: it deletes the old database for good. Skip it if you might want that data.

### What `scripts/cron-deploy` does

- takes a lock so runs never overlap: `flock -n` on `.git/cron-deploy.flock` (released by the kernel on power loss or
  crash); without `flock` it falls back to a lock directory that is taken over, with an alert, when it is older than
  2 hours
- runs `git fetch origin main` and exits quietly when `HEAD` already equals `origin/main`
- refuses to touch a working tree with local changes, and treats a missing `.env` the same way: it logs, posts one Slack
  line per new commit, and leaves everything alone
- otherwise checks out `main`, resets to `origin/main`, builds, restarts with `compose.yaml` plus `compose.pi.yaml`, and
  polls `http://localhost:$PORT/` until it answers
- a failed build or failed container start is treated as transient: the tree goes back to the running commit and the
  commit is retried at most every 15 minutes, up to 8 attempts, with a Slack alert on attempts 1, 4 and 8. After the
  last attempt it says how to retry: `rm .git/cron-deploy-retry`
- a failed health check after a successful build marks the commit bad at once, rolls back to the previous commit
  (rebuild and restart) and says how to retry: `rm .git/cron-deploy-failed` (or push a new commit)
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
