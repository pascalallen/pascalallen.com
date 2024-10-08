# pascalallen.com

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/pascalallen/pascalallen.com)
[![Go Report Card](https://goreportcard.com/badge/github.com/pascalallen/pascalallen.com)](https://goreportcard.com/report/github.com/pascalallen/pascalallen.com)
![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/pascalallen/pascalallen.com/go.yml)
![GitHub](https://img.shields.io/github/license/pascalallen/pascalallen.com)
![GitHub code size in bytes](https://img.shields.io/github/languages/code-size/pascalallen/pascalallen.com)

![Logo](web/static/logo.svg)

pascalallen.com is a containerized web application built with Kubernetes, Docker, RabbitMQ, Postgres, Go, React, 
TypeScript, Sass, Webpack, and WebAssembly. This ongoing project is designed, developed, deployed, and 
maintained by myself, Pascal Allen. 

## Motivation

The motivation behind this project was to develop a scalable and portable framework that can be used as a template 
for web apps and microservices. This project attempts to adhere to [Effective Go](https://go.dev/doc/effective_go) and 
[Organizing a Go module](https://go.dev/doc/modules/layout) by the core Go dev team. Also used in the architecture 
of this project are various design principles such as CQRS, DDD, hexagonal architecture, and SOLID.

## Core Project Tree

```
├── bin/       # Executable CLI commands
├── docs/      # Additional documentation
├── cmd/       # Go commands
├── internal/  # Supporting packages
├── scripts/   # Application-specific scripts
└── web/       # Web app components
```

## Features

- Configurable CI/CD pipeline
- Helper scripts
- MobX store
- Google Wire DI container
- JWT/HMAC authentication services
- RabbitMQ message broker
- Asynchronous command bus
- Asynchronous event dispatcher
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

## Testing

Run tests and create coverage profile:

```bash
bin/exec go test ./... -covermode=count -coverprofile=coverage.out
```

Generate HTML file to view test coverage profile:

```bash
bin/exec go tool cover -html=coverage.out -o coverage.html
```

## Contributing

Pull requests are welcome. For major changes, please open an issue first
to discuss what you would like to change.

Please make sure to update tests as appropriate.

## License

[MIT](LICENSE)
