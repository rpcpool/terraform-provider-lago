# Lago Provider

The Lago provider configures access to the Lago API.

## Example Usage

```hcl
provider "lago" {
  api_endpoint = var.lago_api_endpoint
  api_key      = var.lago_api_key
}
```

You can also configure these values via environment variables:

- `LAGO_API_ENDPOINT`
- `LAGO_API_KEY`

## Schema

### Required/Optional Attributes

- `api_endpoint` (String, Optional) Lago API endpoint. If omitted, the provider reads `LAGO_API_ENDPOINT`.
- `api_key` (String, Optional, Sensitive) Lago API key. If omitted, the provider reads `LAGO_API_KEY`.

## Resources

- `lago_billable_metric`
- `lago_plan`
