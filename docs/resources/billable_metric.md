# Resource: lago_billable_metric

Manages a Lago billable metric.

## Example Usage

```hcl
resource "lago_billable_metric" "api_requests" {
  name             = "API Requests"
  code             = "api_requests"
  aggregation_type = "count_agg"
  description      = "Total number of API requests"
  recurring        = true
  filters = [
    {
      key    = "backend"
      values = ["rpc", "grpc"]
    }
  ]
}
```

## Import

Import a metric by its code:

```bash
terraform import lago_billable_metric.api_requests api_requests
```

## Caveats

- `code` is intentionally treated as immutable by this provider (changing it forces replacement).
- Lago API currently allows updating a billable metric `code` via the update endpoint.
- This immutability is a safety choice to avoid accidental renames and state drift in the MVP.
- TODO: support safe in-place `code` mutation in a future provider release.

## Schema

### Required

- `name` (String)
- `code` (String)
- `aggregation_type` (String) One of: `count_agg`, `sum_agg`, `max_agg`, `latest_agg`, `weighted_sum_agg`, `unique_count_agg`.

### Optional

- `description` (String)
- `field_name` (String)
- `expression` (String)
- `recurring` (Boolean)
- `weighted_interval` (String, only valid when `aggregation_type = "weighted_sum_agg"`)
- `filters` (List of Objects)
  - `key` (String)
  - `values` (List of String)

### Read-Only

- `id` (String)
- `created_at` (String)
- `updated_at` (String)
