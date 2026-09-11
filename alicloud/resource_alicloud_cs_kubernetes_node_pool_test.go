package alicloud

import (
	"reflect"
	"testing"
	"time"

	aliyunAckAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/ack"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func TestCsKubernetesNodePoolSchemaContract(t *testing.T) {
	resource := resourceAliCloudAckNodepool()
	if err := resource.InternalValidate(nil, true); err != nil {
		t.Fatalf("InternalValidate() error = %v", err)
	}

	for _, name := range []string{"cluster_id", "name", "instance_types", "vswitch_ids"} {
		field, ok := resource.Schema[name]
		if !ok || !field.Required {
			t.Fatalf("field %q must be Required: %#v", name, field)
		}
	}
	if !resource.Schema["cluster_id"].ForceNew {
		t.Fatalf("cluster_id must be ForceNew")
	}
	instanceTypes := resource.Schema["instance_types"]
	if instanceTypes.Type != schema.TypeList || instanceTypes.MinItems != 1 {
		t.Fatalf("instance_types schema = %#v", instanceTypes)
	}
	chargeType := resource.Schema["instance_charge_type"]
	if !chargeType.Optional || chargeType.Default != "PostPaid" {
		t.Fatalf("instance_charge_type schema = %#v", chargeType)
	}
	if _, errs := chargeType.ValidateFunc("PostPaid", "instance_charge_type"); len(errs) != 0 {
		t.Fatalf("instance_charge_type=PostPaid rejected: %v", errs)
	}
	if _, errs := chargeType.ValidateFunc("Subscription", "instance_charge_type"); len(errs) == 0 {
		t.Fatalf("instance_charge_type=Subscription must be rejected")
	}
	management := resource.Schema["management"]
	if management.Type != schema.TypeList || management.MaxItems != 1 {
		t.Fatalf("management schema = %#v", management)
	}
	for _, name := range []string{"status", "node_count"} {
		field, ok := resource.Schema[name]
		if !ok || !field.Computed {
			t.Fatalf("field %q must be Computed: %#v", name, field)
		}
	}
	if resource.Timeouts == nil || resource.Timeouts.Create == nil || *resource.Timeouts.Create != 20*time.Minute {
		t.Fatalf("create timeout must be 20m: %#v", resource.Timeouts)
	}
}

