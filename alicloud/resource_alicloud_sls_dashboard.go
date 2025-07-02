package alicloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAliCloudLogDashboard() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudLogDashboardCreate,
		Read:   resourceAliCloudLogDashboardRead,
		Update: resourceAliCloudLogDashboardUpdate,
		Delete: resourceAliCloudLogDashboardDelete,
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
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: chartListDiffSuppress,
			},
			"attribute": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringIsJSON,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					equal, _ := compareJsonTemplateAreEquivalent(old, new)
					return equal
				},
			},
		},
	}
}

func resourceAliCloudLogDashboardCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)

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
	return resourceAliCloudLogDashboardRead(d, meta)
}

func resourceAliCloudLogDashboardRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	logService, err := NewSlsService(client)
	if err != nil {
		return WrapError(err)
	}
	parts, err := ParseResourceId(d.Id(), 2)
	if err != nil {
		return WrapError(err)
	}
	dashboard, err := logService.DescribeSlsDashboard(parts[0], parts[1])
	if err != nil {
		if NotFoundError(err) {
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	d.Set("project_name", parts[0])
	d.Set("dashboard_name", dashboard.Name)
	d.Set("display_name", dashboard.DisplayName)

	if dashboard.Attribute != nil && len(dashboard.Attribute) > 0 {
		if attributeBytes, err := json.Marshal(dashboard.Attribute); err == nil {
			d.Set("attribute", string(attributeBytes))
		} else {
			return WrapError(err)
		}
	}

	if dashboard.ChartList != nil && len(dashboard.ChartList) > 0 {
		// Convert chart list to the expected format
		var charts []interface{}
		for _, chart := range dashboard.ChartList {
			chartMap := make(map[string]interface{})
			chartMap["title"] = chart.Title
			chartMap["type"] = chart.Type

			// Handle search configuration
			if chart.Search != nil {
				search := make(map[string]interface{})
				search["logstore"] = chart.Search.Logstore
				search["topic"] = chart.Search.Topic
				search["query"] = chart.Search.Query
				search["start"] = chart.Search.Start
				search["end"] = chart.Search.End
				search["timeSpanType"] = chart.Search.TimeSpanType
				chartMap["search"] = search
			}

			// Handle display configuration
			if chart.Display != nil {
				display := make(map[string]interface{})
				display["xPos"] = chart.Display.XPos
				display["yPos"] = chart.Display.YPos
				display["width"] = chart.Display.Width
				display["height"] = chart.Display.Height
				display["displayName"] = chart.Display.DisplayName
				chartMap["display"] = display
			}

			charts = append(charts, chartMap)
		}

		if charlistBytes, err := json.Marshal(charts); err == nil {
			d.Set("char_list", string(charlistBytes))
		} else {
			return WrapError(err)
		}
	}

	return nil
}

func resourceAliCloudLogDashboardUpdate(d *schema.ResourceData, meta interface{}) error {
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
	return resourceAliCloudLogDashboardRead(d, meta)
}

func resourceAliCloudLogDashboardDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	logService, err := NewSlsService(client)
	if err != nil {
		return WrapError(err)
	}
	parts, err := ParseResourceId(d.Id(), 2)
	if err != nil {
		return WrapError(err)
	}
	err = resource.Retry(3*time.Minute, func() *resource.RetryError {
		err := logService.DeleteDashboard(parts[0], parts[1])
		if err != nil {
			if IsExpectedErrors(err, []string{LogClientTimeout, "RequestTimeout"}) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		addDebug("DeleteDashboard", nil, nil, map[string]interface{}{
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
	return WrapError(logService.WaitForSlsDashboard(d.Id(), Deleted, DefaultTimeout))
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
