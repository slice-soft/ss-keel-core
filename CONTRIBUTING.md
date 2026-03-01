
# Contributing to ss-keel-core

The base contributing guide — workflow, commit conventions, PR guidelines, and community standards — lives in the [ss-community](https://github.com/slice-soft/ss-community/blob/main/CONTRIBUTING.md) repository. Read it first.

This document covers only what is specific to this repository.

---

## Requirements

- Go 1.25+
- Git

## Setup

```bash
git clone https://github.com/your-username/ss-keel-core.git
cd ss-keel-core
go mod download
```

## Running tests

```bash
go test ./...
go test ./... -cover
```

## Repository-specific rules

- All public API changes require an update to the corresponding test in `openapi/builder_test.go` or `core/route_test.go`
- Use table-driven tests as the preferred pattern — see existing `*_test.go` files for examples
- Do not use external test frameworks — standard `testing` package only
- Use `t.Setenv()` for environment variable tests — cleanup is automatic
- Breaking changes must follow the policy in [VERSIONING.md](https://github.com/slice-soft/ss-community/blob/main/VERSIONING.md)

## Questions

Open a [Discussion](https://github.com/slice-soft/ss-keel-core/discussions) instead of an issue for questions about the codebase or implementation approach.
