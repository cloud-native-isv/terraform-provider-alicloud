package alicloud

import (
	"fmt"
	"regexp"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/kafka"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceAliCloudAlikafkaInstances() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAliCloudAlikafkaInstancesRead,

		Schema: map[string]*schema.Schema{
			"ids": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
				ForceNew: true,
			},
			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.ValidateRegexp,
				ForceNew:     true,
			},
			"output_file": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			// Computed values
			"names": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"enable_details": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"instances": {
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
						"create_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"service_status": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"deploy_type": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"vpc_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"vswitch_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"io_max": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"eip_max": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"disk_type": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"disk_size": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"topic_quota": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"partition_num": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"paid_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"spec_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"zone_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"end_point": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"security_group": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"service_version": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"config": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"expired_time": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"msg_retain": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"ssl_end_point": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"upgrade_service_detail_info": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"current2_open_source_version": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"allowed_list": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"deploy_type": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"vpc_list": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"allowed_ip_list": {
													Type:     schema.TypeList,
													Computed: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"port_range": {
													Type:     schema.TypeString,
													Computed: true,
												},
											},
										},
									},
									"internet_list": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"allowed_ip_list": {
													Type:     schema.TypeList,
													Computed: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"port_range": {
													Type:     schema.TypeString,
													Computed: true,
												},
											},
										},
									},
								},
							},
						},
						"domain_endpoint": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ssl_domain_endpoint": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"sasl_domain_endpoint": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"tags": tagsSchema(),
					},
				},
			},
		},
	}
}

func dataSourceAliCloudAlikafkaInstancesRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	kafkaService, err := NewKafkaService(client)
	if err != nil {
		return WrapError(err)
	}

	idsMap := make(map[string]string)
	if v, ok := d.GetOk("ids"); ok {
		for _, vv := range v.([]interface{}) {
			if vv == nil {
				continue
			}
			idsMap[vv.(string)] = vv.(string)
		}
	}
	var nameRegex *regexp.Regexp
	if v, ok := d.GetOk("name_regex"); ok {
		nameRegex = regexp.MustCompile(v.(string))
	}

	objects, err := kafkaService.ListAlikafkaInstances(client.RegionId)
	if err != nil {
		return WrapErrorf(err, DataDefaultErrorMsg, "alicloud_alikafka_instances", "ListAlikafkaInstances", AlibabaCloudSdkGoERROR)
	}

	ids := make([]string, 0)
	names := make([]interface{}, 0)

	s := make([]map[string]interface{}, 0)
	for _, object := range objects {
		if object == nil {
			continue
		}

		if nameRegex != nil && !nameRegex.MatchString(tea.StringValue(object.Name)) {
			continue
		}

		if len(idsMap) > 0 {
			if _, ok := idsMap[object.InstanceId]; !ok {
				continue
			}
		}

		paidType := PostPaid
		if object.PaidType != nil && *object.PaidType == kafka.KafkaPaidTypePrePay {
			paidType = PrePaid
		}

		diskType := 0
		if object.DiskType != nil {
			diskType = int(*object.DiskType)
		}

		diskSize := 0
		if object.DiskSize != nil {
			diskSize = *object.DiskSize
		}

		deployType := 0
		if object.DeployType != nil {
			deployType = int(*object.DeployType)
		}

		ioMax := 0
		if object.IoMax != nil {
			ioMax = *object.IoMax
		}

		eipMax := 0
		if object.EipMax != nil {
			eipMax = *object.EipMax
		}

		expiredTime := int(object.ExpireTime)

		partitionNum := 0
		if object.PartitionNum != nil {
			partitionNum = *object.PartitionNum
		}

		mapping := map[string]interface{}{
			"id":                   object.InstanceId,
			"name":                 tea.StringValue(object.Name),
			"create_time":          fmt.Sprint(object.CreateTime),
			"service_status":       int(object.ServiceStatus),
			"deploy_type":          deployType,
			"vpc_id":               object.VpcId,
			"vswitch_id":           object.VSwitchId,
			"io_max":               ioMax,
			"eip_max":              eipMax,
			"disk_type":            diskType,
			"disk_size":            diskSize,
			"partition_num":        partitionNum,
			"paid_type":            paidType,
			"service_version":      object.Version,
			"spec_type":            tea.StringValue(object.SpecType),
			"zone_id":              object.ZoneId,
			"end_point":            object.EndPoint,
			"security_group":       object.SecurityGroup,
			"config":               object.Config,
			"expired_time":         expiredTime,
			"ssl_end_point":        object.SslEndPoint,
			"domain_endpoint":      object.DomainEndpoint,
			"ssl_domain_endpoint":  object.SslDomainEndpoint,
			"sasl_domain_endpoint": object.SaslDomainEndpoint,
		}

		tags := make(map[string]interface{})
		tagResp, err := kafkaService.DescribeTags(object.InstanceId, nil, TagResourceInstance)
		if err != nil {
			return WrapError(err)
		}
		for k, v := range kafkaService.tagsToMap(tagResp) {
			tags[k] = v
		}
		mapping["tags"] = tags

		DetailInfoMaps := make([]map[string]interface{}, 0)
		if object.Version != "" {
			UpgradeServiceDetailInfoMap := map[string]interface{}{}
			UpgradeServiceDetailInfoMap["current2_open_source_version"] = object.Version
			DetailInfoMaps = append(DetailInfoMaps, UpgradeServiceDetailInfoMap)
		}
		mapping["upgrade_service_detail_info"] = DetailInfoMaps

		ids = append(ids, fmt.Sprint(mapping["id"]))
		names = append(names, mapping["name"])
		id := object.InstanceId

		if d.Get("enable_details").(bool) {
			// quota, err := AlikaService.GetQuotaTip(id)
			// if err != nil {
			// 	return WrapError(err)
			// }
			// mapping["topic_quota"] = quota["TopicQuota"]
			// mapping["partition_num"] = quota["PartitionNumOfBuy"]

			allowedListResp, err := kafkaService.ListAlikafkaAllowedIps(id)
			if err != nil {
				return WrapError(err)
			}

			allowedListMaps := make([]map[string]interface{}, 0)
			if allowedListResp != nil {
				defaultActionsMap := map[string]interface{}{}
				defaultActionsMap["deploy_type"] = fmt.Sprint(allowedListResp.DeployType)

				vpcList := make([]map[string]interface{}, 0)
				for _, vpc := range allowedListResp.VpcList {
					vpcList = append(vpcList, map[string]interface{}{
						"port_range":      vpc.PortRange,
						"allowed_ip_list": vpc.AllowedIpList,
					})
				}
				defaultActionsMap["vpc_list"] = vpcList

				internetList := make([]map[string]interface{}, 0)
				for _, internet := range allowedListResp.InternetList {
					internetList = append(internetList, map[string]interface{}{
						"port_range":      internet.PortRange,
						"allowed_ip_list": internet.AllowedIpList,
					})
				}
				defaultActionsMap["internet_list"] = internetList

				allowedListMaps = append(allowedListMaps, defaultActionsMap)
			}
			mapping["allowed_list"] = allowedListMaps
		}

		s = append(s, mapping)
	}

	d.SetId(dataResourceIdHash(ids))
	if err := d.Set("ids", ids); err != nil {
		return WrapError(err)
	}
	if err := d.Set("names", names); err != nil {
		return WrapError(err)
	}
	if err := d.Set("instances", s); err != nil {
		return WrapError(err)
	}
	if output, ok := d.GetOk("output_file"); ok && output.(string) != "" {
		writeToFile(output.(string), s)
	}

	return nil
}
