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

resource "lago_tax" "vat" {
  name = "VAT"
  code = "vat_20"
  rate = 20.0
}

resource "lago_customer" "acme" {
  external_id   = "acme-corp-001"
  name          = "Acme Corporation"
  email         = "billing@acme.example.com"
  phone         = "+1-555-000-0000"
  customer_type = "company"
  currency      = "USD"
  timezone      = "America/New_York"

  address_line1 = "123 Main Street"
  address_line2 = "Suite 400"
  city          = "New York"
  state         = "NY"
  zipcode       = "10001"
  country       = "US"

  legal_name                = "Acme Corporation Inc."
  legal_number              = "12-3456789"
  tax_identification_number = "US123456789"

  net_payment_term             = 30
  finalize_zero_amount_invoice = "skip"

  tax_codes = [lago_tax.vat.code]

  billing_configuration {
    invoice_grace_period  = 3
    payment_provider      = "stripe"
    payment_provider_code = "stripe_us"
    provider_customer_id  = "cus_abc123"
    sync_with_provider    = true
    document_locale       = "en"

    provider_payment_methods = ["card", "us_bank_account"]
  }

  shipping_address {
    address_line1 = "456 Warehouse Blvd"
    city          = "Newark"
    state         = "NJ"
    zipcode       = "07101"
    country       = "US"
  }

  metadata {
    key                = "salesforce_id"
    value              = "0015000000XXXXAAA"
    display_in_invoice = false
  }

  metadata {
    key                = "plan_tier"
    value              = "enterprise"
    display_in_invoice = true
  }
}

variable "lago_api_endpoint" {
  type = string
}

variable "lago_api_key" {
  type      = string
  sensitive = true
}
