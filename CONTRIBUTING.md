# Contributing to the Nodela Go SDK

Thank you for your interest in contributing! This guide covers everything you need to know to submit bug reports, propose features, and open pull requests.

---

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Ways to Contribute](#ways-to-contribute)
- [Development Setup](#development-setup)
- [Project Structure](#project-structure)
- [Running Tests](#running-tests)
- [Linting and Formatting](#linting-and-formatting)
- [Writing Tests](#writing-tests)
- [Changelog Management](#changelog-management)
- [Submitting a Pull Request](#submitting-a-pull-request)
- [Commit Message Style](#commit-message-style)
- [Versioning Policy](#versioning-policy)

---

## Code of Conduct

This project follows the [Contributor Covenant Code of Conduct](https://www.contributor-covenant.org/version/2/1/code_of_conduct/). By participating you agree to uphold a welcoming, inclusive environment. Report unacceptable behaviour to the maintainers.

---

## Ways to Contribute

- **Bug reports** — Open a GitHub issue with a minimal reproduction case.
- **Feature requests** — Open a GitHub issue describing the use case before writing any code.
- **Documentation fixes** — Typos, clarifications, and example improvements are always welcome.
- **Code contributions** — Bug fixes and agreed-upon features via pull request.

Please search existing issues before opening a new one to avoid duplicates.

---

## Development Setup

### Prerequisites

| Tool | Minimum Version | Purpose |
| --- | --- | --- |
| Go | 1.22 | Build and test |
| golangci-lint | latest | Static analysis |
| changie | latest | Changelog management |
| pre-commit | latest | Git hook runner |

### Clone and bootstrap

```bash
# 1. Fork the repo on GitHub, then clone your fork
git clone https://github.com/<your-username>/nodela-go-sdk.git
cd nodela-go-sdk

# 2. Install Go dependencies
make install

# 3. Install developer tools (golangci-lint, changie)
make install-tools

# 4. Install git hooks (runs fmt, vet, and tests on every commit)
pre-commit install
```

### Verify the setup

```bash
make test   # all unit tests must pass
make lint   # no linter errors
make vet    # no go vet warnings
```

---

## Project Structure

```
nodela-go-sdk/
├── pkg/
│   ├── client/          # Public API surface
│   │   ├── client.go    # NewClient, Option types
│   │   ├── request.go   # Core HTTP transport (doRequest, handleErrorResponse)
│   │   ├── invoice.go   # Invoices service (Create, Verify)
│   │   └── transaction.go # Transactions service (List)
│   ├── errors/
│   │   └── errors.go    # SDKError, APIError, error codes, predicates
│   └── models/
│       ├── base.go      # Shared types (Metadata, PaginationParams)
│       ├── invoice.go   # Invoice request/response models + SupportedInvoiceCurrencies
│       └── transaction.go # Transaction request/response models
├── tests/
│   ├── unit/            # Fast, isolated tests with mock HTTP server
│   └── integration/     # End-to-end workflow tests (also use mock server)
├── .changes/            # Changeset files managed by changie
├── .github/workflows/   # CI pipelines (test, lint, release)
├── Makefile             # Developer task runner
├── .golangci.yml        # Linter configuration
└── .changie.yaml        # Changelog configuration
```

### Guiding principles

1. **No external runtime dependencies** — the SDK uses only the Go standard library and `testify` (test-only). Do not add dependencies without discussing with maintainers first.
2. **Functional options pattern** — all optional client configuration is done via `Option` functions, not struct fields.
3. **Typed errors** — all errors must be either `*SDKError` or `*APIError`. Do not return raw `fmt.Errorf` errors from public functions.
4. **Pointer fields for optional params** — use `*int`, `*string`, etc. to distinguish "not provided" from a zero value in request parameter structs.

---

## Running Tests

### Unit tests (fast, no network)

```bash
make test
```

This runs all tests under `tests/unit/` with race detection and prints a coverage summary.

### Integration tests

```bash
make test-integration
```

Integration tests use an in-process mock HTTP server — no real Nodela API credentials are needed.

### Coverage report

```bash
make coverage   # opens an HTML report in your browser
```

### Running a single test

```bash
go test ./tests/unit/... -run TestInvoiceCreate_Success -v
go test ./tests/integration/... -run TestInvoiceLifecycle -v
```

---

## Linting and Formatting

```bash
make fmt    # runs gofmt + go mod tidy
make lint   # runs golangci-lint with the project config
make vet    # runs go vet
```

All three must pass before a pull request is merged. The CI pipeline enforces this automatically. The enabled linters are defined in [.golangci.yml](.golangci.yml) and include:

- `gofmt` / `goimports` — code formatting
- `govet` — correctness checks (including shadow variable detection)
- `errcheck` — unhandled errors
- `staticcheck` — advanced static analysis
- `gosec` — security linting
- `revive` — Go style guide enforcement
- `misspell` — common spelling mistakes

---

## Writing Tests

### Unit tests

- Place all unit tests in `tests/unit/`.
- Use the `newTestClient()` helper from `helpers_test.go` to create a client wired to an `httptest.Server`.
- Each test function must be deterministic and independent. Do not share mutable state between tests.
- Cover both the happy path and all documented error cases.
- Use `testify/assert` and `testify/require` for assertions.

```go
func TestMyFeature_Success(t *testing.T) {
    client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
        assert.Equal(t, http.MethodPost, r.Method)
        writeJSON(t, w, http.StatusOK, mySuccessResponse)
    })
    defer server.Close()

    result, err := client.MyService.DoSomething(context.Background(), params)
    require.NoError(t, err)
    assert.Equal(t, "expected_value", result.Data.Field)
}
```

### Integration tests

- Place workflow tests in `tests/integration/`.
- Test multi-step flows (e.g., create invoice → verify payment).
- Use the same mock server pattern but focus on the interaction between multiple SDK calls.

### Test naming convention

```
Test<TypeName>_<Scenario>
Test<TypeName>_<Scenario>_<Condition>
```

Examples:
- `TestInvoiceCreate_Success`
- `TestInvoiceCreate_InvalidCurrency`
- `TestTransactionList_Pagination_HasMore`

---

## Changelog Management

This project uses [changie](https://changie.dev) for changelog management. **Every pull request that changes behaviour (bug fix, new feature, deprecation, breaking change) must include a changeset file.**

### Create a changeset

```bash
make changeset
```

This launches an interactive prompt asking for the kind of change and a short description. It generates a file under `.changes/unreleased/`. Commit this file along with your code changes.

### Change kinds

| Kind | When to use |
| --- | --- |
| `Added` | New features or capabilities |
| `Changed` | Behaviour changes to existing features |
| `Deprecated` | Features marked for future removal |
| `Removed` | Deleted features or APIs |
| `Fixed` | Bug fixes |
| `Security` | Vulnerability fixes |

### Preparing a release (maintainers only)

```bash
make version VERSION=v1.2.0
```

This merges all unreleased changesets into `CHANGELOG.md` and bumps the version.

---

## Submitting a Pull Request

1. **Fork** the repository and create a branch from `main`:
   ```bash
   git checkout -b fix/invoice-currency-validation
   ```

2. **Make your changes**, keeping each commit focused on a single concern.

3. **Write or update tests** — new behaviour must be covered by tests. Bug fixes must include a regression test.

4. **Run the full check suite locally**:
   ```bash
   make fmt && make vet && make lint && make test
   ```

5. **Create a changeset**:
   ```bash
   make changeset
   ```

6. **Push and open a pull request** against the `main` branch. Fill in the PR template, linking the relevant issue if applicable.

7. A maintainer will review your PR. Please respond to feedback promptly. Once approved, a maintainer will merge it.

### Pull request checklist

- [ ] Tests pass (`make test`)
- [ ] Linter passes (`make lint`)
- [ ] Changeset created (`make changeset`)
- [ ] Public API changes are reflected in [docs/api-reference.md](./docs/api-reference.md)
- [ ] No new external runtime dependencies added (unless discussed)

---

## Commit Message Style

Use the [Conventional Commits](https://www.conventionalcommits.org/) format:

```
<type>(<scope>): <short description>

[optional body]

[optional footer]
```

| Type | Use for |
| --- | --- |
| `feat` | New feature |
| `fix` | Bug fix |
| `docs` | Documentation only |
| `refactor` | Code restructuring without behaviour change |
| `test` | Adding or fixing tests |
| `chore` | Build scripts, CI, tooling |
| `perf` | Performance improvements |

Examples:

```
feat(invoices): add webhook URL validation
fix(errors): correct IsRateLimitError to check 429 status code
docs: update supported currencies table in README
test(transactions): add pagination boundary test cases
```

---

## Versioning Policy

This project follows [Semantic Versioning](https://semver.org/):

- **PATCH** (`v1.0.x`) — backwards-compatible bug fixes.
- **MINOR** (`v1.x.0`) — new backwards-compatible features.
- **MAJOR** (`vx.0.0`) — breaking API changes.

Breaking changes must be discussed in an issue before implementation. Deprecations are announced one minor version before removal.

---

Thank you for contributing!
