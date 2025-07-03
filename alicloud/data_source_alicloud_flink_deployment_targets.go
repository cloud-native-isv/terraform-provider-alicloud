package alicloud

import (
	"regexp"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceAliCloudFlinkDeploymentTargets() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAliCloudFlinkDeploymentTargetsRead,

		Schema: map[string]*schema.Schema{
			"workspace_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"namespace_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.ValidateRegexp,
				ForceNew:     true,
			},
			"names": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"ids": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"output_file": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"targets": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"namespace_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"quota": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"limit": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"cpu": {
													Type:     schema.TypeFloat,
													Computed: true,
												},
												"memory_gb": {
													Type:     schema.TypeFloat,
													Computed: true,
												},
												"disk": {
													Type:     schema.TypeInt,
													Computed: true,
												},
											},
										},
									},
									"used": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"cpu": {
													Type:     schema.TypeFloat,
													Computed: true,
												},
												"memory_gb": {
													Type:     schema.TypeFloat,
													Computed: true,
												},
												"disk": {
													Type:     schema.TypeInt,
													Computed: true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func dataSourceAliCloudFlinkDeploymentTargetsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	deploymentTargetService := NewFlinkDeploymentTargetService(client)

	workspaceId := d.Get("workspace_id").(string)
	namespaceName := d.Get("namespace_name").(string)

	objects, err := deploymentTargetService.DescribeFlinkDeploymentTargets(workspaceId, namespaceName)
	if err != nil {
		return WrapError(err)
	}

	var filteredTargets []flink.DeploymentTarget
	var targetNames []string
	var ids []string

	nameRegex, ok := d.GetOk("name_regex")
	if ok && nameRegex.(string) != "" {
		r, err := regexp.Compile(nameRegex.(string))
		if err != nil {
			return WrapError(err)
		}
		for _, target := range objects {
			if r.MatchString(target.Name) {
				filteredTargets = append(filteredTargets, target)
			}
		}
	} else {
		filteredTargets = objects
	}

	for _, target := range filteredTargets {
		targetNames = append(targetNames, target.Name)
		ids = append(ids, formatDeploymentTargetId(workspaceId, namespaceName, target.Name))
	}

	d.SetId(dataResourceIdHash(ids))
	d.Set("ids", ids)
	d.Set("names", targetNames)

	targets := make([]map[string]interface{}, 0)
	for _, target := range filteredTargets {
		mapping := map[string]interface{}{
			"id":        formatDeploymentTargetId(workspaceId, namespaceName, target.Name),
			"name":      target.Name,
			"namespace_name": target.Namespace,
		}

		if target.Quota != nil {
			mapping["quota"] = flattenResourceQuota(target.Quota)
		}

		targets = append(targets, mapping)
	}

	d.Set("targets", targets)

	if output, ok := d.GetOk("output_file"); ok && output.(string) != "" {
		writeToFile(output.(string), targets)
	}

	return nil
}
