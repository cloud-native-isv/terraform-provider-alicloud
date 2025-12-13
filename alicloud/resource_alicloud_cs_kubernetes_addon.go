package alicloud

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

const ResourceAliCloudCSKubernetesAddon = "resourceAliCloudCSKubernetesAddon"

func resourceAliCloudCSKubernetesAddon() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudCSKubernetesAddonCreate,
		Read:   resourceAliCloudCSKubernetesAddonRead,
		Update: resourceAliCloudCSKubernetesAddonUpdate,
		Delete: resourceAliCloudCSKubernetesAddonDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"next_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"can_upgrade": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"required": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"config": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"cleanup_cloud_resources": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func resourceAliCloudCSKubernetesAddonRead(d *schema.ResourceData, meta interface{}) error {
	client, err := meta.(*connectivity.AliyunClient).NewRoaCsClient()
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, ResourceAliCloudCSKubernetesAddon, "InitializeClient", err)
	}
	csClient := CsClient{client}

	parts, err := ParseResourceId(d.Id(), 2)
	if err != nil {
		return WrapError(err)
	}

	addon, err := csClient.DescribeCsKubernetesAddon(d.Id())
	if err != nil {
		if !d.IsNewResource() && NotFoundError(err) {
			d.SetId("")
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, ResourceAliCloudCSKubernetesAddon, "DescribeCsKubernetesAddon", err)
	}

	d.Set("cluster_id", parts[0])
	d.Set("name", addon.ComponentName)
	d.Set("version", addon.Version)
	d.Set("next_version", addon.NextVersion)
	d.Set("can_upgrade", addon.CanUpgrade)
	d.Set("required", addon.Required)
	if addon.Config != "" {
		config := d.Get("config").(string)
		if config == "" {
			config = "{}"
		}
		localConfig := map[string]interface{}{}
		err := json.Unmarshal([]byte(config), &localConfig)
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, ResourceAliCloudCSKubernetesAddon, "json config parse error", err)
		}
		remoteConfig, err := convertJsonStringToMap(addon.Config)
		if err != nil {
			return WrapError(err)
		}
		resultMap := mergeConfigValues(localConfig, remoteConfig)

		result, err := json.Marshal(resultMap)
		if err != nil {
			return WrapError(err)
		}
		if string(result) != "{}" {
			d.Set("config", string(result))
		}
	}

	return nil
}

func resourceAliCloudCSKubernetesAddonCreate(d *schema.ResourceData, meta interface{}) error {
	clusterId := d.Get("cluster_id").(string)
	name := d.Get("name").(string)

	client, err := meta.(*connectivity.AliyunClient).NewRoaCsClient()
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, ResourceAliCloudCSKubernetesAddon, "InitializeClient", err)
	}
	csClient := CsClient{client}

	d.SetId(fmt.Sprintf("%s%s%s", clusterId, COLON_SEPARATED, name))
	addon, err := csClient.DescribeCsKubernetesAddon(d.Id())
	if err != nil {
		if !NotFoundError(err) {
			return WrapErrorf(err, DefaultErrorMsg, ResourceAliCloudCSKubernetesAddon, "DescribeCsKubernetesAddonStatus", err)
		} else {
			// Addon not installed
			err := csClient.installAddon(d)
			if err != nil {
				return WrapErrorf(err, DefaultErrorMsg, ResourceAliCloudCSKubernetesAddon, "installAddon", err)
			}

			stateConf := BuildStateConf([]string{}, []string{"active", "Success", "NoUpgrade"}, d.Timeout(schema.TimeoutCreate), 10*time.Second, csClient.CsKubernetesAddonStateRefreshFunc(clusterId, name, []string{"Failed", "Canceled", "unhealthy"}))
			if _, err = stateConf.WaitForState(); err != nil {
				status, _ := csClient.DescribeCsKubernetesAddonStatus(clusterId, name)
				if status != nil && status.Status != "Success" && status.Error != nil {
					return WrapErrorf(err, DefaultErrorMsg, ResourceAliCloudCSKubernetesAddon, "WaitForSuccessAfterCreate", status.Error)
				}
			}
			// double check addon installed
			_, err = csClient.DescribeCsKubernetesAddon(d.Id())
			if NotFoundError(err) {
				return WrapErrorf(err, DefaultErrorMsg, ResourceAliCloudCSKubernetesAddon, "installAddon", "Unknown Reason. Please check addon config value type first")
			}
		}
	} else {
		// Addon has been installed
		err = updateAddon(csClient, d, addon)
		if err != nil {
			return WrapError(err)
		}
	}

	return resourceAliCloudCSKubernetesAddonRead(d, meta)
}

func resourceAliCloudCSKubernetesAddonUpdate(d *schema.ResourceData, meta interface{}) error {
	client, err := meta.(*connectivity.AliyunClient).NewRoaCsClient()
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, ResourceAliCloudCSKubernetesAddon, "InitializeClient", err)
	}
	csClient := CsClient{client}

	addon, err := csClient.DescribeCsKubernetesAddon(d.Id())
	if err != nil && !NotFoundError(err) {
		return WrapErrorf(err, DefaultErrorMsg, ResourceAliCloudCSKubernetesAddon, "DescribeCsKubernetesAddonStatus", err)
	}

	if d.HasChange("version") || d.HasChange("config") {
		err = updateAddon(csClient, d, addon)
		if err != nil {
			return WrapError(err)
		}
	}

	return resourceAliCloudCSKubernetesAddonRead(d, meta)
}

