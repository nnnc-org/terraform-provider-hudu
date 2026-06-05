terraform {
  required_providers {
    hudu = {
      source  = "nnnc-org/hudu"
      version = "~> 0.1"
    }
  }
}

# host and api_key fall back to HUDU_API_HOST / HUDU_API_KEY.
provider "hudu" {
  host    = "https://hudu.example.com"
  api_key = var.hudu_api_key
}

variable "hudu_api_key" {
  type      = string
  sensitive = true
}
