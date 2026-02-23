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

resource "lago_subscription" "acme_pro" {
  external_id          = "acme-corp-001-pro"
  external_customer_id = lago_customer.acme.external_id
  plan_code            = "pro_plan"
  name                 = "Acme Pro Subscription"
  billing_time         = "anniversary"

  on_termination_credit_note = "credit"
  on_termination_invoice     = "generate"
}

variable "lago_api_endpoint" {
  type = string
}

variable "lago_api_key" {
  type      = string
  sensitive = true
}
