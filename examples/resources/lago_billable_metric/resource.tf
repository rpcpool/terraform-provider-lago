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

resource "lago_billable_metric" "api_requests" {
  name             = "API Requests"
  code             = "api_requests"
  aggregation_type = "count_agg"
  description      = "Total number of API requests"
  recurring        = true
}

variable "lago_api_endpoint" {
  type = string
}

variable "lago_api_key" {
  type      = string
  sensitive = true
}