func resourceAliCloudCSKubernetesAddonDelete(d *schema.ResourceData, meta interface{}) error {
	clusterId := d.Get("cluster_id").(string)
	name := d.Get("name").(string)

	client, err := meta.(*connectivity.AliyunClient).NewRoaCsClient()
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, ResourceAliCloudCSKubernetesAddon, "InitializeClient", err)
	}
	csClient := CsClient{client}

	addonsMetadata, err := csClient.DescribeClusterAddonsMetadata(clusterId)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, ResourceAliCloudCSKubernetesAddon, "DescribeCsKubernetesAddonsMetadata", err)
	}
	addon, ok := addonsMetadata[name]
	if !ok {
		return WrapErrorf(err, DefaultErrorMsg, ResourceAliCloudCSKubernetesAddon, "DescribeCsKubernetesAddonMetadata", err)
	}
	if addon.Required || !IsContain(addon.SupportedActions, "Uninstall") {
		log.Printf("[DEBUG] Skip delete system addon %s\n", name)
		return nil
	}

	err = csClient.uninstallAddon(d)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, ResourceAliCloudCSKubernetesAddon, "uninstallAddon", err)
	}

	stateConf := BuildStateConf([]string{"deleting"}, []string{"deleted"}, d.Timeout(schema.TimeoutDelete), 10*time.Second, csClient.CsKubernetesAddonExistRefreshFunc(clusterId, name))
	if _, err := stateConf.WaitForState(); err != nil {
		status, _ := csClient.DescribeCsKubernetesAddonStatus(clusterId, name)
		if status != nil && status.Error != nil {
			return WrapErrorf(err, DefaultErrorMsg, ResourceAliCloudCSKubernetesAddon, "WaitForSuccessAfterDelete", status.Error)
		}
		return WrapErrorf(err, DefaultErrorMsg, ResourceAliCloudCSKubernetesAddon, "WaitForSuccessAfterDelete", d.Id())
	}

	return nil
}

func updateAddon(csClient CsClient, d *schema.ResourceData, addon *Component) error {
	clusterId := d.Get("cluster_id").(string)
	name := d.Get("name").(string)

	updateVersion, updateConfig, err := needUpgrade(d, addon)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, ResourceAliCloudCSKubernetesAddon, "upgradeAddon", err)
	}
	if updateVersion {
		err := csClient.upgradeAddon(d, updateVersion, updateConfig)
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, ResourceAliCloudCSKubernetesAddon, "upgradeAddon", err)
		}
	} else if updateConfig {
		err := csClient.updateAddonConfig(d)
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, ResourceAliCloudCSKubernetesAddon, "updateAddonConfig", err)
		}
	}
	if !updateVersion && !updateConfig {
		return nil
	}

	stateConf := BuildStateConf([]string{}, []string{"Success"}, d.Timeout(schema.TimeoutUpdate), 10*time.Second, csClient.CsKubernetesAddonTaskRefreshFunc(clusterId, name, []string{"Failed"}))
	if _, err := stateConf.WaitForState(); err != nil {
		status, _ := csClient.DescribeCsKubernetesAddonStatus(clusterId, name)
		if status != nil && status.Error != nil {
			return WrapErrorf(err, DefaultErrorMsg, ResourceAliCloudCSKubernetesAddon, "WaitForSuccessAfterUpdate", status.Error)
		}
		return WrapErrorf(err, DefaultErrorMsg, ResourceAliCloudCSKubernetesAddon, "WaitForSuccessAfterUpdate", d.Id())
	}

	return nil
}

func needUpgrade(d *schema.ResourceData, c *Component) (updateVersion bool, updateConfig bool, err error) {
	// Is version changed
	version := d.Get("version").(string)
	if version != "" && c.Version != "" && c.Version != version {
		updateVersion = true
	}

	// Is config changed
	if _, ok := d.GetOk("config"); ok {
		localConfig := map[string]interface{}{}
		err := json.Unmarshal([]byte(d.Get("config").(string)), &localConfig)
		if err != nil {
			return false, false, WrapErrorf(err, DefaultErrorMsg, ResourceAliCloudCSKubernetesAddon, "json config parse error", err)
		}

		if c.Config == "" {
			c.Config = "{}"
		}
		remoteConfig := map[string]interface{}{}
		err = json.Unmarshal([]byte(c.Config), &remoteConfig)
		if err != nil {
			return false, false, WrapErrorf(err, DefaultErrorMsg, ResourceAliCloudCSKubernetesAddon, "json config parse error", err)
		}
		// if localConfig is subset of remoteConfig, do not call update to do action
		updateConfig = !isSubset(localConfig, remoteConfig)
	}
	return
}

// return true if oldMap is subset of newMap
func isSubset(oldMap, newMap map[string]interface{}) bool {
	for key, oldVal := range oldMap {
		newVal, exists := newMap[key]
		if !exists {
			return false
		}
		oldValMap, oldValIsMap := oldVal.(map[string]interface{})
		newValMap, newValIsMap := newVal.(map[string]interface{})
		if oldValIsMap && newValIsMap {
			if !isSubset(oldValMap, newValMap) {
				return false
			}
			continue
		}

		if oldValIsMap != newValIsMap || !reflect.DeepEqual(oldVal, newVal) {
			return false
		}
	}
	return true
}

func mergeConfigValues(oldMap, newMap map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})
	for key, newValue := range newMap {
		if oldValue, exists := oldMap[key]; exists {
			switch oldValueTyped := oldValue.(type) {
			case map[string]interface{}:
				if newMapTyped, ok := newValue.(map[string]interface{}); ok {
					merged[key] = mergeConfigValues(oldValueTyped, newMapTyped)
					continue
				}
			default:
				merged[key] = newValue
			}
		}
	}

	return merged
}
