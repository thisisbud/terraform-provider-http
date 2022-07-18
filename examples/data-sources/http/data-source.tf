terraform {
  required_providers {
    http = {
      source  = "MehdiAtBud/http"
      version = "2.2.27"
    }
  }
}

data "http-wait" "example" {
  provider = http

  url = "https://checkpoint-api.hashicorp.com/v1/check/terraform"

  # Optional request headers
  request_headers = {
    Accept = "application/json"
  }

  max_elapsed_time     = 10
  initial_interval     = 100
  multiplier           = "1.2"
  max_interval         = 50000
  randomization_factor = 3
}


resource "http-wait" "example" {
  provider = http
  url      = "https://example.com"

  max_elapsed_time = 60
  initial_interval = 100
  multiplier       = "1.2"
  max_interval     = 50000
}