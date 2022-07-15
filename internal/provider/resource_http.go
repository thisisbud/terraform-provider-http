package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceUser() *schema.Resource {
	return &schema.Resource{
		Create: CreateUser,
		Update: UpdateUser,
		Read:   ReadUser,
		Delete: DeleteUser,

		Schema: map[string]*schema.Schema{
			"user": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"url": {
				Description: "The URL for the request. Supported schemes are `http` and `https`.",
				Type:        schema.TypeString,
				Required:    true,
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

func CreateUser(d *schema.ResourceData, meta interface{}) error {

	url := d.Get("url").(string)

	request, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {

		return err
	}

	var response *http.Response

	errSummary, errDesc := makeExponentialBackoffRequest(context.Background(),
		request,
		response,
		int64(d.Get("initial_interval").(int)),
		int64(d.Get("max_elapsed_time").(int)),
		int64(d.Get("max_interval").(int)),
		d.Get("randomization_factor").(string),
		d.Get("multiplier").(string),
	)

	if len(errSummary) > 0 {
		return fmt.Errorf("%s : %s", errSummary, errDesc)
	}

	d.SetId(url)

	return nil
}

func UpdateUser(d *schema.ResourceData, meta interface{}) error {

	return nil
}

func ReadUser(d *schema.ResourceData, meta interface{}) error {

	return nil
}

func DeleteUser(d *schema.ResourceData, meta interface{}) error {

	d.SetId("")

	return nil
}
