terraform {
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

resource "lago_customer" "acme" {
  external_id = "acme-corp-001"
  name        = "Acme Corp"
  email       = "billing@acme.example.com"
}

resource "lago_wallet" "acme_credits" {
  external_customer_id = lago_customer.acme.external_id
  name                 = "Acme Prepaid Credits"
  currency             = "USD"
  rate_amount          = "1.0"

  # Credits purchased and granted at creation time.
  # These fields are write-only and are not returned by the API.
  paid_credits    = "500.0"
  granted_credits = "50.0"

  invoice_requires_successful_payment = false

  # Optional: auto-top-up the wallet every month.
  recurring_transaction_rules = [
    {
      interval        = "monthly"
      method          = "fixed"
      trigger         = "interval"
      paid_credits    = "100.0"
      granted_credits = "10.0"
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
