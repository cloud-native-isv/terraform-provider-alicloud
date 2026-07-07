package alicloud

import (
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceAliCloudAdbpgBackups() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAliCloudAdbpgBackupsRead,
		Schema: map[string]*schema.Schema{
			"db_instance_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"output_file": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"backups": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"backup_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"backup_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"backup_status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"backup_size": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"create_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceAliCloudAdbpgBackupsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	adbpgService, err := NewAdbpgService(client)
	if err != nil {
		return WrapError(err)
	}

	instanceId := d.Get("db_instance_id").(string)
	allBackups, err := adbpgService.ListAdbpgBackups(instanceId)
	if err != nil {
		return WrapError(err)
	}

	var ids []string
	s := make([]map[string]interface{}, 0)

	for _, backup := range allBackups {
		mapping := map[string]interface{}{
			"backup_id":     backup.BackupId,
			"backup_type":   backup.BackupType,
			"backup_status": backup.BackupStatus,
			"backup_size":   int(backup.BackupSize),
			"create_time":   backup.CreateTime,
		}
		ids = append(ids, backup.BackupId)
		s = append(s, mapping)
	}

	d.SetId(dataResourceIdHash(ids))
	d.Set("backups", s)

	if output, ok := d.GetOk("output_file"); ok && output.(string) != "" {
		writeToFile(output.(string), s)
	}

	return nil
}
