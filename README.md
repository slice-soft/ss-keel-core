<img src="assets/boat.svg" width="400" />

# Keel

Keel is a Go framework for building REST APIs with modular 
architecture, automatic OpenAPI, and built-in validation.

[![CI](https://github.com/slice-soft/ss-keel-core/actions/workflows/ci.yml/badge.svg)](https://github.com/slice-soft/ss-keel-core/actions)
![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go&logoColor=white)
[![Go Report Card](https://goreportcard.com/badge/github.com/slice-soft/ss-keel-core)](https://goreportcard.com/report/github.com/slice-soft/ss-keel-core)
[![Go Reference](https://pkg.go.dev/badge/github.com/slice-soft/ss-keel-core.svg)](https://pkg.go.dev/github.com/slice-soft/ss-keel-core)
![License](https://img.shields.io/badge/License-MIT-green)
![Made in Colombia](https://img.shields.io/badge/Made%20in-Colombia-FCD116?labelColor=003893)

## Philosophy
Go is excellent at giving you the tools to build anything. Keel is opinionated about *how* to organize it.

**Structure over flexibility** — a clear module → controller → service → repository pattern so any developer can navigate the codebase without a map.

**Explicit over magic** — no decorators, no reflection tricks, no code generation. Everything is plain Go that you can read and understand.

**Documentation as a first-class citizen** — OpenAPI is generated at runtime from your route definitions. If the route exists, the docs exist.

**Idiomatic Go** — builder pattern, interfaces, standard library where possible. No compromises on how Go should feel.

## Getting Started
```go
go get github.com/slice-soft/ss-keel-core
```

Check out the documentation at [keel-go.dev](https://keel-go.dev) for guides and API reference.


## Contributing
https://keel-go.dev/contributing