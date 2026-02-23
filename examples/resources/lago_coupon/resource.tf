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

# Fixed-amount coupon with a time limit
resource "lago_coupon" "fixed_discount" {
  name        = "New Customer Discount"
  code        = "new_customer_50"
  description = "50 USD off for new customers"

  coupon_type     = "fixed_amount"
  amount_cents    = 5000
  amount_currency = "USD"

  expiration    = "time_limit"
  expiration_at = "2025-12-31T23:59:59Z"

  frequency = "once"
  reusable  = false
}

# Percentage coupon with recurring frequency limited to specific plans
resource "lago_coupon" "percentage_recurring" {
  name = "Loyalty 10% Off"
  code = "loyalty_10pct"

  coupon_type     = "percentage"
  percentage_rate = 10.0

  expiration = "no_expiration"

  frequency          = "recurring"
  frequency_duration = 3
  reusable           = true

  plan_codes = ["starter_plan", "pro_plan"]
}

variable "lago_api_endpoint" {
  type = string
}

variable "lago_api_key" {
  type      = string
  sensitive = true
}
