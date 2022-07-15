package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
func New() tfsdk.Provider {
	return &provider{}
}

var _ tfsdk.Provider = (*provider)(nil)

type provider struct{}

func (p *provider) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{}, nil
}

func (p *provider) Configure(context.Context, tfsdk.ConfigureProviderRequest, *tfsdk.ConfigureProviderResponse) {
}

func (p *provider) GetResources(context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
	return map[string]tfsdk.ResourceType{
		"mysql_user": resourceUser(),
	}, nil
}

func (p *provider) GetDataSources(context.Context) (map[string]tfsdk.DataSourceType, diag.Diagnostics) {
	return map[string]tfsdk.DataSourceType{
		"http": &httpDataSourceType{},
	}, nil
}


*/

func New(version string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			DataSourcesMap: map[string]*schema.Resource{
				"http-wait": dataSourceScaffolding(),
			},
			ResourcesMap: map[string]*schema.Resource{
				"http-wait": resourceUser(),
			},
		}

		p.ConfigureContextFunc = configure(version, p)

		return p
	}
}

type apiClient struct {
	// Add whatever fields, client or connection info, etc. here
	// you would need to setup to communicate with the upstream
	// API.
}

func configure(version string, p *schema.Provider) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
		// Setup a User-Agent for your API client (replace the provider name for yours):
		// userAgent := p.UserAgent("terraform-provider-scaffolding", version)
		// TODO: myClient.UserAgent = userAgent

		return &apiClient{}, nil
	}
}