func TestCsKubernetesNodePoolBuildConfig(t *testing.T) {
	base := func() map[string]interface{} {
		return map[string]interface{}{
			"cluster_id":     "c1a2b3",
			"name":           "worker-pool",
			"instance_types": []interface{}{"ecs.u1-c1m1.large"},
			"vswitch_ids":    []interface{}{"vsw-1", "vsw-2"},
		}
	}
	for _, tc := range []struct {
		name     string
		mutate   func(config map[string]interface{})
		validate func(t *testing.T, config *aliyunAckAPI.AckNodePoolConfig)
	}{
		{
			name:   "base config with defaults",
			mutate: func(map[string]interface{}) {},
			validate: func(t *testing.T, config *aliyunAckAPI.AckNodePoolConfig) {
				if config.NodePoolName != "worker-pool" {
					t.Fatalf("NodePoolName = %q", config.NodePoolName)
				}
				if config.InstanceType != "ecs.u1-c1m1.large" {
					t.Fatalf("InstanceType = %q, want the first instance_types entry", config.InstanceType)
				}
				if !reflect.DeepEqual(config.VSwitchIds, []string{"vsw-1", "vsw-2"}) {
					t.Fatalf("VSwitchIds = %v", config.VSwitchIds)
				}
				if config.InstanceChargeType != "PostPaid" {
					t.Fatalf("InstanceChargeType = %q, want default PostPaid", config.InstanceChargeType)
				}
				if config.SystemDisk != nil {
					t.Fatalf("SystemDisk must be nil when no disk field is set")
				}
				if config.Management != nil {
					t.Fatalf("Management must be nil when not configured")
				}
			},
		},
		{
			name: "scaling and charge fields",
			mutate: func(config map[string]interface{}) {
				config["desired_size"] = 1
				config["min_size"] = 1
				config["max_size"] = 5
				config["auto_scaling"] = true
				config["instance_charge_type"] = "PrePaid"
				config["image_type"] = "AliyunLinux3"
				config["security_group_id"] = "sg-1"
				config["tags"] = map[string]interface{}{"pool": "workers"}
			},
			validate: func(t *testing.T, config *aliyunAckAPI.AckNodePoolConfig) {
				if config.DesiredSize != 1 || config.MinSize != 1 || config.MaxSize != 5 {
					t.Fatalf("sizes = %d/%d/%d, want 1/1/5", config.DesiredSize, config.MinSize, config.MaxSize)
				}
				if !config.AutoScaling {
					t.Fatalf("AutoScaling = false, want true")
				}
				if config.InstanceChargeType != "PrePaid" {
					t.Fatalf("InstanceChargeType = %q", config.InstanceChargeType)
				}
				if config.ImageType != "AliyunLinux3" || config.SecurityGroupId != "sg-1" {
					t.Fatalf("ImageType/SecurityGroupId = %q/%q", config.ImageType, config.SecurityGroupId)
				}
				if !reflect.DeepEqual(config.Tags, []aliyunAckAPI.AckTag{{Key: "pool", Value: "workers"}}) {
					t.Fatalf("Tags = %v", config.Tags)
				}
			},
		},
		{
			name: "system disk is attached only when a disk field is set",
			mutate: func(config map[string]interface{}) {
				config["system_disk_category"] = "cloud_essd"
				config["system_disk_size"] = 40
				config["system_disk_performance_level"] = "PL1"
			},
			validate: func(t *testing.T, config *aliyunAckAPI.AckNodePoolConfig) {
				if config.SystemDisk == nil {
					t.Fatalf("SystemDisk must be set")
				}
				if config.SystemDisk.Category != "cloud_essd" || config.SystemDisk.Size != 40 || config.SystemDisk.PerformanceLevel != "PL1" {
					t.Fatalf("SystemDisk = %+v", config.SystemDisk)
				}
			},
		},
		{
			name: "runtime labels and taints",
			mutate: func(config map[string]interface{}) {
				config["runtime_name"] = "containerd"
				config["runtime_version"] = "1.6.20"
				config["labels"] = []interface{}{
					map[string]interface{}{"key": "workload", "value": "api"},
				}
				config["taints"] = []interface{}{
					map[string]interface{}{"key": "spot", "value": "true", "effect": "NoSchedule"},
				}
			},
			validate: func(t *testing.T, config *aliyunAckAPI.AckNodePoolConfig) {
				if config.Runtime == nil || config.Runtime.Name != "containerd" || config.Runtime.Version != "1.6.20" {
					t.Fatalf("Runtime = %+v", config.Runtime)
				}
				if !reflect.DeepEqual(config.Labels, []aliyunAckAPI.AckNodeLabel{{Key: "workload", Value: "api"}}) {
					t.Fatalf("Labels = %v", config.Labels)
				}
				if !reflect.DeepEqual(config.Taints, []aliyunAckAPI.AckNodeTaint{{Key: "spot", Value: "true", Effect: "NoSchedule"}}) {
					t.Fatalf("Taints = %v", config.Taints)
				}
			},
		},
		{
			name: "management block enables policies",
			mutate: func(config map[string]interface{}) {
				config["management"] = []interface{}{map[string]interface{}{
					"auto_repair":  true,
					"auto_upgrade": true,
					"auto_vulfix":  true,
				}}
			},
			validate: func(t *testing.T, config *aliyunAckAPI.AckNodePoolConfig) {
				if config.Management == nil {
					t.Fatalf("Management must be set")
				}
				if config.Management.AutoRepairPolicy == nil || !config.Management.AutoRepairPolicy.RestartNodes {
					t.Fatalf("AutoRepairPolicy = %+v", config.Management.AutoRepairPolicy)
				}
				if config.Management.AutoUpgradePolicy == nil || !config.Management.AutoUpgradePolicy.KubeletUpgrade {
					t.Fatalf("AutoUpgradePolicy = %+v", config.Management.AutoUpgradePolicy)
				}
				if config.Management.AutoVulFixPolicy == nil || !config.Management.AutoVulFixPolicy.RestartNodes || config.Management.AutoVulFixPolicy.VulLevels != "asap" {
					t.Fatalf("AutoVulFixPolicy = %+v", config.Management.AutoVulFixPolicy)
				}
			},
		},
		{
			name: "management block with all flags false maps to nil",
			mutate: func(config map[string]interface{}) {
				config["management"] = []interface{}{map[string]interface{}{
					"auto_repair":  false,
					"auto_upgrade": false,
					"auto_vulfix":  false,
				}}
			},
			validate: func(t *testing.T, config *aliyunAckAPI.AckNodePoolConfig) {
				if config.Management != nil {
					t.Fatalf("Management = %+v, want nil when no policy is enabled", config.Management)
				}
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			config := base()
			tc.mutate(config)
			d := schema.TestResourceDataRaw(t, resourceAliCloudAckNodepool().Schema, config)
			tc.validate(t, buildAckNodePoolConfig(d))
		})
	}
}

