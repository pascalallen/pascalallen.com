# pascalallen.com

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/pascalallen/pascalallen.com)
[![Go Report Card](https://goreportcard.com/badge/github.com/pascalallen/pascalallen.com)](https://goreportcard.com/report/github.com/pascalallen/pascalallen.com)
![GitHub Workflow Status (with branch)](https://img.shields.io/github/actions/workflow/status/pascalallen/pascalallen.com/go.yml?branch=main)
![GitHub](https://img.shields.io/github/license/pascalallen/pascalallen.com)
![GitHub code size in bytes](https://img.shields.io/github/languages/code-size/pascalallen/pascalallen.com)

My personal website.

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
bin/up
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
bin/down
```
