# Terraform Provider: Lago

This provider manages Lago resources using the Lago API.

## Current MVP

- Provider configuration with `api_endpoint` and `api_key`
- Resource: `lago_billable_metric`
- Resource: `lago_plan`

## Development

```bash
make test
make build
```

Acceptance tests require:

- `LAGO_API_ENDPOINT`
- `LAGO_API_KEY`
- `LAGO_ACC=1`

Then run:

```bash
make testacc
```
