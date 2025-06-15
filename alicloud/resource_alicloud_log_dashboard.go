package alicloud

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	sls "github.com/aliyun/aliyun-log-go-sdk"
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunSlsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/sls"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAlicloudLogDashboard() *schema.Resource {
	return &schema.Resource{
		Create: resourceAlicloudLogDashboardCreate,
		Read:   resourceAlicloudLogDashboardRead,
		Update: resourceAlicloudLogDashboardUpdate,
		Delete: resourceAlicloudLogDashboardDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"project_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"dashboard_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"display_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"char_list": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validation.ValidateJsonString,
				DiffSuppressFunc: chartListDiffSuppress,
			},
			"attribute": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.ValidateJsonString,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					equal, _ := compareJsonTemplateAreEquivalent(old, new)
					return equal
				},
			},
		},
	}
}

func resourceAlicloudLogDashboardCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	logService := NewLogService(client)

	dashboard := map[string]interface{}{
		"dashboardName": d.Get("dashboard_name").(string),
		"displayName":   d.Get("display_name").(string),
	}

	if v, ok := d.GetOk("attribute"); ok {
		attribute := map[string]interface{}{}
		if err := json.Unmarshal([]byte(v.(string)), &attribute); err != nil {
			return WrapError(err)
		}
		dashboard["attribute"] = attribute
	}

	chartList := []interface{}{}
	jsonErr := json.Unmarshal([]byte(d.Get("char_list").(string)), &chartList)
	if jsonErr != nil {
		return WrapError(jsonErr)
	}
	dashboard["charts"] = chartList
	dashboardBytes, err := json.Marshal(dashboard)
	if err != nil {
		return WrapError(err)
	}
	dashboardStr := string(dashboardBytes)

	if err := resource.Retry(2*time.Minute, func() *resource.RetryError {
		// Create dashboard using RPC call directly since CWS-Lib-Go doesn't have CreateDashboardString method
		_, err := client.RpcPost("sls", "2020-12-30", "CreateDashboard", nil, map[string]interface{}{
			"project":   d.Get("project_name").(string),
			"dashboard": dashboardStr,
		}, true)
		if err != nil {
			if IsExpectedErrors(err, []string{LogClientTimeout}) {
				time.Sleep(5 * time.Second)
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		addDebug("CreateDashboard", dashboard, nil, map[string]interface{}{
			"dashBoard": dashboard,
		})
		d.SetId(fmt.Sprintf("%s%s%s", d.Get("project_name").(string), COLON_SEPARATED, d.Get("dashboard_name").(string)))
		return nil
	}); err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_log_dashboard", "CreateDashboard", AliyunLogGoSdkERROR)
	}
	return resourceAlicloudLogDashboardRead(d, meta)
}

func resourceAlicloudLogDashboardRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	logService := NewLogService(client)
	parts, err := ParseResourceId(d.Id(), 2)
	if err != nil {
		return WrapError(err)
	}
	object, err := logService.DescribeLogDashboard(d.Id())
	if err != nil {
		if NotFoundError(err) {
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	// Handle response based on the actual return type from the new service
	var objectBytes []byte
	var dashboard map[string]interface{}

	// Try to handle different response types
	if objectStr, ok := object["content"].(string); ok {
		objectBytes = []byte(objectStr)
	} else {
		// Direct conversion of object to JSON bytes
		if objBytes, err := json.Marshal(object); err == nil {
			objectBytes = objBytes
		} else {
			return WrapError(err)
		}
	}

	err = json.Unmarshal(objectBytes, &dashboard)
	if err != nil {
		return WrapError(err)
	}

	d.Set("project_name", parts[0])
	d.Set("dashboard_name", dashboard["dashboardName"])
	d.Set("display_name", dashboard["displayName"])
	if v, ok := dashboard["attribute"]; ok {
		if attributeBytes, err := json.Marshal(v); err == nil {
			d.Set("attribute", string(attributeBytes))
		} else {
			return WrapError(err)
		}
	}

	if charts, ok := dashboard["charts"].([]interface{}); ok {
		for _, v := range charts {
			if chartMap, isMap := v.(map[string]interface{}); isMap {
				if action, actionOK := chartMap["action"]; actionOK {
					if action == nil {
						delete(chartMap, "action")
					}
				}
			}
		}
		charlist, err := json.Marshal(charts)
		if err != nil {
			return WrapError(err)
		}
		d.Set("char_list", string(charlist))
	}

	return nil
}

func resourceAlicloudLogDashboardUpdate(d *schema.ResourceData, meta interface{}) error {
	parts, err := ParseResourceId(d.Id(), 2)
	if err != nil {
		return WrapError(err)
	}

	update := false
	if d.HasChange("display_name") {
		update = true
	}
	if d.HasChange("char_list") {
		update = true
	}
	if d.HasChange("attribute") {
		update = true
	}

	if update {
		client := meta.(*connectivity.AliyunClient)
		dashboard := map[string]interface{}{
			"dashboardName": parts[1],
			"displayName":   d.Get("display_name").(string),
		}
		if v, ok := d.GetOk("attribute"); ok {
			attribute := map[string]interface{}{}
			if err := json.Unmarshal([]byte(v.(string)), &attribute); err != nil {
				return WrapError(err)
			}
			dashboard["attribute"] = attribute
		}
		chartList := []interface{}{}
		jsonErr := json.Unmarshal([]byte(d.Get("char_list").(string)), &chartList)
		if jsonErr != nil {
			return WrapError(jsonErr)
		}
		dashboard["charts"] = chartList
		dashboardBytes, err := json.Marshal(dashboard)
		if err != nil {
			return WrapError(err)
		}
		dashboardStr := string(dashboardBytes)

		// Update dashboard using RPC call directly since CWS-Lib-Go doesn't have UpdateDashboardString method
		_, err = client.RpcPost("sls", "2020-12-30", "UpdateDashboard", nil, map[string]interface{}{
			"project":   parts[0],
			"dashboard": dashboardStr,
		}, true)
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "UpdateDashboard", AliyunLogGoSdkERROR)
		}
	}
	return resourceAlicloudLogDashboardRead(d, meta)
}

func resourceAlicloudLogDashboardDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	logService := NewLogService(client)
	parts, err := ParseResourceId(d.Id(), 2)
	if err != nil {
		return WrapError(err)
	}
	var requestInfo *sls.Client
	err = resource.Retry(3*time.Minute, func() *resource.RetryError {
		raw, err := client.WithSlsAPIClient(func(slsClient *aliyunSlsAPI.SlsAPI) (interface{}, error) {
			ctx := context.Background()
			requestInfo = slsClient
			return nil, slsClient.DeleteDashboard(parts[0], parts[1])
		})
		if err != nil {
			if IsExpectedErrors(err, []string{LogClientTimeout, "RequestTimeout"}) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		addDebug("DeleteDashboard", raw, requestInfo, map[string]interface{}{
			"project_name": parts[0],
			"dashboard":    parts[1],
		})
		return nil
	})
	if err != nil {
		if IsExpectedErrors(err, []string{"DashboardNotExist", "ProjectNotExist"}) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteDashboard", AliyunLogGoSdkERROR)
	}
	return WrapError(logService.WaitForLogDashboard(d.Id(), Deleted, DefaultTimeout))
}

func chartListDiffSuppress(k, old, new string, d *schema.ResourceData) bool {
	if old == "" && new == "" {
		return true
	}

	obj1 := []map[string]interface{}{}
	err := json.Unmarshal([]byte(old), &obj1)
	if err != nil {
		return false
	}
	canonicalJson1, _ := json.Marshal(obj1)

	obj2 := []map[string]interface{}{}
	err = json.Unmarshal([]byte(new), &obj2)
	if err != nil {
		return false
	}
	canonicalJson2, _ := json.Marshal(obj2)

	equal := bytes.Equal(canonicalJson1, canonicalJson2)
	if !equal {
		log.Printf("[DEBUG] Canonical template are not equal.\nFirst: %s\nSecond: %s\n",
			canonicalJson1, canonicalJson2)
	}
	return equal
}
