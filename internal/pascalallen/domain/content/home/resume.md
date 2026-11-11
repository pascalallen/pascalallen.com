# Pascal Allen

Senior Software Engineer · Go / PHP / React / TypeScript / Postgres · Austin, TX

Email     pascal.allen88@gmail.com
Site      https://pascalallen.com
GitHub    https://github.com/pascalallen
Medium    https://pascalallen.medium.com
LinkedIn  https://www.linkedin.com/in/pascal-allen-942749112

# Summary

Full-stack software engineer with 10+ years of experience designing, developing, and deploying business applications. My current stack includes Go, React, TypeScript, Postgres, Kubernetes, Docker, and Claude Code. I've deployed and managed production systems on GCP, AWS, and Digital Ocean.

# Technical Skills

Languages
  Go, PHP, React, TypeScript, SQL, Kotlin

Backend & APIs
  REST, gRPC, Twirp, GraphQL, WebSockets, Stripe, Gin, Symfony, Laravel

Data & Messaging
  Postgres, MySQL, MSSQL, Redis, RabbitMQ, Kafka, EventStoreDB

Auth & Security
  JWT, TOTP 2FA, refresh-token rotation, HMAC request signing, AES-GCM secret encryption

Frontend
  React, TypeScript, TanStack Query, Webpack

Cloud & Infra
  GCP, AWS, DigitalOcean, Kubernetes, Helm, Docker, CI/CD, GitHub Actions, Jenkins, Terraform, Flux, Spacelift, Elastic Cloud, Elasticsearch

Observability
  Prometheus, Grafana, Splunk, CloudWatch, PagerDuty

AI Tooling
  MCP server authoring, Claude Code skills, multi-agent pipelines

Practices
  DDD, CQRS, TDD, SOLID, event-driven architecture, event sourcing, hexagonal architecture, unit/integration testing, progressive feature-flag delivery, code review, Agile/Scrum, event storming

# Professional Experience

Sectigo — Senior Software Engineer (Contract, via imagineX)
Remote (Austin, TX) · 2026 – Present

- Currently developing https://www.sectigo.com/qspm-quantum-ready

Stack: GCP, Docker, Kubernetes, Terraform, Helm, Flux, Spacelift, Grafana, Elasticsearch, Node, Python, JavaScript, Claude, Copilot

Intuit Mailchimp — Senior Software Engineer (Contract, via Tensure)
Remote (Austin, TX) · 2024 – 2026

- Full-stack engineer on the Registration, Delivery, and Segmentation squads, working across a PHP/React monolith, Go microservices, and a Kotlin/Spring CDP service: 764 merged PRs across 18 repositories in ~20 months
- Found that a shipped segmentation upgrade was silently returning zero matches despite a green test suite; traced it to a query shape the downstream data platform didn't support, got the cross-team fix in, and re-checked against production data before re-shipping across four service layers.
- Shipped five audience-segmentation conditions end-to-end (PHP, React, ANTLR-driven Kotlin CDP service, experiment + feature flag), including the regulated Age Gating rollout and Product Collection/Variant targeting; the latter needed a cross-org initiative I drove to expose e-commerce event attributes to segmentation.
- Built WhatsApp campaign creation, Quiet Hours, message-event processing, and retry/throttling; delivered the RCS rollout across five services (rate cards, credit reservations, delivery gateway).
- Fixed performance and correctness bugs on high-traffic paths: a memory-exhaustion bug in the segment-members API caused by pagination-cursor tracking, plus three audience bugs: corrupt segment ASTs that matched everyone, and two inverted booleans that returned the wrong audience.
- Led the SMS Delivery test-coverage epic, adding unit and integration tests that raised coverage on the core Go delivery services (7+) from ~40% to 85%. On-call for messaging delivery; 300+ staged production ramps and 120+ feature-flag cleanups with a revert rate under 1%.
- Built Claude Code tooling for multi-repo work: 19 skills (feature-flag lifecycle, full-stack code generation, test payloads) and a multi-agent pipeline that takes a Jira ticket through triage, parallel implementation, and automated review to a commit that still goes through normal PR review and CI.

Stack: Go, PHP, Kotlin, React, TypeScript, gRPC, Twirp, GraphQL, Kafka, Prometheus, GCP, Kubernetes, Spanner, BigQuery, Terraform, Jenkins, Argo Rollouts, Splunk, PagerDuty

Monumental — Senior Software Engineer
Austin, TX · 2022 – 2024

