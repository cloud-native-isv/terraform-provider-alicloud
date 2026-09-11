package alicloud

import (
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunAckAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/ack"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

const resourceAliCloudCSKubernetesAddonName = "alicloud_cs_kubernetes_addon"

// ackAddonCleanupCloudResources releases the cloud resources created by the
// addon when it is uninstalled.
const ackAddonCleanupCloudResources = true

func resourceAliCloudCSKubernetesAddon() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudCSKubernetesAddonCreate,
		Read:   resourceAliCloudCSKubernetesAddonRead,
		Update: resourceAliCloudCSKubernetesAddonUpdate,
		Delete: resourceAliCloudCSKubernetesAddonDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the cluster that the addon is installed in.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the addon, e.g. `terway-eniip`.",
			},
			"version": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The version of the addon. Computed when omitted (service default).",
			},
			"config": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The addon configuration as a JSON string. Computed when omitted.",
			},
			"state": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The state of the installed addon.",
			},
		},
	}
}

// buildAckAddonInstallOptions maps the resource schema onto the addon install
// or modify input. Exposed for offline table-driven tests.
func buildAckAddonInstallOptions(d *schema.ResourceData) *aliyunAckAPI.AckAddonInstallOptions {
	return &aliyunAckAPI.AckAddonInstallOptions{
		Name:    d.Get("name").(string),
		Version: d.Get("version").(string),
		Config:  d.Get("config").(string),
	}
}

func resourceAliCloudCSKubernetesAddonCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	ackService, err := NewAckService(client)
	if err != nil {
		return WrapError(err)
	}

	clusterId := d.Get("cluster_id").(string)
	name := d.Get("name").(string)
	d.SetId(clusterId + ":" + name)

	if _, err := ackService.GetAPI().InstallClusterAddons(clusterId, []aliyunAckAPI.AckAddonInstallOptions{*buildAckAddonInstallOptions(d)}); err != nil {
		return WrapErrorf(err, DefaultErrorMsg, resourceAliCloudCSKubernetesAddonName, "InstallClusterAddons", err)
	}

	// Wait until the installed addon instance is readable.
	stateConf := buildAckStateConf(
		nil, []string{ackAddonStateInstalled},
		d.Timeout(schema.TimeoutCreate), 5*time.Second, 5*time.Second,
		ackService.AckAddonInstanceStateRefreshFunc(d.Id()))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return resourceAliCloudCSKubernetesAddonRead(d, meta)
}

func resourceAliCloudCSKubernetesAddonRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	ackService, err := NewAckService(client)
	if err != nil {
		return WrapError(err)
	}

	clusterId, name, err := ackParseTwoPartId(d.Id())
	if err != nil {
		return WrapError(err)
	}

	instance, err := ackService.GetAPI().DescribeClusterAddonInstance(clusterId, name)
	if err != nil {
		if NotFoundError(err) {
			d.SetId("")
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, resourceAliCloudCSKubernetesAddonName, "DescribeClusterAddonInstance", err)
	}

	d.Set("cluster_id", clusterId)
	d.Set("name", name)
	d.Set("version", instance.Version)
	d.Set("config", instance.Config)
	d.Set("state", instance.State)
	return nil
}

func resourceAliCloudCSKubernetesAddonUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	ackService, err := NewAckService(client)
	if err != nil {
		return WrapError(err)
	}

	clusterId, name, err := ackParseTwoPartId(d.Id())
	if err != nil {
		return WrapError(err)
	}

	if _, err := ackService.GetAPI().ModifyClusterAddon(clusterId, name, buildAckAddonInstallOptions(d)); err != nil {
		return WrapErrorf(err, DefaultErrorMsg, resourceAliCloudCSKubernetesAddonName, "ModifyClusterAddon", err)
	}

	stateConf := buildAckStateConf(
		nil, []string{ackAddonStateInstalled},
		d.Timeout(schema.TimeoutUpdate), 5*time.Second, 5*time.Second,
		ackService.AckAddonInstanceStateRefreshFunc(d.Id()))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return resourceAliCloudCSKubernetesAddonRead(d, meta)
}

func resourceAliCloudCSKubernetesAddonDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	ackService, err := NewAckService(client)
	if err != nil {
		return WrapError(err)
	}

	clusterId, name, err := ackParseTwoPartId(d.Id())
	if err != nil {
		return WrapError(err)
	}

	if _, err := ackService.GetAPI().UnInstallClusterAddons(clusterId, []string{name}, ackAddonCleanupCloudResources); err != nil {
		if NotFoundError(err) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, resourceAliCloudCSKubernetesAddonName, "UnInstallClusterAddons", err)
	}

	// Empty target is the fork's absence-wait idiom (see
	// resourceAliCloudFlinkNamespaceDelete): once the addon is gone the
	// refresh returns nil and WaitForState succeeds, instead of counting
	// NotFoundChecks retries and failing the destroy (B1).
	stateConf := buildAckStateConf(
		nil, []string{},
		d.Timeout(schema.TimeoutDelete), 5*time.Second, 5*time.Second,
		ackService.AckAddonInstanceDeleteStateRefreshFunc(d.Id()))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}
	return nil
}
