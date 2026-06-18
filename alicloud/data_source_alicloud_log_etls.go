// Package alicloud. This file is generated automatically. Please do not modify it manually, thank you!
package alicloud

import (
	"fmt"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceAliCloudLogETLs() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAliCloudLogETLRead,
		Schema: map[string]*schema.Schema{
			"project_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"logstore_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"etls": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"configuration": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"script": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"to_time": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									// lintignore: S006
									"parameters": {
										Type:     schema.TypeMap,
										Computed: true,
									},
									"sink": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"datasets": {
													Type:     schema.TypeList,
													Computed: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"project_name": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"endpoint": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"logstore_name": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"role_arn": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"name": {
													Type:     schema.TypeString,
													Computed: true,
												},
											},
										},
									},
									"logstore_name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"lang": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"from_time": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"role_arn": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"create_time": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"display_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"job_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"last_modified_time": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"schedule_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceAliCloudLogETLRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapError(err)
	}

	project := d.Get("project_name").(string)
	logstore := d.Get("logstore_name").(string)
	etls, err := slsService.ListSlsETLs(project, "", logstore)
	if err != nil {
		return WrapError(err)
	}

	ids := make([]string, 0)
	s := make([]map[string]interface{}, 0)
	for _, etl := range etls {
		if etl == nil {
			continue
		}

		mapping := map[string]interface{}{}
		mapping["id"] = fmt.Sprintf("%s:%s", project, etl.Name)

		mapping["create_time"] = int(etl.CreateTime)
		mapping["description"] = etl.Description
		mapping["display_name"] = etl.DisplayName
		mapping["last_modified_time"] = int(etl.CreateTime)
		mapping["status"] = etl.Status
		mapping["job_name"] = etl.Name
		if etl.Schedule != nil {
			mapping["schedule_id"] = etl.Schedule.Type
		} else {
			mapping["schedule_id"] = ""
		}

		configurationMaps := make([]map[string]interface{}, 0)
		if etl.Configuration != nil {
			configurationMap := make(map[string]interface{})
			configurationMap["from_time"] = int(etl.Configuration.FromTime)
			configurationMap["lang"] = ""
			configurationMap["logstore_name"] = etl.Configuration.Logstore
			configurationMap["parameters"] = convertETLParametersToMap(etl.Configuration.Parameters)
			configurationMap["role_arn"] = etl.Configuration.RoleArn
			configurationMap["script"] = etl.Configuration.Script
			configurationMap["to_time"] = int(etl.Configuration.ToTime)

			sinkMaps := make([]map[string]interface{}, 0)
			for _, sink := range etl.Configuration.Sinks {
				sinkMap := make(map[string]interface{})
				sinkMap["endpoint"] = ""
				sinkMap["logstore_name"] = sink.Logstore
				sinkMap["name"] = sink.Name
				sinkMap["project_name"] = sink.Project
				sinkMap["role_arn"] = sink.RoleArn
				sinkMap["datasets"] = []interface{}{}
				sinkMaps = append(sinkMaps, sinkMap)
			}
			configurationMap["sink"] = sinkMaps
			configurationMaps = append(configurationMaps, configurationMap)
		}
		mapping["configuration"] = configurationMaps

		ids = append(ids, fmt.Sprint(mapping["id"]))
		s = append(s, mapping)
	}

	d.SetId(dataResourceIdHash(ids))
	if err := d.Set("etls", s); err != nil {
		return WrapError(err)
	}
	return nil
}

func convertETLParametersToMap(parameters []string) map[string]string {
	result := make(map[string]string)
	for idx, parameter := range parameters {
		if parameter == "" {
			continue
		}
		for i := 0; i < len(parameter); i++ {
			if parameter[i] == '=' {
				result[parameter[:i]] = parameter[i+1:]
				goto next
			}
		}
		result[fmt.Sprintf("param_%d", idx)] = parameter
	next:
	}
	return result
}
