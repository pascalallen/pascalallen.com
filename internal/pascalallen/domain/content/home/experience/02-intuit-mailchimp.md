# Intuit Mailchimp — Senior Software Engineer (Contract, via Tensure)
Remote (Austin, TX) · 2024 – 2026

- Full-stack engineer on the Registration, Delivery, and Segmentation squads, working across a PHP/React monolith, Go microservices, and a Kotlin/Spring CDP service: 764 merged PRs across 18 repositories in ~20 months
- Found that a shipped segmentation upgrade was silently returning zero matches despite a green test suite; traced it to a query shape the downstream data platform didn't support, got the cross-team fix in, and re-checked against production data before re-shipping across four service layers.
- Shipped five audience-segmentation conditions end-to-end (PHP, React, ANTLR-driven Kotlin CDP service, experiment + feature flag), including the regulated Age Gating rollout and Product Collection/Variant targeting; the latter needed a cross-org initiative I drove to expose e-commerce event attributes to segmentation.
- Built WhatsApp campaign creation, Quiet Hours, message-event processing, and retry/throttling; delivered the RCS rollout across five services (rate cards, credit reservations, delivery gateway).
- Fixed performance and correctness bugs on high-traffic paths: a memory-exhaustion bug in the segment-members API caused by pagination-cursor tracking, plus three audience bugs: corrupt segment ASTs that matched everyone, and two inverted booleans that returned the wrong audience.
- Led the SMS Delivery test-coverage epic, adding unit and integration tests that raised coverage on the core Go delivery services (7+) from ~40% to 85%. On-call for messaging delivery; 300+ staged production ramps and 120+ feature-flag cleanups with a revert rate under 1%.
- Built Claude Code tooling for multi-repo work: 19 skills (feature-flag lifecycle, full-stack code generation, test payloads) and a multi-agent pipeline that takes a Jira ticket through triage, parallel implementation, and automated review to a commit that still goes through normal PR review and CI.

Stack: Go, PHP, Kotlin, React, TypeScript, gRPC, Twirp, GraphQL, Kafka, Prometheus, GCP, Kubernetes, Spanner, BigQuery, Terraform, Jenkins, Argo Rollouts, Splunk, PagerDuty
