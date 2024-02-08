# pascalallen.com

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/pascalallen/pascalallen.com)
[![Go Report Card](https://goreportcard.com/badge/github.com/pascalallen/pascalallen.com)](https://goreportcard.com/report/github.com/pascalallen/pascalallen.com)
![GitHub Workflow Status (with branch)](https://img.shields.io/github/actions/workflow/status/pascalallen/pascalallen.com/go.yml?branch=main)
![GitHub](https://img.shields.io/github/license/pascalallen/pascalallen.com)
![GitHub code size in bytes](https://img.shields.io/github/languages/code-size/pascalallen/pascalallen.com)

pascalallen.com is a containerized web application built with Kubernetes, Docker, RabbitMQ, Postgres, Go, React, 
TypeScript, Sass, Webpack, and WebAssembly. This ongoing project is designed, developed, deployed, and 
maintained by myself, Pascal Allen. The motivation behind this project was to develop a scalable and portable 
framework that can be used as a template for future web applications. This project is based on domain-driven design, 
hexagonal architecture, and the SOLID principles.

## Features

- Configurable CI/CD pipeline
- Helper scripts
- MobX store
- JWT/HMAC authentication services
- Middleware
- Frontend linting with ESLint and Prettier
- GORM ORM
- Database seeds for permissions, roles, and users
- Database seeder
- Domain models
- Kubernetes config with deployment instructions
- API endpoints for authentication and registration
- API endpoints for server-sent events
- Repositories
- WebAssembly template

## Prerequisites

- [Docker](https://www.docker.com/)
- [Docker Compose](https://docs.docker.com/compose/)

## Development Environment Setup

### Clone Repository

```bash
cd <projects-parent-directory> && git clone https://github.com/pascalallen/pascalallen.com.git
```

### Copy & Modify `.env` File

```bash
cp .env.example .env
```

### Bring Up Environment

```bash
bin/up <prod>
``` 

or (to watch for backend changes)

```bash
bin/watch
```

You will find the site running at [http://localhost:9990/](http://localhost:9990/)

### Install JavaScript Dependencies

```bash
bin/yarn ci
```

### Compile TypeScript with Webpack

```bash
bin/yarn build
```

### Watch For Frontend Changes

```bash
bin/yarn watch
```

### Take Down Environment

```bash
bin/down <prod>
```
