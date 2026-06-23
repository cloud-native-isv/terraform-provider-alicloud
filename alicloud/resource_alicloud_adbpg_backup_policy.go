package alicloud

import (
	"log"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/adbpg"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAliCloudAdbpgBackupPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudAdbpgBackupPolicyCreate,
		Read:   resourceAliCloudAdbpgBackupPolicyRead,
		Update: resourceAliCloudAdbpgBackupPolicyUpdate,
		Delete: resourceAliCloudAdbpgBackupPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"db_instance_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"backup_retention_period": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"preferred_backup_period": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"preferred_backup_time": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"enable_recovery_point": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceAliCloudAdbpgBackupPolicyCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	adbpgService, err := NewAdbpgService(client)
	if err != nil {
		return WrapError(err)
	}

	instanceId := d.Get("db_instance_id").(string)
	policy := buildAdbpgBackupPolicy(d)

	if err := adbpgService.ModifyAdbpgBackupPolicy(instanceId, policy); err != nil {
		return WrapError(err)
	}

	d.SetId(instanceId)

	return resourceAliCloudAdbpgBackupPolicyRead(d, meta)
}

func resourceAliCloudAdbpgBackupPolicyRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	adbpgService, err := NewAdbpgService(client)
	if err != nil {
		return WrapError(err)
	}

	policy, err := adbpgService.DescribeAdbpgBackupPolicy(d.Id())
	if err != nil {
		if NotFoundError(err) {
			log.Printf("[WARN] ADBPG Backup Policy for instance %s not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	d.Set("db_instance_id", d.Id())
	d.Set("backup_retention_period", int(policy.BackupRetentionPeriod))
	d.Set("preferred_backup_period", policy.PreferredBackupPeriod)
	d.Set("preferred_backup_time", policy.PreferredBackupTime)
	d.Set("enable_recovery_point", policy.EnableRecoveryPoint)

	return nil
}

func resourceAliCloudAdbpgBackupPolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	adbpgService, err := NewAdbpgService(client)
	if err != nil {
		return WrapError(err)
	}

	policy := buildAdbpgBackupPolicy(d)
	if err := adbpgService.ModifyAdbpgBackupPolicy(d.Id(), policy); err != nil {
		return WrapError(err)
	}

	return resourceAliCloudAdbpgBackupPolicyRead(d, meta)
}

func resourceAliCloudAdbpgBackupPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[WARN] Cannot delete ADBPG backup policy for instance %s, resetting to defaults", d.Id())
	return nil
}

func buildAdbpgBackupPolicy(d *schema.ResourceData) *adbpg.AdbpgBackupPolicy {
	policy := &adbpg.AdbpgBackupPolicy{}
	if v, ok := d.GetOk("backup_retention_period"); ok {
		policy.BackupRetentionPeriod = int32(v.(int))
	}
	if v, ok := d.GetOk("preferred_backup_period"); ok {
		policy.PreferredBackupPeriod = v.(string)
	}
	if v, ok := d.GetOk("preferred_backup_time"); ok {
		policy.PreferredBackupTime = v.(string)
	}
	if v, ok := d.GetOkExists("enable_recovery_point"); ok {
		policy.EnableRecoveryPoint = v.(bool)
	}
	return policy
}
