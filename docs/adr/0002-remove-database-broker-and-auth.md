# 0002. Remove the database, the message broker and auth

Status: accepted

## Context

Users, roles, JWT, GORM with Postgres and RabbitMQ buses exist only to guard a `/temp` demo; no visitor uses them. They cost a database and a broker on a Raspberry Pi, a startup dependency chain and a growing list of vulnerable modules. The page config also shipped a GitHub token and a Slack webhook URL to every visitor.

## Decision

Remove auth, users, Postgres/GORM, RabbitMQ, `/login`, `/temp` and the SSE demo. The site is one Go binary with no database or broker. Commands and queries use synchronous in-process buses with `ctx` threaded through. Secrets stay server-side: live data is fetched by the server and the visit ping runs behind a rate-limited endpoint.

## Consequences

- The binary boots with no external dependencies; compose loses postgres, rabbitmq and adminer.
- Single-instance deploys only, which already matches the Pi.
- Handlers run in the request, so an action returns after its command has run.
- The leaked GitHub token and Slack webhook must be rotated by the owner.
- Reintroducing persistence or async work later needs a new ADR.
