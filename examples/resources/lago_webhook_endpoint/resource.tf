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

resource "lago_webhook_endpoint" "main" {
  webhook_url    = "https://example.com/webhooks/lago"
  signature_algo = "hmac"
}

variable "lago_api_endpoint" {
  type = string
}

variable "lago_api_key" {
  type      = string
  sensitive = true
}
