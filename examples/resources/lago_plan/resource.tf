terraform {
  required_version = ">= 1.11.0"

  required_providers {
    lago = {
      source  = "registry.opentofu.org/rpcpool/lago"
      version = ">= 0.1.0"
    }
  }
}

provider "lago" {
  api_endpoint = var.lago_api_endpoint
  api_key      = var.lago_api_key
}

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

variable "lago_api_endpoint" {
  type = string
}

variable "lago_api_key" {
  type      = string
  sensitive = true
}
