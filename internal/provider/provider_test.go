package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

//nolint:unparam
var providerFactories = map[string]func() (*schema.Provider, error){
	"http-wait": func() (*schema.Provider, error) {
		return New("dev")(), nil
	},
}
