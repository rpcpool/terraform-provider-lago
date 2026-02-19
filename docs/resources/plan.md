# Resource: lago_plan

Manages a Lago plan.

## Example Usage

```hcl
resource "lago_plan" "starter" {
  name            = "Starter"
  code            = "starter"
  interval        = "monthly"
  amount_cents    = 1000
  amount_currency = "USD"
  description     = "Starter subscription plan"

  tax_codes = ["tax_usa"]

  minimum_commitment = {
    amount_cents = 500
    tax_codes    = ["tax_usa"]
  }

  charges = [
    {
      billable_metric_id = "bm_requests"
      charge_model       = "standard"
      properties_json    = jsonencode({ amount = "0.001" })
    }
  ]
}
```

## Import

Import a plan by its code:

```bash
terraform import lago_plan.starter starter
```

## Caveats

- `code` is intentionally treated as immutable by this provider (changing it forces replacement).

## Schema

### Required

- `name` (String)
- `code` (String)
- `interval` (String) One of: `weekly`, `monthly`, `quarterly`, `yearly`
- `amount_cents` (Number)
- `amount_currency` (String)

### Optional

- `description` (String)
- `trial_period` (Number)
- `pay_in_advance` (Boolean)
- `bill_charges_monthly` (Boolean)
- `bill_fixed_charges_monthly` (Boolean)
- `invoice_display_name` (String)
- `tax_codes` (List of String)
- `metadata` (Map of String)
- `charges` (List of Objects)
  - `billable_metric_id` (String, required)
  - `charge_model` (String, required)
  - `invoiceable` (Boolean)
  - `invoice_display_name` (String)
  - `pay_in_advance` (Boolean)
  - `regroup_paid_fees` (Boolean)
  - `prorated` (Boolean)
  - `min_amount_cents` (Number)
  - `properties_json` (String, JSON payload)
  - `tax_codes` (List of String)
- `minimum_commitment` (Object)
  - `amount_cents` (Number, required)
  - `invoice_display_name` (String)
  - `tax_codes` (List of String)
- `fixed_charges` (List of Objects)
  - `add_on_id` (String)
  - `add_on_code` (String)
  - `charge_model` (String)
  - `invoice_display_name` (String)
  - `pay_in_advance` (Boolean)
  - `prorated` (Boolean)
  - `units` (Number)
  - `properties_json` (String, JSON payload)
  - `tax_codes` (List of String)
- `usage_thresholds` (List of Objects)
  - `amount_cents` (Number)
  - `threshold_display_name` (String)
  - `recurring` (Boolean)
  - `properties_json` (String, JSON payload)
- `entitlements` (List of Objects)
  - `code` (String, required)
  - `name` (String)
  - `description` (String)
  - `recurring` (Boolean)
  - `privileges_json` (String, JSON payload)

### Read-Only

- `id` (String)
- `lago_id` (String)
- `created_at` (String)
- `updated_at` (String)
