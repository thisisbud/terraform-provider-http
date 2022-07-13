terraform {
  required_providers {
    http = {
      source  = "MehdiAtBud/http"
      version = "2.2.15"
    }
  }
}

# The following example shows how to issue an HTTP GET request supplying
# an optional request header.
data "http" "example" {
  url = "https://checkpoin-api.hashicorp.com/v1/check/terraform"

  # Optional request headers
  request_headers = {
    Accept = "application/json"
  }

  max_elapsed_time = 10
  initial_interval = 100
  multiplier       = 1.2
  max_interval     = 500000000000000000
}




provider "http" {
  # Configuration options
}