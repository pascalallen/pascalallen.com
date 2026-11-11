# 0006. Remove the Kubernetes deploy path

Status: accepted. Supersedes the Kubernetes lines of `docs/DEPLOYING.md` and any mention of manifests as an alternate deploy path in earlier ADRs.

## Context

Production is one Raspberry Pi running Docker Compose, redeployed by `scripts/cron-deploy`. The Kubernetes manifests and the Docker Hub push in `bin/deploy` existed only to feed each other, nothing ran them.

## Decision

Delete `internal/pascalallen/infrastructure/etc/k8s/`, `bin/deploy` and the Kubernetes section of `docs/DEPLOYING.md`. Compose on the Pi is the only deploy path.

## Consequences

- No unused deploy artefact can drift from the app.
- A second deploy target needs a new ADR and its own check in `bin/verify`.
- Kubernetes still appears in the content tree as the owner's work history (resume, skills, experience, a writing entry); that is content, not infrastructure.
