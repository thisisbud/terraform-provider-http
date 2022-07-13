package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ tfsdk.DataSourceType = (*httpDataSourceType)(nil)

type httpDataSourceType struct{}

func (d *httpDataSourceType) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: `
The ` + "`http`" + ` data source makes an HTTP GET request to the given URL and exports
information about the response.

The given URL may be either an ` + "`http`" + ` or ` + "`https`" + ` URL. At present this resource
can only retrieve data from URLs that respond with ` + "`text/*`" + ` or
` + "`application/json`" + ` content types, and expects the result to be UTF-8 encoded
regardless of the returned content type header.

~> **Important** Although ` + "`https`" + ` URLs can be used, there is currently no
mechanism to authenticate the remote server except for general verification of
the server certificate's chain of trust. Data retrieved from servers not under
your control should be treated as untrustworthy.

In addition to this there is possibility to configure exponential backoff retries that can be bounded
both by max elapsed time and max interval between retries.`,

		Attributes: map[string]tfsdk.Attribute{
			"url": {
				Description: "The URL for the request. Supported schemes are `http` and `https`.",
				Type:        types.StringType,
				Required:    true,
			},

			"request_headers": {
				Description: "A map of request header field names and values.",
				Type: types.MapType{
					ElemType: types.StringType,
				},
				Optional: true,
			},

			"response_body": {
				Description: "The response body returned as a string.",
				Type:        types.StringType,
				Computed:    true,
			},

			"response_headers": {
				Description: `A map of response header field names and values.` +
					` Duplicate headers are concatenated according to [RFC2616](https://www.w3.org/Protocols/rfc2616/rfc2616-sec4.html#sec4.2).`,
				Type: types.MapType{
					ElemType: types.StringType,
				},
				Computed: true,
			},

			"status_code": {
				Description: `The HTTP response status code.`,
				Type:        types.Int64Type,
				Computed:    true,
			},

			"id": {
				Description: "The ID of this resource.",
				Type:        types.StringType,
				Computed:    true,
			},

			"initial_interval": {
				Description: "The initial exponential backoff interval.",
				Type:        types.Int64Type,
				Optional:    true,
			},

			"max_elapsed_time": {
				Description: "The maximum time to wait for.",
				Type:        types.Int64Type,
				Optional:    true,
			},
			"randomization_factor": {
				Description: "Randomization factor for exponential backoff.",
				Type:        types.StringType,
				Optional:    true,
			},
			"multiplier": {
				Description: "Multiplier for exponential backoff.",
				Type:        types.StringType,
				Optional:    true,
			},
			"max_interval": {
				Description: "Maximum interval factor for exponential backoff.",
				Type:        types.Int64Type,
				Optional:    true,
			},
		},
	}, nil
}

func (d *httpDataSourceType) NewDataSource(context.Context, tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	return &httpDataSource{}, nil
}

var _ tfsdk.DataSource = (*httpDataSource)(nil)

type httpDataSource struct{}

