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

resource "lago_add_on" "setup_fee" {
  name            = "Setup Fee"
  code            = "setup_fee"
  description     = "One-time setup fee charged at onboarding"
  amount_cents    = 10000
  amount_currency = "USD"

  invoice_display_name = "Onboarding Setup Fee"

  tax_codes = [lago_tax.vat.code]
}

variable "lago_api_endpoint" {
  type = string
}

variable "lago_api_key" {
  type      = string
  sensitive = true
}
