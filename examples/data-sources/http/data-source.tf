terraform {
  required_providers {
    http = {
      source  = "MehdiAtBud/http"
      version = "2.2.16"
    }
  }
}

# The following example shows how to issue an HTTP GET request supplying
# an optional request header.
data "http" "example" {
  url = "https://checkpoint-api.hashicorp.com/v1/check/terraform"

  # Optional request headers
  request_headers = {
    Accept = "application/json"
  }

  max_elapsed_time = 10
  initial_interval = 100
  multiplier       = "1.2"
  max_interval     = 50000
}