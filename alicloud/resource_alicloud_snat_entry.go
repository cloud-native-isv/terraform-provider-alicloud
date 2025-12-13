// Package alicloud. This file is generated automatically. Please do not modify it manually, thank you!
package alicloud

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAliCloudNATGatewaySnatEntry() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudNATGatewaySnatEntryCreate,
		Read:   resourceAliCloudNATGatewaySnatEntryRead,
		Update: resourceAliCloudNATGatewaySnatEntryUpdate,
		Delete: resourceAliCloudNATGatewaySnatEntryDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"eip_affinity": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"snat_entry_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"snat_entry_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"snat_ip": {
				Type:     schema.TypeString,
				Required: true,
			},
			"snat_table_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"source_cidr": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				Computed:      true,
				ConflictsWith: strings.Fields("source_vswitch_id"),
			},
			"source_vswitch_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				Computed:      true,
				ConflictsWith: strings.Fields("source_cidr"),
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAliCloudNATGatewaySnatEntryCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	svc, _ := NewVpcSnatEntryService(client)

	// Build request
	request := make(map[string]interface{})
	request["SnatTableId"] = d.Get("snat_table_id")
	request["SnatIp"] = d.Get("snat_ip")
	if v, ok := d.GetOk("source_vswitch_id"); ok {
		request["SourceVSwitchId"] = v
	}
	if v, ok := d.GetOk("snat_entry_name"); ok {
		request["SnatEntryName"] = v
	}
	if v, ok := d.GetOk("source_cidr"); ok {
		request["SourceCIDR"] = v
	}
	if v, ok := d.GetOkExists("eip_affinity"); ok {
		request["EipAffinity"] = v
	}

	snatEntryId, err := svc.CreateSnatEntryWithTimeout(request, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return WrapError(err)
	}

	d.SetId(fmt.Sprintf("%v:%v", request["SnatTableId"], snatEntryId))

	if err := svc.WaitForSnatEntryCreating(d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return WrapError(err)
	}

	return resourceAliCloudNATGatewaySnatEntryRead(d, meta)
}

func resourceAliCloudNATGatewaySnatEntryRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	svc, _ := NewVpcSnatEntryService(client)

	// compatible with previous id which in under 1.37.0
	if strings.HasPrefix(d.Id(), "snat-") {
		d.SetId(fmt.Sprintf("%s%s%s", d.Get("snat_table_id").(string), COLON_SEPARATED, d.Id()))
	}

	objectRaw, err := svc.DescribeSnatEntry(d.Id())
	if err != nil {
		if !d.IsNewResource() && NotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_snat_entry DescribeNATGatewaySnatEntry Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	if objectRaw["EipAffinity"] != nil {
		d.Set("eip_affinity", formatInt(objectRaw["EipAffinity"]))
	}
	if objectRaw["SnatEntryName"] != nil {
		d.Set("snat_entry_name", objectRaw["SnatEntryName"])
	}
	if objectRaw["SnatIp"] != nil {
		d.Set("snat_ip", objectRaw["SnatIp"])
	}
	if objectRaw["SourceCIDR"] != nil {
		d.Set("source_cidr", objectRaw["SourceCIDR"])
	}
	if objectRaw["SourceVSwitchId"] != nil {
		d.Set("source_vswitch_id", objectRaw["SourceVSwitchId"])
	}
	if objectRaw["Status"] != nil {
		d.Set("status", objectRaw["Status"])
	}
	if objectRaw["SnatEntryId"] != nil {
		d.Set("snat_entry_id", objectRaw["SnatEntryId"])
	}
	if objectRaw["SnatTableId"] != nil {
		d.Set("snat_table_id", objectRaw["SnatTableId"])
	}

	return nil
}

func resourceAliCloudNATGatewaySnatEntryUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	svc, _ := NewVpcSnatEntryService(client)
	update := false

	// compatible with previous id which in under 1.37.0
	if strings.HasPrefix(d.Id(), "snat-") {
		d.SetId(fmt.Sprintf("%s%s%s", d.Get("snat_table_id").(string), COLON_SEPARATED, d.Id()))
	}

	// Build attrs for update
	attrs := map[string]interface{}{}
	if d.HasChange("snat_entry_name") {
		update = true
	}
	if v, ok := d.GetOk("snat_entry_name"); ok {
		attrs["SnatEntryName"] = v
	}

	if d.HasChange("snat_ip") {
		update = true
	}
	attrs["SnatIp"] = d.Get("snat_ip")

	if d.HasChange("eip_affinity") {
		update = true

		if v, ok := d.GetOkExists("eip_affinity"); ok {
			attrs["EipAffinity"] = v
		}
	}

	if update {
		if err := svc.ModifySnatEntryWithTimeout(d.Id(), attrs, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return WrapError(err)
		}
		if err := svc.WaitForSnatEntryCreating(d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return WrapError(err)
		}
	}

	return resourceAliCloudNATGatewaySnatEntryRead(d, meta)
}

func resourceAliCloudNATGatewaySnatEntryDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	svc, _ := NewVpcSnatEntryService(client)

	// compatible with previous id which in under 1.37.0
	if strings.HasPrefix(d.Id(), "snat-") {
		d.SetId(fmt.Sprintf("%s%s%s", d.Get("snat_table_id").(string), COLON_SEPARATED, d.Id()))
	}

	if err := svc.DeleteSnatEntryWithTimeout(d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return WrapError(err)
	}
	if err := svc.WaitForSnatEntryDeleting(d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return WrapError(err)
	}
	return nil
}