func TestCsKubernetesNodePoolManagementMapping(t *testing.T) {
	for _, tc := range []struct {
		name       string
		management map[string]interface{}
		validate   func(t *testing.T, management *aliyunAckAPI.AckNodePoolManagement)
	}{
		{
			name:       "only auto repair",
			management: map[string]interface{}{"auto_repair": true},
			validate: func(t *testing.T, management *aliyunAckAPI.AckNodePoolManagement) {
				if management == nil || management.AutoRepairPolicy == nil {
					t.Fatalf("management = %+v, want AutoRepairPolicy", management)
				}
				if management.AutoUpgradePolicy != nil || management.AutoVulFixPolicy != nil {
					t.Fatalf("management = %+v, want no other policies", management)
				}
			},
		},
		{
			name:       "only auto upgrade",
			management: map[string]interface{}{"auto_upgrade": true},
			validate: func(t *testing.T, management *aliyunAckAPI.AckNodePoolManagement) {
				if management == nil || management.AutoUpgradePolicy == nil || !management.AutoUpgradePolicy.KubeletUpgrade {
					t.Fatalf("management = %+v, want AutoUpgradePolicy with kubelet upgrade", management)
				}
				if management.AutoUpgradePolicy.RuntimeUpgrade || management.AutoUpgradePolicy.OsUpgrade {
					t.Fatalf("AutoUpgradePolicy = %+v, want only kubelet upgrade from the boolean flag", management.AutoUpgradePolicy)
				}
			},
		},
		{
			name:       "only auto vulfix",
			management: map[string]interface{}{"auto_vulfix": true},
			validate: func(t *testing.T, management *aliyunAckAPI.AckNodePoolManagement) {
				if management == nil || management.AutoVulFixPolicy == nil || management.AutoVulFixPolicy.VulLevels != "asap" {
					t.Fatalf("management = %+v, want AutoVulFixPolicy with VulLevels asap", management)
				}
			},
		},
		{
			name:       "empty management",
			management: map[string]interface{}{},
			validate: func(t *testing.T, management *aliyunAckAPI.AckNodePoolManagement) {
				if management != nil {
					t.Fatalf("management = %+v, want nil", management)
				}
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc.validate(t, buildAckNodePoolManagement(tc.management))
		})
	}
}

