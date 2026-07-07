package alicloud

import (
	"fmt"
	"log"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/adbpg"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAliCloudAdbpgVectorCollection() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudAdbpgVectorCollectionCreate,
		Read:   resourceAliCloudAdbpgVectorCollectionRead,
		Delete: resourceAliCloudAdbpgVectorCollectionDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"db_instance_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"collection": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"namespace": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "public",
			},
			"manager_account": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"manager_account_password": {
				Type:      schema.TypeString,
				Optional:  true,
				ForceNew:  true,
				Sensitive: true,
			},
			"dimension": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"metrics": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"parser": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"full_text_retrieval_fields": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"metadata": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAliCloudAdbpgVectorCollectionCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	adbpgService, err := NewAdbpgService(client)
	if err != nil {
		return WrapError(err)
	}

	instanceId := d.Get("db_instance_id").(string)
	collection := d.Get("collection").(string)
	namespace := d.Get("namespace").(string)

	input := &adbpg.AdbpgCollectionCreate{
		Namespace: namespace,
	}
	if v, ok := d.GetOk("manager_account"); ok {
		input.ManagerAccount = v.(string)
	}
	if v, ok := d.GetOk("manager_account_password"); ok {
		input.ManagerAccountPassword = v.(string)
	}
	if v, ok := d.GetOk("dimension"); ok {
		input.Dimension = int64(v.(int))
	}
	if v, ok := d.GetOk("metrics"); ok {
		input.Metrics = v.(string)
	}
	if v, ok := d.GetOk("parser"); ok {
		input.Parser = v.(string)
	}
	if v, ok := d.GetOk("full_text_retrieval_fields"); ok {
		input.FullTextRetrievalFields = v.(string)
	}
	if v, ok := d.GetOk("metadata"); ok {
		input.Metadata = v.(string)
	}

	if err := adbpgService.CreateAdbpgVectorCollection(instanceId, collection, input); err != nil {
		return WrapError(err)
	}

	d.SetId(fmt.Sprintf("%s:%s:%s", instanceId, namespace, collection))

	return resourceAliCloudAdbpgVectorCollectionRead(d, meta)
}

func resourceAliCloudAdbpgVectorCollectionRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	adbpgService, err := NewAdbpgService(client)
	if err != nil {
		return WrapError(err)
	}

	collection, err := adbpgService.DescribeAdbpgVectorCollection(d.Id())
	if err != nil {
		if NotFoundError(err) {
			log.Printf("[WARN] ADBPG Vector Collection %s not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	parts, err := ParseResourceId(d.Id(), 3)
	if err != nil {
		return WrapError(err)
	}
	instanceId := parts[0]
	namespace := parts[1]

	d.Set("db_instance_id", instanceId)
	d.Set("namespace", namespace)
	d.Set("collection", collection.CollectionName)
	d.Set("status", collection.Status)

	// Best-effort enrichment: DescribeCollection requires the namespace password.
	if pwd, ok := d.GetOk("manager_account_password"); ok {
		detail, derr := adbpgService.DescribeAdbpgVectorCollectionDetail(instanceId, namespace, pwd.(string), collection.CollectionName)
		if derr == nil && detail != nil {
			if detail.Dimension > 0 {
				d.Set("dimension", int(detail.Dimension))
			}
			if detail.Metrics != "" {
				d.Set("metrics", detail.Metrics)
			}
			if detail.Parser != "" {
				d.Set("parser", detail.Parser)
			}
			if detail.FullTextRetrievalFields != "" {
				d.Set("full_text_retrieval_fields", detail.FullTextRetrievalFields)
			}
			if detail.Status != "" {
				d.Set("status", detail.Status)
			}
		}
	}

	return nil
}

func resourceAliCloudAdbpgVectorCollectionDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	adbpgService, err := NewAdbpgService(client)
	if err != nil {
		return WrapError(err)
	}

	parts, err := ParseResourceId(d.Id(), 3)
	if err != nil {
		return WrapError(err)
	}
	instanceId := parts[0]
	namespace := parts[1]
	collection := parts[2]

	namespacePassword := ""
	if v, ok := d.GetOk("manager_account_password"); ok {
		namespacePassword = v.(string)
	}

	return WrapError(adbpgService.DeleteAdbpgVectorCollection(instanceId, namespace, namespacePassword, collection))
}
