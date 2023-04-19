# pascalallen.com

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/pascalallen/pascalallen.com)
[![Go Report Card](https://goreportcard.com/badge/github.com/pascalallen/pascalallen.com)](https://goreportcard.com/report/github.com/pascalallen/pascalallen.com)
![GitHub Workflow Status (with branch)](https://img.shields.io/github/actions/workflow/status/pascalallen/pascalallen.com/go.yml?branch=main)
![GitHub](https://img.shields.io/github/license/pascalallen/pascalallen.com)
![GitHub code size in bytes](https://img.shields.io/github/languages/code-size/pascalallen/pascalallen.com)

My personal website.

## Prerequisites

- [Go](https://go.dev/doc/install)
- [Yarn](https://classic.yarnpkg.com/en/docs/install)

## Development Environment Setup

### Clone Repository

```bash
cd <projects-parent-directory> && git clone https://github.com/pascalallen/pascalallen.com.git
```

### Copy & Modify `.env` File

```bash
cp .env.example .env
```

### Compile & Run Main Go Package

```bash
go run main.go
``` 

### Install JavaScript Dependencies

```bash
cd client && yarn install --frozen-lockfile
```

### Compile TypeScript

```bash
cd client && yarn tsc
```

You will find the site running at [http://localhost:9990/](http://localhost:9990/)