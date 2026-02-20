# Agent Instructions: terraform-provider-lago

This is a Terraform provider for [Lago](https://www.getlago.com/), an open-source usage-based billing platform. It is written in Go using the [Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework) and targets the OpenTofu registry at `registry.opentofu.org/rpcpool/lago`.

## Commit Messages

All commits **must** follow the [Conventional Commits specification](https://www.conventionalcommits.org/en/v1.0.0/):

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

**Allowed types:** `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `chore`, `ci`

**Version bump mapping:**
- `fix:` → patch release
- `feat:` → minor release
- `feat!:` or `BREAKING CHANGE:` footer → major release

**Examples:**
```
feat: add lago_subscription resource
fix(client): retry on 429 response codes
docs: update provider configuration example
chore: bump golangci-lint to v2
feat!: remove deprecated api_token attribute
```

PR titles must also follow this format as they become the squash commit message.

## Repository Structure

```
.
├── internal/
│   ├── client/          # HTTP client for the Lago REST API
│   │   ├── client.go    # Base client: auth, retry, error handling
│   │   ├── billable_metrics.go
│   │   └── plans.go
│   └── provider/        # Terraform Plugin Framework resources
│       ├── provider.go  # Provider configuration (api_endpoint, api_key)
│       ├── resource_billable_metric.go
│       └── resource_plan.go
├── examples/            # Example .tf files for documentation generation
├── docs/                # Generated provider documentation (do not edit manually)
├── tools/               # Code generation tooling (tfplugindocs)
├── main.go              # Provider entry point
├── .goreleaser.yaml     # Multi-platform release build config
└── .golangci.yml        # Linter configuration
```

## Development Commands

```bash
make build       # Compile the provider
make install     # Build and install locally
make fmt         # Format Go code with gofmt -s
make lint        # Run golangci-lint
make generate    # Regenerate docs from schema (requires Terraform)
make test        # Run unit tests (no API required)
make testacc     # Run acceptance tests (requires live Lago instance)
```

## Adding a New Resource

1. Add API types and CRUD methods to a new file in `internal/client/` (e.g. `subscriptions.go`), following the patterns in `billable_metrics.go`.
2. Add the resource implementation in `internal/provider/resource_<name>.go` implementing `resource.Resource`, `resource.ResourceWithConfigure`, and `resource.ResourceWithImportState`.
3. Register it in the `Resources()` slice in `internal/provider/provider.go`.
4. Add an example in `examples/resources/lago_<name>/resource.tf`.
5. Run `make generate` to regenerate documentation.
6. Add acceptance tests in `internal/provider/resource_<name>_test.go`.

## Client Conventions

- The `client.Client` authenticates with a Bearer token via the `Authorization` header.
- All requests go through `doRequest()`, which handles JSON marshalling, retries (up to 3 attempts), and error parsing.
- Retries are performed on network errors, `429 Too Many Requests`, and `5xx` responses.
- Use `client.IsNotFound(err)` to check for 404 responses — resources should call `resp.State.RemoveResource(ctx)` when this occurs in `Read`.
- API errors are returned as `*client.APIError` with a `StatusCode` field.

## Resource Conventions

- Resource IDs are set to the resource's unique `code` field (not a UUID), matching Lago's API identifier.
- Import is supported via `code` on all resources.
- Optional fields that are empty strings from the API should be mapped to `types.StringNull()`, not empty string values.
- Use `stringplanmodifier.RequiresReplace()` on fields that cannot be updated in-place (e.g. `code`).
- Schema descriptions use `MarkdownDescription`, not `Description`.

## Provider Configuration

The provider is configured with:
- `api_endpoint` — Lago API base URL (also read from `LAGO_API_ENDPOINT`)
- `api_key` — Lago API key (also read from `LAGO_API_KEY`, marked `Sensitive: true`)

Provider data is passed to resources as `*lagoProviderData` via `Configure()`.

## Testing

- Unit tests: `make test` — no external dependencies required.
- Acceptance tests: `make testacc` — requires a running Lago instance and:
  - `LAGO_API_ENDPOINT` — e.g. `http://localhost:3000/api/v1`
  - `LAGO_API_KEY` — API key for the instance
  - `TF_ACC=1` — set automatically by `make testacc`

## CI / Release Workflow

- **Tests** run on every PR and push via `.github/workflows/test.yml` (build, lint, generate check, acceptance tests).
- **Conventional commit enforcement** runs on PRs via `.github/workflows/conventional-commits.yml` — all commits and the PR title are checked.
- **Releases** are managed by [release-please](https://github.com/googleapis/release-please) via `.github/workflows/release-please.yml`. Do **not** push `v*` tags manually — merge the release-please PR to trigger a release.
- GoReleaser handles multi-platform binary builds and GPG signing on release.

## Linting

The project uses `golangci-lint` with the configuration in `.golangci.yml`. Key rules:
- Do not import `github.com/hashicorp/terraform-plugin-sdk/v2` — use `terraform-plugin-framework` instead.
- `gofmt -s` formatting is enforced.
- Run `make lint` before pushing.
