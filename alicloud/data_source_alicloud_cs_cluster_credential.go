package alicloud

import (
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunAckAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/ack"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

const dataSourceAliCloudCSClusterCredentialName = "alicloud_cs_cluster_credential"

func dataSourceAliCloudCSClusterCredential() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAliCloudCSClusterCredentialRead,

		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the cluster.",
			},
			"temporary_duration_minutes": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The validity period (minutes) of the temporary kubeconfig credential.",
			},
			"private_ip_address": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Whether to use the private API server endpoint in the kubeconfig.",
			},

			// Computed values
			"config": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "The kubeconfig YAML of the cluster. Sensitive: it is a cluster access credential.",
			},
			"expiration": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The expiration time of the kubeconfig credential.",
			},
		},
	}
}

// buildAckKubeConfigOptions maps the data source schema onto the kubeconfig
// retrieval options. Exposed for offline table-driven tests.
func buildAckKubeConfigOptions(d *schema.ResourceData) *aliyunAckAPI.AckKubeConfigOptions {
	return &aliyunAckAPI.AckKubeConfigOptions{
		PrivateIpAddress:         d.Get("private_ip_address").(bool),
		TemporaryDurationMinutes: int64(d.Get("temporary_duration_minutes").(int)),
	}
}

func dataSourceAliCloudCSClusterCredentialRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	ackService, err := NewAckService(client)
	if err != nil {
		return WrapError(err)
	}

	clusterId := d.Get("cluster_id").(string)
	kubeConfig, err := ackService.GetAPI().DescribeClusterUserKubeconfig(clusterId, buildAckKubeConfigOptions(d))
	if err != nil {
		return WrapErrorf(err, DataDefaultErrorMsg, dataSourceAliCloudCSClusterCredentialName, "DescribeClusterUserKubeconfig", err)
	}

	d.SetId(clusterId)
	d.Set("cluster_id", clusterId)
	d.Set("config", kubeConfig.Config)
	d.Set("expiration", kubeConfig.Expiration)
	return nil
}
