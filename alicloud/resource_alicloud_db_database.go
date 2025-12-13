package alicloud

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAliCloudDBDatabase() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudDBDatabaseCreate,
		Read:   resourceAliCloudDBDatabaseRead,
		Update: resourceAliCloudDBDatabaseUpdate,
		Delete: resourceAliCloudDBDatabaseDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		CustomizeDiff: func(d *schema.ResourceDiff, v interface{}) error {
			// Plan-time read-only detection for adoption
			instanceIdRaw, nameRaw := d.Get("instance_id"), d.Get("name")
			if instanceIdRaw == nil || nameRaw == nil {
				return nil
			}
			instanceId := fmt.Sprint(instanceIdRaw)
			name := fmt.Sprint(nameRaw)
			if instanceId == "" || name == "" {
				return nil
			}

			client := v.(*connectivity.AliyunClient)
			rdsService, err := NewRdsService(client)
			if err != nil {
				d.SetNew("adoption_notice", "Could not initialize RDS service for adoption detection during plan")
				return nil
			}
			id := EncodeDBId(instanceId, name)
			obj, err := rdsService.DescribeDBDatabase(id)
			if err != nil {
				if NotFoundError(err) {
					d.SetNew("adopt_existing", false)
					d.SetNewComputed("adoption_notice")
					return nil
				}
				// Permission or throttling degradation without failing plan
				if rdsService.IsPermissionDenied(err) {
					d.SetNew("adoption_notice", "Insufficient permission for read-only detection during plan. Require: Describe/List Databases permission on the target RDS instance.")
				} else if IsExpectedErrors(err, []string{"ServiceUnavailable", "ThrottlingException", "InternalError", "Throttling", "SystemBusy"}) || NeedRetry(err) {
					d.SetNew("adoption_notice", "Throttling or temporary error during plan detection; proceeding without adoption confirmation.")
				} else {
					d.SetNew("adoption_notice", "Could not confirm whether to adopt existing database during plan (unknown error)")
				}
				return nil
			}
			d.SetNew("adopt_existing", true)
			// If description differs, inform that adoption will not align it in the same apply round
			if vDesc, ok := d.GetOk("description"); ok {
				desired := fmt.Sprint(vDesc)
				actual := fmt.Sprint(obj["DBDescription"])
				if desired != "" && !strings.EqualFold(desired, actual) {
					d.SetNew("adoption_notice", "Detected existing database and will adopt it on apply; description differs and won't be aligned in this apply.")
				} else {
					d.SetNew("adoption_notice", "Detected existing database and will adopt it on apply")
				}
			} else {
				d.SetNew("adoption_notice", "Detected existing database and will adopt it on apply")
			}
			return nil
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"instance_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},

			"name": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-z][a-z0-9_-]*[a-z0-9]$`), "The name can consist of lowercase letters, numbers, underscores, and middle lines, and must begin with letters and end with letters or numbers"),
			},

			"character_set": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "utf8mb4",
				ForceNew: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if strings.ToLower(old) == strings.ToLower(new) {
						return true
					}
					newArray := strings.Split(new, ",")
					oldArray := strings.Split(old, ",")
					if d.Id() != "" && len(oldArray) > 1 && len(newArray) == 1 && strings.ToLower(newArray[0]) == strings.ToLower(oldArray[0]) {
						return true
					}
					/*
					  SQLServer creates a database, when a non native engine character set is passed in, the SDK will assign the default character set.
					*/
					if old == "Chinese_PRC_CI_AS" && new == "utf8" {
						return true
					}
					return false
				},
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			// Read-only adoption hints for plan/apply transparency
			"adopt_existing": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Whether the provider will adopt an existing database on apply.",
			},
			"adoption_notice": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Human-readable notice about adoption behavior shown at plan/apply.",
			},
		},
	}
}

func resourceAliCloudDBDatabaseCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	rdsService, err := NewRdsService(client)
	if err != nil {
		return WrapError(err)
	}

	instanceId := d.Get("instance_id").(string)
	name := d.Get("name").(string)
	id := EncodeDBId(instanceId, name)

	// Adoption path: if DB already exists
	if obj, err := rdsService.DescribeDBDatabase(id); err == nil {
		// Immutable field conflict checks per engine rules (US2)
		if v, ok := d.GetOk("character_set"); ok {
			desired := fmt.Sprint(v)
			if desired != "" {
				engine := fmt.Sprint(obj["Engine"])
				if engine == string(PostgreSQL) {
					// For PostgreSQL, only enforce when desired is a triplet "Charset,Collate,Ctype"
					if strings.Contains(desired, ",") {
						actual := fmt.Sprintf("%s,%s,%s", fmt.Sprint(obj["CharacterSetName"]), fmt.Sprint(obj["Collate"]), fmt.Sprint(obj["Ctype"]))
						if !strings.EqualFold(desired, actual) {
							return WrapError(fmt.Errorf("immutable field conflict: character_set for engine %s differs (existing=%s, desired=%s). Adoption aborted; remove character_set or recreate DB with desired settings.", engine, actual, desired))
						}
					}
				} else if engine == string(MySQL) || engine == string(SQLServer) {
					actual := fmt.Sprint(obj["CharacterSetName"])
					if !strings.EqualFold(desired, actual) {
						return WrapError(fmt.Errorf("immutable field conflict: character_set for engine %s differs (existing=%s, desired=%s). Adoption aborted; remove character_set or recreate DB with desired settings.", engine, actual, desired))
					}
				}
			}
		}
		log.Printf("[INFO] Adopting existing RDS database: %s", id)
		d.SetId(id)
		return resourceAliCloudDBDatabaseRead(d, meta)
	} else if !NotFoundError(err) {
		if NeedRetry(err) {
			time.Sleep(5 * time.Second)
		} else {
			log.Printf("[WARN] DescribeDBDatabase during create returned error (ignored): %v", err)
		}
	}

	// Create new database
	characterSet := d.Get("character_set").(string)
	description := ""
	if v, ok := d.GetOk("description"); ok {
		description = v.(string)
	}

	if err := rdsService.CreateDBDatabase(instanceId, name, characterSet, description); err != nil {
		return WrapErrorf(err, DefaultErrorMsg, id, "CreateDBDatabase", AlibabaCloudSdkGoERROR)
	}

	d.SetId(id)
	// Wait until DB exists
	if wErr := rdsService.WaitForDBDatabaseCreating(d.Id(), d.Timeout(schema.TimeoutCreate)); wErr != nil {
		return WrapErrorf(wErr, IdMsg, d.Id())
	}
	return resourceAliCloudDBDatabaseRead(d, meta)
}

func resourceAliCloudDBDatabaseRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	rdsService, err := NewRdsService(client)
	if err != nil {
		return WrapError(err)
	}
	object, err := rdsService.DescribeDBDatabase(d.Id())
	if err != nil {
		if NotFoundError(err) {
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	d.Set("instance_id", object["DBInstanceId"])
	d.Set("name", object["DBName"])
	if string(PostgreSQL) == object["Engine"] {
		// Use safe string conversion to avoid panics when fields are missing or types vary
		var strArray = []string{fmt.Sprint(object["CharacterSetName"]), fmt.Sprint(object["Collate"]), fmt.Sprint(object["Ctype"])}
		postgreSQLCharacterSet := strings.Join(strArray, ",")
		d.Set("character_set", postgreSQLCharacterSet)
	} else {
		d.Set("character_set", fmt.Sprint(object["CharacterSetName"]))
	}
	d.Set("description", object["DBDescription"])

	// Adoption transparency fields (stable values)
	d.Set("adoption_notice", "Database is under Terraform management.")
	d.Set("adopt_existing", true)

	return nil
}

func resourceAliCloudDBDatabaseUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	rdsService, err := NewRdsService(client)
	if err != nil {
		return WrapError(err)
	}
	if d.HasChange("description") && !d.IsNewResource() {
		if err := rdsService.ModifyDBDatabaseDescription(d.Id(), d.Get("description").(string)); err != nil {
			return WrapError(err)
		}
	}
	return resourceAliCloudDBDatabaseRead(d, meta)
}

func resourceAliCloudDBDatabaseDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	rdsService, err := NewRdsService(client)
	if err != nil {
		return WrapError(err)
	}
	instanceId, _, err := DecodeDBId(d.Id())
	if err != nil {
		parts, e2 := ParseResourceId(d.Id(), 2)
		if e2 != nil {
			return WrapError(err)
		}
		instanceId = parts[0]
	}
	// wait instance status is running before deleting database
	if err := rdsService.WaitForDBInstance(instanceId, Running, 1800); err != nil {
		return WrapError(err)
	}
	if err := rdsService.DeleteDBDatabase(d.Id()); err != nil {
		return WrapError(err)
	}
	// wait for deletion finished
	if err := rdsService.WaitForDBDatabaseDeleted(d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return WrapError(err)
	}
	return nil
}
