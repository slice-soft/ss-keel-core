
# Contributing to Keel

Thank you for your interest in contributing to Keel! This document explains how to get started.

---

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Ways to Contribute](#ways-to-contribute)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Pull Request Guidelines](#pull-request-guidelines)
- [Commit Messages](#commit-messages)
- [Testing](#testing)
- [Reporting Issues](#reporting-issues)

---

## Code of Conduct

Be respectful. We welcome contributors regardless of experience level, background, or origin. Constructive feedback is encouraged — personal attacks are not.

---

## Ways to Contribute

You don't need to write code to contribute:

- **Report a bug** — open an issue with a clear description and reproduction steps
- **Suggest a feature** — open an issue describing the problem you're trying to solve
- **Improve documentation** — fix typos, clarify examples, add missing content
- **Write tests** — increase coverage for existing functionality
- **Fix a bug** — pick an open issue labeled `good first issue` or `bug`
- **Implement a feature** — check the roadmap and open issues first to avoid duplicating work

---

## Getting Started

**Requirements:**

- Go 1.25+
- Git

**Fork and clone:**

```bash
git clone https://github.com/your-username/ss-keel-core.git
cd ss-keel-core
go mod download
```

**Run tests:**

```bash
go test ./...
```

**Run tests with coverage:**

```bash
go test ./... -cover
```

---

## Development Workflow

1. **Open or find an issue** — all significant changes should have a corresponding issue before work begins
2. **Fork the repository** and create a branch from `main`
3. **Name your branch** descriptively — `fix/parse-body-error`, `feat/graceful-shutdown`, `docs/validation-examples`
4. **Make your changes** — keep them focused and minimal
5. **Write or update tests** — all new functionality must have tests
6. **Run the full test suite** before submitting
7. **Open a Pull Request** against `main`

---

## Pull Request Guidelines

- **One PR per concern** — do not mix bug fixes with new features
- **Add a semver label** — your PR must have one of: `patch`, `minor`, or `major`
  - `patch` — bug fix, typo, small improvement
  - `minor` — new feature, non-breaking change
  - `major` — breaking change
- **Fill out the PR description** — explain what changed and why, not just what
- **Link the related issue** — use `Closes #123` in the description
- **Keep the PR small** — large PRs are harder to review and slower to merge

---

## Commit Messages

Keel follows **Conventional Commits** to support:

- Automated semantic versioning
- Auto-generated changelogs
- Predictable releases

### Format

```
<type>(optional-scope): short descriptive summary
```

### Examples

```
feat(config): add GetEnvUint helper
fix(parser): handle nil body in ParseBody
docs(readme): add validation examples
test(router): add table-driven tests for route builder
refactor(core): simplify openapi path conversion
chore(ci): update workflow permissions
```

### Rules

- Use present tense
- Keep messages concise but descriptive
- Do not use vague messages like `update`, `fix stuff`, or `wip`
- Do not mix unrelated concerns in a single commit
- Separate features, fixes, refactors, and docs into different commits

### Allowed Types

- `feat` — new feature
- `fix` — bug fix
- `docs` — documentation changes
- `test` — test additions or modifications
- `refactor` — code restructuring without behavior change
- `chore` — maintenance, tooling, config
- `ci` — CI/CD updates
- `perf` — performance improvements

---

## Merge Strategy

Keel uses **Squash & Merge**.

This means:

- Development may contain multiple structured commits
- The Pull Request title becomes the final commit in `main`
- The PR title MUST follow Conventional Commits format

### PR Title Format

```
<type>(optional-scope): high-level summary
```

Example:

```
feat(server): implement graceful shutdown
```

### Important

- Internal commits should remain meaningful
- The PR title should summarize the overall change
- The semver label (`patch`, `minor`, `major`) must match the impact of the PR
---

## Testing

Keel uses standard Go testing with table-driven tests as the preferred pattern.

```go
func TestSomething(t *testing.T) {
    tests := []struct {
        name  string
        input string
        want  string
    }{
        {name: "case one", input: "foo", want: "bar"},
        {name: "case two", input: "baz", want: "qux"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Something(tt.input)
            if got != tt.want {
                t.Errorf("got %v, want %v", got, tt.want)
            }
        })
    }
}
```

**Rules:**

- Every new function must have tests
- Tests live in the same package with `_test.go` suffix
- Use `t.Setenv()` for environment variable tests — cleanup is automatic
- Do not use external test frameworks — standard `testing` package only

---

## Reporting Issues

When opening a bug report, include:

- Go version (`go version`)
- Keel version
- Minimal reproduction code
- Expected vs actual behavior
- Error output if applicable

For security vulnerabilities, do **not** open a public issue. Contact us directly at `slicesoft@juancadev.com` with details.

---

## Questions

If you have questions about the codebase or how to implement something, open a [Discussion](https://github.com/slice-soft/ss-keel-core/discussions) instead of an issue.

---

*Keel is built by [SliceSoft](https://slicesoft.dev) — Colombia 💙*