func TestCsKubernetesNodePoolSetAttributes(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceAliCloudAckNodepool().Schema, map[string]interface{}{
		"cluster_id":     "c1a2b3",
		"name":           "will-be-overwritten",
		"instance_types": []interface{}{"ecs.old"},
		"vswitch_ids":    []interface{}{"vsw-old"},
	})
	pool := &aliyunAckAPI.AckNodePool{
		ClusterId:          "c1a2b3",
		NodePoolId:         "np-42",
		NodePoolName:       "worker-pool",
		InstanceType:       "ecs.u1-c1m1.large",
		InstanceChargeType: "PostPaid",
		DesiredSize:        1,
		MinSize:            1,
		MaxSize:            5,
		NodeCount:          1,
		AutoScaling:        true,
		Status:             "active",
		SystemDisk:         &aliyunAckAPI.AckNodeDisk{Category: "cloud_essd", Size: 40, PerformanceLevel: "PL1"},
		Runtime:            &aliyunAckAPI.AckNodeRuntime{Name: "containerd", Version: "1.6.20"},
		ImageType:          "AliyunLinux3",
		VSwitchIds:         []string{"vsw-a", "vsw-b"},
		SecurityGroupId:    "sg-1",
		Labels:             []aliyunAckAPI.AckNodeLabel{{Key: "workload", Value: "api"}},
		Taints:             []aliyunAckAPI.AckNodeTaint{{Key: "spot", Value: "true", Effect: "NoSchedule"}},
		Tags:               []aliyunAckAPI.AckTag{{Key: "pool", Value: "workers"}},
		Management: &aliyunAckAPI.AckNodePoolManagement{
			AutoRepairPolicy: &aliyunAckAPI.AckNodeAutoRepairPolicy{RestartNodes: true},
			AutoUpgradePolicy: &aliyunAckAPI.AckNodeAutoUpgradePolicy{
				RuntimeUpgrade: true,
				OsUpgrade:      true,
				KubeletUpgrade: true,
			},
			AutoVulFixPolicy: &aliyunAckAPI.AckNodeAutoVulFixPolicy{RestartNodes: true, VulLevels: "asap"},
		},
	}
	if err := setAckNodePoolAttributes(d, pool); err != nil {
		t.Fatalf("setAckNodePoolAttributes() error = %v", err)
	}

	assertions := map[string]interface{}{
		"cluster_id":                    "c1a2b3",
		"name":                          "worker-pool",
		"desired_size":                  1,
		"min_size":                      1,
		"max_size":                      5,
		"node_count":                    1,
		"auto_scaling":                  true,
		"status":                        "active",
		"image_type":                    "AliyunLinux3",
		"security_group_id":             "sg-1",
		"runtime_name":                  "containerd",
		"runtime_version":               "1.6.20",
		"system_disk_category":          "cloud_essd",
		"system_disk_size":              40,
		"system_disk_performance_level": "PL1",
	}
	for field, want := range assertions {
		if got := d.Get(field); !reflect.DeepEqual(got, want) {
			t.Fatalf("d.Get(%q) = %#v, want %#v", field, got, want)
		}
	}
	if got := d.Get("instance_types").([]interface{}); !reflect.DeepEqual(got, []interface{}{"ecs.u1-c1m1.large"}) {
		t.Fatalf("instance_types = %v, want [ecs.u1-c1m1.large]", got)
	}
	if got := d.Get("vswitch_ids").([]interface{}); !reflect.DeepEqual(got, []interface{}{"vsw-a", "vsw-b"}) {
		t.Fatalf("vswitch_ids = %v, want [vsw-a vsw-b]", got)
	}
	if got := d.Get("labels").([]interface{}); !reflect.DeepEqual(got, []interface{}{
		map[string]interface{}{"key": "workload", "value": "api"},
	}) {
		t.Fatalf("labels = %v", got)
	}
	if got := d.Get("taints").([]interface{}); !reflect.DeepEqual(got, []interface{}{
		map[string]interface{}{"key": "spot", "value": "true", "effect": "NoSchedule"},
	}) {
		t.Fatalf("taints = %v", got)
	}
	if got := d.Get("tags").(map[string]interface{}); !reflect.DeepEqual(got, map[string]interface{}{"pool": "workers"}) {
		t.Fatalf("tags = %v", got)
	}
	// The detail flavor only surfaces the kubelet upgrade switch.
	management := d.Get("management").([]interface{})
	if len(management) != 1 {
		t.Fatalf("management = %v, want exactly one block", management)
	}
	block := management[0].(map[string]interface{})
	if block["auto_repair"] != true || block["auto_upgrade"] != true || block["auto_vulfix"] != true {
		t.Fatalf("management block = %v", block)
	}

	// A nil pool must be a no-op, not a panic.
	if err := setAckNodePoolAttributes(d, nil); err != nil {
		t.Fatalf("setAckNodePoolAttributes(nil) error = %v", err)
	}
}

// TestCsKubernetesNodePoolCompositeIdFormat pins the id convention
// "clusterId:nodepoolId": Create sets it and Read/Delete parse it back with
// ackParseTwoPartId.
func TestCsKubernetesNodePoolCompositeIdFormat(t *testing.T) {
	const clusterId, nodepoolId = "c1a2b3", "np-42"
	d := schema.TestResourceDataRaw(t, resourceAliCloudAckNodepool().Schema, map[string]interface{}{
		"cluster_id":     clusterId,
		"name":           "worker-pool",
		"instance_types": []interface{}{"ecs.u1-c1m1.large"},
		"vswitch_ids":    []interface{}{"vsw-1"},
	})
	d.SetId(clusterId + ":" + nodepoolId)

	gotCluster, gotPool, err := ackParseTwoPartId(d.Id())
	if err != nil {
		t.Fatalf("ackParseTwoPartId(%q) error = %v", d.Id(), err)
	}
	if gotCluster != clusterId || gotPool != nodepoolId {
		t.Fatalf("ackParseTwoPartId(%q) = (%q, %q), want (%q, %q)", d.Id(), gotCluster, gotPool, clusterId, nodepoolId)
	}
}
