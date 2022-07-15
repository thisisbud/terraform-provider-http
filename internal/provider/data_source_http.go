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
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
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
*/

func dataSourceScaffolding() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
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

		ReadContext: Read,

		Schema: map[string]*schema.Schema{
			"url": {
				Description: "The URL for the request. Supported schemes are `http` and `https`.",
				Type:        schema.TypeString,
				Required:    true,
			},

			"request_headers": {
				Description: "A map of request header field names and values.",
				Type:        schema.TypeMap,
				Elem:        schema.TypeString,
				Optional:    true,
			},

			"response_body": {
				Description: "The response body returned as a string.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"response_headers": {
				Description: `A map of response header field names and values.` +
					` Duplicate headers are concatenated according to [RFC2616](https://www.w3.org/Protocols/rfc2616/rfc2616-sec4.html#sec4.2).`,
				Type: schema.TypeMap,
				Elem: schema.TypeString,

				Computed: true,
			},

			"status_code": {
				Description: `The HTTP response status code.`,
				Type:        schema.TypeInt,
				Computed:    true,
			},

			"id": {
				Description: "The ID of this resource.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			"initial_interval": {
				Description: "The initial exponential backoff interval.",
				Type:        schema.TypeInt,
				Optional:    true,
			},

			"max_elapsed_time": {
				Description: "The maximum time to wait for.",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"randomization_factor": {
				Description: "Randomization factor for exponential backoff.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"multiplier": {
				Description: "Multiplier for exponential backoff.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"max_interval": {
				Description: "Maximum interval factor for exponential backoff.",
				Type:        schema.TypeInt,
				Optional:    true,
			},
		},
	}
}

func Read(ctx context.Context, req *schema.ResourceData, meta interface{}) diag.Diagnostics {
	d := diag.Diagnostics{}
	url := req.Get("url").(string)
	headers := req.Get("request_headers").(map[string]interface{})

	request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		d = append(d, diag.Diagnostic{
			Summary: "Error creating request",
			Detail:  fmt.Sprintf("Error creating request: %s", err),
		})
		return d
	}

	for name, value := range headers {
		request.Header.Set(name, value.(string))
	}

	var response *http.Response
	errSummary, errDesc := makeExponentialBackoffRequest(ctx,
		request,
		response,
		req.Get("initial_interval").(int64),
		req.Get("max_elapsed_time").(int64),
		req.Get("max_interval").(int64),
		req.Get("randomization_factor").(string),
		req.Get("multiplier").(string),
	)

	if len(errSummary) > 0 {
		d = append(d, diag.Diagnostic{Summary: errSummary, Detail: errDesc})
		return d
	}

	defer response.Body.Close()

	contentType := response.Header.Get("Content-Type")
	if !isContentTypeText(contentType) {
		d = append(d, diag.Diagnostic{
			Summary: fmt.Sprintf("Content-Type is not recognized as a text type, got %q", contentType),
			Detail:  "If the content is binary data, Terraform may not properly handle the contents of the response.",
		})
		return d
	}

	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {

		d = append(d, diag.Diagnostic{
			Summary: "Error reading response body",
			Detail:  fmt.Sprintf("Error reading response body: %s", err),
		})
		return d
	}

	responseBody := string(bytes)

	responseHeaders := make(map[string]string)
	for k, v := range response.Header {
		// Concatenate according to RFC2616
		// cf. https://www.w3.org/Protocols/rfc2616/rfc2616-sec4.html#sec4.2
		responseHeaders[k] = strings.Join(v, ", ")
	}

	tflog.Info(ctx, fmt.Sprintf("%v", responseBody))

	return nil
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

func makeExponentialBackoffRequest(ctx context.Context, request *http.Request, response *http.Response, initialInterval, maxElapsedTime, maxInterval int64, randomization_fct, multpl string) (string, string) {
	var randomization_factor, multiplier float64
	var err error
	client := &http.Client{}

	randomization_factor = backoff.DefaultRandomizationFactor
	multiplier = backoff.DefaultMultiplier

	if initialInterval == 0 {
		initialInterval = int64(backoff.DefaultInitialInterval)
	}

	if maxElapsedTime == 0 {
		maxElapsedTime = int64(backoff.DefaultMaxElapsedTime)
	}

	if maxInterval == 0 {
		maxInterval = int64(backoff.DefaultMaxInterval)
	}

	if len(randomization_fct) > 0 {
		randomization_factor, err = strconv.ParseFloat(randomization_fct, 64)
		if err != nil {
			return "error converting randomization_factor to float64", fmt.Sprintf("%s", err)
		}
	}

	if len(multpl) > 0 {
		multiplier, err = strconv.ParseFloat(multpl, 64)
		if err != nil {
			return "error converting multiplier to float64", fmt.Sprintf("%s", err)
		}
	}

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = time.Duration(maxElapsedTime) * time.Second
	b.InitialInterval = time.Duration(initialInterval) * time.Millisecond
	b.RandomizationFactor = randomization_factor
	b.Multiplier = multiplier
	b.MaxInterval = time.Duration(maxInterval) * time.Millisecond
	s, err := json.MarshalIndent(b, "", "   ")
	tflog.Info(ctx, fmt.Sprintf("Backoff configuration :  %s", s))

	retries := 0
	err = backoff.Retry(func() error {
		tflog.Info(ctx, "Calling http.Do function")
		response, err = client.Do(request)
		tflog.Info(ctx, fmt.Sprintf("\nNumber of retries %d\n", retries))
		retries++
		return err
	}, b)

	if err != nil {
		return "Error making request", fmt.Sprintf("Error making request: %s", err)
	}

	return "", ""
}