- Designed and launched the NFT marketplace product letting customers purchase NFTs using their attached wallets.
- Designed and launched the CMS product letting companies publish and manage assets available to the public on the NFT marketplace product.
- Both products were acquired for $6M+.
- Contributed frontend improvements to the Arcane: League of Legends web application.
- Ran code reviews and mentored junior developers.
- Authored the org's engineering docs including version control, releasing, context mapping, and event storming.

Stack: AWS, Kubernetes, Postgres, PHP, Symfony, Mercure, RabbitMQ, Python, React, TypeScript

OnlineMedEd — Software Engineer II
Austin, TX · 2018 – 2021

- Designed and launched the admin BI platform with dashboards on product usage, revenue, and analytics. Was used frequently by stakeholders and executives.
- Designed and launched the faculty product letting institutions delegate premium subscriptions to students on our core product
  https://www.onlinemeded.com/
- Extracted the billing service from our core product to its own microservice and added support for different currencies.

Stack: AWS, Kubernetes, PHP, Laravel, Symfony, Mercure, RabbitMQ, React, TypeScript, Webpack, Redux Saga, Postgres, MySQL

Far West Capital — Lead Software Engineer
Austin, TX · 2016 – 2018

- Led development and architecture of the factoring firm's internal operations software and client-facing web apps (client portal, credit application, secure document upload), primarily PHP against MSSQL/MySQL/T-SQL data sources; built the proprietary reporting software (SQL Server Reporting Services plus custom reports) used in daily operations.
- Owned the infrastructure alongside the code: server management and hardware upgrades, LogicMonitor host monitoring with alerting, Cisco Meraki network security, Rapid7 assessments, and Trend Micro; earned the Far West Five Award for execution and excellence.

Stack: Go, PHP, React, TypeScript, MSSQL, MySQL, T-SQL

# Consulting Projects

Crimson Drive Design
Austin, TX · 2015 – Present

- Independent consultancy and product studio delivering RESTful services, web apps, and CI/CD infrastructure for business clients, rooted in domain-driven design and event storming.
- Currently operating a school-dismissal app that streamlines end-of-day student pickup for parents and faculty. Leverages TanStack for state management on the FE, Stripe payment processing and subscriptions, JWT refresh-token rotation, and WebSocket fan-out over Postgres NOTIFY/LISTEN.
- Designed and launched a SaaS emailing meeting-prep briefs from users' calendars: Microsoft Graph, Google Calendar, and LinkedIn integrations, Stripe billing, dual bearer/cookie SPA auth.

Stack: AWS, GCP, Digital Ocean, Go, Gin, PHP, Laravel, React, TanStack, TypeScript, Postgres, JWT

# Open Source & Writing

- pgmcp — An MCP server developed in Go that offers read-only Postgres ops to query, analyze, and inspect your database.
- hmac (v2) — Repository for the pascalallen/hmac Go module. Offers services for HTTP HMAC authentication and request signing.
- pubsub (v2) — Repository for the pascalallen/pubsub Go module. Offers a small, concurrency-safe, in-memory pub/sub, built on goroutines and channels. A Topic fans out every published Message to each of its subscribers.
- pgqueue (v1.1) — Repository for the pascalallen/pgqueue Go module. Offers a Postgres-backed durable job queue and NOTIFY/LISTEN pub/sub.
- ollama-cloud-go — Dependency-free Go client for the Ollama Cloud API.
- mcp-server-go — An MCP server in Go that offers tools to read and write to a contextual knowledge base.
- es-go — Open source software offering a set of guidelines to use Event Sourcing and Domain-Driven Design in Go. Uses EventStoreDB.
- rpi-security-cam — Live camera streaming on a Raspberry Pi, served over HTTP in Go.
- adk-agent — Leverages Google's Agent Development Kit to build a basic AI agent.
- vllm-server — A high-performance Docker Compose setup for running vLLM as an OpenAI-compatible API server, specifically optimized for ARM64 CPU environments.
- react-form-components (v3.1) — Repository for the @pascalallen/react-form-components NPM package. Simple HTML form elements for React.
- react-event-dispatcher (v2) — Repository for the @pascalallen/react-event-dispatcher NPM package. A lightweight event dispatcher for React.
- pascalallen.com — Personal portfolio website. Go/React/TypeScript site on Kubernetes (k3s) with a RabbitMQ-backed async command bus, SSE, and a WebAssembly module.

# Education & Certifications

- Codeup — Full-Stack Web Development
- Harvard Online — CS50
- Google — Google Cloud Sales Credential
- Intuit — Secure Coding for Web Applications
- Anthropic — Claude Code 101
- Anthropic — Claude Code in Action
- Docker — Docker Foundations Professional Certificate
