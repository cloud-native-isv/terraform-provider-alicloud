package alicloud

import (
	"regexp"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceAliCloudAdbpgInstancePlans() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAliCloudAdbpgInstancePlansRead,
		Schema: map[string]*schema.Schema{
			"db_instance_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.ValidateRegexp,
			},
			"output_file": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"plans": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"plan_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"plan_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"plan_status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"schedule_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceAliCloudAdbpgInstancePlansRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	adbpgService, err := NewAdbpgService(client)
	if err != nil {
		return WrapError(err)
	}

	instanceId := d.Get("db_instance_id").(string)
	allPlans, err := adbpgService.ListAdbpgInstancePlans(instanceId)
	if err != nil {
		return WrapError(err)
	}

	var nameRegex *regexp.Regexp
	if v, ok := d.GetOk("name_regex"); ok {
		nameRegex = regexp.MustCompile(v.(string))
	}

	var ids []string
	s := make([]map[string]interface{}, 0)

	for _, plan := range allPlans {
		if nameRegex != nil && !nameRegex.MatchString(plan.PlanName) {
			continue
		}
		mapping := map[string]interface{}{
			"plan_id":       plan.PlanId,
			"plan_name":     plan.PlanName,
			"plan_status":   plan.PlanStatus,
			"schedule_type": plan.ScheduleType,
		}
		ids = append(ids, plan.PlanId)
		s = append(s, mapping)
	}

	d.SetId(dataResourceIdHash(ids))
	d.Set("plans", s)

	if output, ok := d.GetOk("output_file"); ok && output.(string) != "" {
		writeToFile(output.(string), s)
	}

	return nil
}
