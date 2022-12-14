package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestResourceSetsUrlInState(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"http-wait": New("dev")(),
		},
		ExternalProviders: map[string]resource.ExternalProvider{
			"http": {
				VersionConstraint: "2.3.1",
				Source:            "MehdiAtBud/http",
			},
		},
		Steps: []resource.TestStep{
			{
				Config: `
				terraform {
					required_providers {
						http = {
						source = "MehdiAtBud/http"
						version ="2.3.1"
						  }
					}
				  }
				
				resource "http-wait" "example" {
					provider = http
					url = "https://example.com"
				  
					max_elapsed_time = 60
					initial_interval = 100
					multiplier       = "1.2"
					max_interval     = 50000
				  }`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("http-wait.example", "url", "https://example.com"),
				),
			},
		},
	})
}

func TestResourceSNonExistingURL(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"http-wait": New("dev")(),
		},
		ExternalProviders: map[string]resource.ExternalProvider{
			"http": {
				VersionConstraint: "2.3.1",
				Source:            "MehdiAtBud/http",
			},
		},
		Steps: []resource.TestStep{
			{
				Config: `
				terraform {
					required_providers {
						http = {
						source = "MehdiAtBud/http"
						version ="2.3.1"
						  }
					}
				  }
				
				resource "http-wait" "example" {
					provider = http
					url = "https://non-existing.thisisbud.com"
				  
					max_elapsed_time = 60
					initial_interval = 100
					multiplier       = "1.2"
					max_interval     = 50000
				  }`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("http-wait.example", "url", "https://example.com"),
				),
				ExpectError: regexp.MustCompile("no such host"),
			},
		},
	})
}
