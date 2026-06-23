package alicloud

import (
	"regexp"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceAliCloudAdbpgAccounts() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAliCloudAdbpgAccountsRead,
		Schema: map[string]*schema.Schema{
			"db_instance_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.ValidateRegexp,
			},
			"output_file": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"accounts": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"account_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"account_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"account_status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"account_description": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceAliCloudAdbpgAccountsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	adbpgService, err := NewAdbpgService(client)
	if err != nil {
		return WrapError(err)
	}

	instanceId := d.Get("db_instance_id").(string)
	allAccounts, err := adbpgService.ListAdbpgAccounts(instanceId)
	if err != nil {
		return WrapError(err)
	}

	var nameRegex *regexp.Regexp
	if v, ok := d.GetOk("name_regex"); ok {
		nameRegex = regexp.MustCompile(v.(string))
	}

	var names []string
	s := make([]map[string]interface{}, 0)

	for _, acct := range allAccounts {
		if nameRegex != nil && !nameRegex.MatchString(acct.AccountName) {
			continue
		}
		mapping := map[string]interface{}{
			"account_name":        acct.AccountName,
			"account_type":        acct.AccountType,
			"account_status":      acct.AccountStatus,
			"account_description": acct.AccountDescription,
		}
		names = append(names, acct.AccountName)
		s = append(s, mapping)
	}

	d.SetId(dataResourceIdHash(names))
	d.Set("accounts", s)

	if output, ok := d.GetOk("output_file"); ok && output.(string) != "" {
		writeToFile(output.(string), s)
	}

	return nil
}
