package alicloud

import (
	"fmt"
	"log"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/adbpg"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAliCloudAdbpgDatabase() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudAdbpgDatabaseCreate,
		Read:   resourceAliCloudAdbpgDatabaseRead,
		Delete: resourceAliCloudAdbpgDatabaseDelete,
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
			"db_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"db_description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"character_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAliCloudAdbpgDatabaseCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	adbpgService, err := NewAdbpgService(client)
	if err != nil {
		return WrapError(err)
	}

	instanceId := d.Get("db_instance_id").(string)
	input := &adbpg.AdbpgDatabaseCreate{
		DBName: d.Get("db_name").(string),
	}
	if v, ok := d.GetOk("db_description"); ok {
		input.DBDescription = v.(string)
	}
	if v, ok := d.GetOk("character_name"); ok {
		input.CharacterName = v.(string)
	}

	if err := adbpgService.CreateAdbpgDatabase(instanceId, input); err != nil {
		return WrapError(err)
	}

	d.SetId(fmt.Sprintf("%s:%s", instanceId, input.DBName))

	return resourceAliCloudAdbpgDatabaseRead(d, meta)
}

func resourceAliCloudAdbpgDatabaseRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	adbpgService, err := NewAdbpgService(client)
	if err != nil {
		return WrapError(err)
	}

	database, err := adbpgService.DescribeAdbpgDatabase(d.Id())
	if err != nil {
		if NotFoundError(err) {
			log.Printf("[WARN] ADBPG Database %s not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	parts, err := ParseResourceId(d.Id(), 2)
	if err != nil {
		return WrapError(err)
	}

	d.Set("db_instance_id", parts[0])
	d.Set("db_name", database.DBName)
	d.Set("db_description", database.DBDescription)
	d.Set("character_name", database.CharacterName)

	return nil
}

func resourceAliCloudAdbpgDatabaseDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	adbpgService, err := NewAdbpgService(client)
	if err != nil {
		return WrapError(err)
	}

	parts, err := ParseResourceId(d.Id(), 2)
	if err != nil {
		return WrapError(err)
	}

	return WrapError(adbpgService.DeleteAdbpgDatabase(parts[0], parts[1]))
}