func (d *httpDataSource) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	var model modelV0
	diags := req.Config.Get(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := model.URL.Value
	headers := model.RequestHeaders

	client := &http.Client{}

	request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating request",
			fmt.Sprintf("Error creating request: %s", err),
		)
		return
	}

	for name, value := range headers.Elems {
		var header string
		diags = tfsdk.ValueAs(ctx, value, &header)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		request.Header.Set(name, header)
	}

	var randomization_factor, multiplier float64

	randomization_factor = backoff.DefaultRandomizationFactor
	multiplier = backoff.DefaultMultiplier

	if model.InitialInterval.Value == 0 {
		model.InitialInterval.Value = int64(backoff.DefaultInitialInterval)
	}

	if model.MaxElapsedTime.Value == 0 {
		model.MaxElapsedTime.Value = int64(backoff.DefaultMaxElapsedTime)
	}

	if model.MaxElapsedTime.Value == 0 {
		model.MaxInterval.Value = int64(backoff.DefaultMaxElapsedTime)
	}

	if len(model.RandomizationFactor.Value) > 0 {
		randomization_factor, err = strconv.ParseFloat(model.RandomizationFactor.Value, 64)
		if err != nil {
			resp.Diagnostics.AddError(
				"error converting string to float64",
				fmt.Sprintf("%s", err),
			)
			return
		}
	}

	if len(model.Multiplier.Value) > 0 {
		multiplier, err = strconv.ParseFloat(model.Multiplier.Value, 64)
		if err != nil {
			resp.Diagnostics.AddError(
				"error converting string to float64",
				fmt.Sprintf("%s", err),
			)
			return
		}
	}

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = time.Duration(model.MaxElapsedTime.Value) * time.Second
	b.InitialInterval = time.Duration(model.InitialInterval.Value) * time.Millisecond
	b.RandomizationFactor = randomization_factor
	b.Multiplier = multiplier
	b.MaxInterval = time.Duration(model.MaxInterval.Value) * time.Millisecond
	s, err := json.MarshalIndent(b, "", "   ")
	tflog.Info(ctx, fmt.Sprintf("Backoff configuration :  %s", s))

	var response *http.Response
	retries := 0
	err = backoff.Retry(func() error {
		tflog.Info(ctx, "Calling http.Do function")
		response, err = client.Do(request)
		tflog.Info(ctx, fmt.Sprintf("Number of retries %d", retries))
		retries++
		return err
	}, b)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error making request",
			fmt.Sprintf("Error making request: %s", err),
		)
		return
	}

	defer response.Body.Close()

	contentType := response.Header.Get("Content-Type")
	if !isContentTypeText(contentType) {
		resp.Diagnostics.AddWarning(
			fmt.Sprintf("Content-Type is not recognized as a text type, got %q", contentType),
			"If the content is binary data, Terraform may not properly handle the contents of the response.",
		)
	}

	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading response body",
			fmt.Sprintf("Error reading response body: %s", err),
		)
		return
	}

	responseBody := string(bytes)

	responseHeaders := make(map[string]string)
	for k, v := range response.Header {
		// Concatenate according to RFC2616
		// cf. https://www.w3.org/Protocols/rfc2616/rfc2616-sec4.html#sec4.2
		responseHeaders[k] = strings.Join(v, ", ")
	}

	respHeadersState := types.Map{}

	diags = tfsdk.ValueFrom(ctx, responseHeaders, types.Map{ElemType: types.StringType}.Type(ctx), &respHeadersState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	model.ID = types.String{Value: url}
	model.ResponseHeaders = respHeadersState
	model.ResponseBody = types.String{Value: responseBody}
	model.StatusCode = types.Int64{Value: int64(response.StatusCode)}

	diags = resp.State.Set(ctx, model)
	resp.Diagnostics.Append(diags...)
}

// This is to prevent potential issues w/ binary files
// and generally unprintable characters
// See https://github.com/hashicorp/terraform/pull/3858#issuecomment-156856738
func isContentTypeText(contentType string) bool {

	parsedType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return false
	}

	allowedContentTypes := []*regexp.Regexp{
		regexp.MustCompile("^text/.+"),
		regexp.MustCompile("^application/json$"),
		regexp.MustCompile(`^application/samlmetadata\+xml`),
	}

	for _, r := range allowedContentTypes {
		if r.MatchString(parsedType) {
			charset := strings.ToLower(params["charset"])
			return charset == "" || charset == "utf-8" || charset == "us-ascii"
		}
	}

	return false
}

type modelV0 struct {
	ID                  types.String `tfsdk:"id"`
	URL                 types.String `tfsdk:"url"`
	RequestHeaders      types.Map    `tfsdk:"request_headers"`
	ResponseHeaders     types.Map    `tfsdk:"response_headers"`
	ResponseBody        types.String `tfsdk:"response_body"`
	StatusCode          types.Int64  `tfsdk:"status_code"`
	InitialInterval     types.Int64  `tfsdk:"initial_interval"`
	MaxElapsedTime      types.Int64  `tfsdk:"max_elapsed_time"`
	RandomizationFactor types.String `tfsdk:"randomization_factor"`
	Multiplier          types.String `tfsdk:"multiplier"`
	MaxInterval         types.Int64  `tfsdk:"max_interval"`
}
