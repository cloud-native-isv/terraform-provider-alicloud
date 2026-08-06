package alicloud

import (
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	cmsapi "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/cms"
	commonapi "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/common"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// resourceAliCloudCmsPrometheus manages a CMS 2.0 Prometheus instance
// through the cws-lib-go cms API layer (full CRUD). Views, virtual
// instances and the account-level user setting are exposed through the
// service layer but are not separate resources in this minimal subset.
func resourceAliCloudCmsPrometheus() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudCmsPrometheusCreate,
		Read:   resourceAliCloudCmsPrometheusRead,
		Update: resourceAliCloudCmsPrometheusUpdate,
		Delete: resourceAliCloudCmsPrometheusDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"prometheus_instance_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the Prometheus instance.",
			},
			"payment_type": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The payment type of the Prometheus instance.",
			},
			"workspace": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: "The CMS workspace the Prometheus instance belongs to.",
			},
			"region_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The region in which the Prometheus instance lives.",
			},
			"resource_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The resource type of the Prometheus instance.",
			},
			"instance_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The instance type of the Prometheus instance.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the Prometheus instance.",
			},
			"version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Prometheus version of the instance.",
			},
			"product": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The product the Prometheus instance serves.",
			},
			"access_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The access type of the Prometheus instance.",
			},
			"create_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The creation time of the Prometheus instance.",
			},
			"user_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The owning account of the Prometheus instance.",
			},
		},
	}
}

func resourceAliCloudCmsPrometheusCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewCmsService(client)
	if err != nil {
		return WrapError(err)
	}

	instance := &cmsapi.CmsPrometheusInstance{
		PrometheusInstanceName: d.Get("prometheus_instance_name").(string),
		PaymentType:            d.Get("payment_type").(string),
		Workspace:              d.Get("workspace").(string),
	}
	result, err := service.CreateCmsPrometheusInstance(instance)
	if err != nil {
		return WrapError(err)
	}

	id := result.PrometheusInstanceId
	if id == "" {
		id = result.InstanceId
	}
	if id == "" {
		d.SetId(result.PrometheusInstanceName)
	} else {
		d.SetId(id)
	}
	return resourceAliCloudCmsPrometheusRead(d, meta)
}

func resourceAliCloudCmsPrometheusRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewCmsService(client)
	if err != nil {
		return WrapError(err)
	}

	instance, err := service.GetCmsPrometheusInstance(d.Id())
	if err != nil {
		if commonapi.IsNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	if instance.PrometheusInstanceName != "" {
		d.Set("prometheus_instance_name", instance.PrometheusInstanceName)
	}
	if instance.PaymentType != "" {
		d.Set("payment_type", instance.PaymentType)
	}
	if instance.Workspace != "" {
		d.Set("workspace", instance.Workspace)
	}
	d.Set("region_id", instance.RegionId)
	d.Set("resource_type", instance.ResourceType)
	d.Set("instance_type", instance.InstanceType)
	d.Set("status", instance.Status)
	d.Set("version", instance.Version)
	d.Set("product", instance.Product)
	d.Set("access_type", instance.AccessType)
	d.Set("create_time", instance.CreateTime)
	d.Set("user_id", instance.UserId)
	return nil
}

func resourceAliCloudCmsPrometheusUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewCmsService(client)
	if err != nil {
		return WrapError(err)
	}

	if d.HasChange("prometheus_instance_name") {
		instance := &cmsapi.CmsPrometheusInstance{
			PrometheusInstanceName: d.Get("prometheus_instance_name").(string),
		}
		if _, err := service.UpdateCmsPrometheusInstance(d.Id(), instance); err != nil {
			return WrapError(err)
		}
	}
	return resourceAliCloudCmsPrometheusRead(d, meta)
}

func resourceAliCloudCmsPrometheusDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewCmsService(client)
	if err != nil {
		return WrapError(err)
	}

	if err := service.DeleteCmsPrometheusInstance(d.Id()); err != nil {
		return WrapError(err)
	}
	return nil
}
