package alicloud

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAlicloudLogtailConfig() *schema.Resource {
	return &schema.Resource{
		Create: resourceAlicloudLogtailConfigCreate,
		Read:   resourceAlicloudLogtailConfigRead,
		Update: resourceAlicloudLogtailConfiglUpdate,
		Delete: resourceAlicloudLogtailConfigDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"input_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"file", "plugin"}, false),
			},
			"log_sample": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"last_modify_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"project": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"logstore": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"output_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"SlsService"}, false),
			},
			"input_detail": {
				Type:     schema.TypeString,
				Required: true,
				StateFunc: func(v interface{}) string {
					yaml, _ := normalizeJsonString(v)
					return yaml
				},
				ValidateFunc: validation.ValidateJsonString,
			},
		},
	}
}

func resourceAlicloudLogtailConfigCreate(d *schema.ResourceData, meta interface{}) error {
	// TODO: Logtail configuration management is not yet fully implemented in the new SLS API
	// This resource needs to be updated to use the new API methods when they become available
	return WrapError(fmt.Errorf("logtail configuration management is temporarily unavailable during API migration"))
}

func resourceAlicloudLogtailConfigRead(d *schema.ResourceData, meta interface{}) error {
	// TODO: Logtail configuration management is not yet fully implemented in the new SLS API
	d.SetId("")
	return nil
}

func resourceAlicloudLogtailConfiglUpdate(d *schema.ResourceData, meta interface{}) error {
	// TODO: Logtail configuration management is not yet fully implemented in the new SLS API
	return WrapError(fmt.Errorf("logtail configuration management is temporarily unavailable during API migration"))
}

func resourceAlicloudLogtailConfigDelete(d *schema.ResourceData, meta interface{}) error {
	// TODO: Logtail configuration management is not yet fully implemented in the new SLS API
	return nil
}
