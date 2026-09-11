package alicloud

import (
	"reflect"
	"testing"
	"time"

	aliyunAckAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/ack"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestCsKubernetesSchemaContract(t *testing.T) {
	resource := resourceAliCloudCSKubernetes()
	if err := resource.InternalValidate(nil, true); err != nil {
		t.Fatalf("InternalValidate() error = %v", err)
	}

	for _, name := range []string{"name", "vpc_id", "vswitch_ids"} {
		field, ok := resource.Schema[name]
		if !ok || !field.Required {
			t.Fatalf("field %q must be Required: %#v", name, field)
		}
		if !field.ForceNew {
			t.Fatalf("field %q must be ForceNew", name)
		}
	}
	vswitchIds := resource.Schema["vswitch_ids"]
	if vswitchIds.Type != schema.TypeList || vswitchIds.MinItems != 1 {
		t.Fatalf("vswitch_ids schema = %#v", vswitchIds)
	}

	// Defaults per the frozen schema contract.
	clusterType := resource.Schema["cluster_type"]
	if !clusterType.Optional || !clusterType.ForceNew || clusterType.Default != "Kubernetes" {
		t.Fatalf("cluster_type schema = %#v", clusterType)
	}
	profile := resource.Schema["profile"]
	if !profile.Optional || !profile.ForceNew || profile.Default != ackClusterProfileManaged {
		t.Fatalf("profile schema = %#v", profile)
	}
	if profile.ValidateFunc == nil {
		t.Fatalf("profile must restrict values to Managed|Default")
	}
	if _, errs := profile.ValidateFunc("Managed", "profile"); len(errs) != 0 {
		t.Fatalf("profile=Managed rejected: %v", errs)
	}
	if _, errs := profile.ValidateFunc("Default", "profile"); len(errs) != 0 {
		t.Fatalf("profile=Default rejected: %v", errs)
	}
	if _, errs := profile.ValidateFunc("Serverless", "profile"); len(errs) == 0 {
		t.Fatalf("profile=Serverless must be rejected")
	}

	snatEntry := resource.Schema["snat_entry"]
	if !snatEntry.Optional || !snatEntry.ForceNew || snatEntry.Default != true {
		t.Fatalf("snat_entry schema = %#v", snatEntry)
	}
	nodeCidrMask := resource.Schema["node_cidr_mask"]
	if nodeCidrMask.Default != 24 || !nodeCidrMask.ForceNew {
		t.Fatalf("node_cidr_mask schema = %#v", nodeCidrMask)
	}
	if _, errs := nodeCidrMask.ValidateFunc(28, "node_cidr_mask"); len(errs) != 0 {
		t.Fatalf("node_cidr_mask=28 rejected: %v", errs)
	}
	if _, errs := nodeCidrMask.ValidateFunc(23, "node_cidr_mask"); len(errs) == 0 {
		t.Fatalf("node_cidr_mask=23 must be rejected")
	}

	// Updatable (no ForceNew) per the contract: deletion_protection,
	// resource_group_id and tags; retain_all_resources only affects delete.
	for _, name := range []string{"deletion_protection", "resource_group_id", "tags", "retain_all_resources"} {
		field, ok := resource.Schema[name]
		if !ok || field.ForceNew {
			t.Fatalf("field %q must be updatable: %#v", name, field)
		}
	}
	if resource.Schema["tags"].Type != schema.TypeMap {
		t.Fatalf("tags must be a map: %#v", resource.Schema["tags"])
	}

	// Immutable fields must be ForceNew.
	for _, name := range []string{"cluster_spec", "container_cidr", "service_cidr", "kubernetes_version", "security_group_id", "worker_instance_type", "num_of_workers"} {
		field, ok := resource.Schema[name]
		if !ok || !field.ForceNew {
			t.Fatalf("field %q must be ForceNew: %#v", name, field)
		}
	}

	// Computed attributes.
	for _, name := range []string{"state", "version", "master_url", "subnet_cidr", "worker_ram_role_name", "zone_id", "node_count"} {
		field, ok := resource.Schema[name]
		if !ok || !field.Computed || field.Optional || field.Required {
			t.Fatalf("field %q must be Computed-only: %#v", name, field)
		}
	}

	// Timeouts: create/delete 30m, update 60m per the contract.
	if resource.Timeouts == nil || resource.Timeouts.Create == nil || *resource.Timeouts.Create != 30*time.Minute {
		t.Fatalf("create timeout must be 30m: %#v", resource.Timeouts)
	}
	if resource.Timeouts.Delete == nil || *resource.Timeouts.Delete != 30*time.Minute {
		t.Fatalf("delete timeout must be 30m: %#v", resource.Timeouts)
	}
	if resource.Timeouts.Update == nil || *resource.Timeouts.Update != 60*time.Minute {
		t.Fatalf("update timeout must be 60m: %#v", resource.Timeouts)
	}
}

func TestCsKubernetesBuildCreateRequest(t *testing.T) {
	for _, tc := range []struct {
		name   string
		config map[string]interface{}
		region string
		want   *aliyunAckAPI.AckClusterCreate
	}{
		{
			name: "full managed pro cluster",
			config: map[string]interface{}{
				"name":                        "prod-cluster",
				"cluster_type":                "Kubernetes",
				"profile":                     "Managed",
				"cluster_spec":                "ack.pro.small",
				"vpc_id":                      "vpc-abc",
				"vswitch_ids":                 []interface{}{"vsw-1", "vsw-2"},
				"container_cidr":              "10.0.0.0/16",
				"service_cidr":                "172.21.0.0/20",
				"kubernetes_version":          "1.31.1-aliyun.1",
				"security_group_id":           "sg-1",
				"node_cidr_mask":              26,
				"snat_entry":                  false,
				"worker_instance_type":        "ecs.u1-c1m1.large",
				"worker_system_disk_category": "cloud_essd",
				"worker_system_disk_size":     40,
				"num_of_workers":              2,
				"deletion_protection":         true,
				"resource_group_id":           "rg-1",
				"tags":                        map[string]interface{}{"env": "prod"},
			},
			region: "cn-shanghai",
			want: &aliyunAckAPI.AckClusterCreate{
				Name:                     "prod-cluster",
				RegionId:                 "cn-shanghai",
				ClusterType:              "Kubernetes",
				Profile:                  "Managed",
				ClusterSpec:              "ack.pro.small",
				VpcId:                    "vpc-abc",
				VSwitchIds:               []string{"vsw-1", "vsw-2"},
				ContainerCidr:            "10.0.0.0/16",
				ServiceCidr:              "172.21.0.0/20",
				KubernetesVersion:        "1.31.1-aliyun.1",
				SecurityGroupId:          "sg-1",
				NodeCidrMask:             "26",
				SnatEntry:                false,
				WorkerInstanceType:       "ecs.u1-c1m1.large",
				WorkerSystemDiskCategory: "cloud_essd",
				WorkerSystemDiskSize:     40,
				NumOfWorkers:             2,
				Tags:                     []aliyunAckAPI.AckTag{{Key: "env", Value: "prod"}},
			},
		},
		{
			name: "minimal config with schema defaults",
			config: map[string]interface{}{
				"name":        "minimal",
				"vpc_id":      "vpc-min",
				"vswitch_ids": []interface{}{"vsw-min"},
			},
			region: "cn-beijing",
			want: &aliyunAckAPI.AckClusterCreate{
				Name:         "minimal",
				RegionId:     "cn-beijing",
				ClusterType:  "Kubernetes",
				Profile:      "Managed",
				VpcId:        "vpc-min",
				VSwitchIds:   []string{"vsw-min"},
				NodeCidrMask: "24",
				SnatEntry:    true,
				// expandAckTags always yields a (possibly empty) slice.
				Tags: []aliyunAckAPI.AckTag{},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, resourceAliCloudCSKubernetes().Schema, tc.config)
			got := buildAckClusterCreate(d, tc.region)
			if got.Name != tc.want.Name || got.RegionId != tc.want.RegionId ||
				got.ClusterType != tc.want.ClusterType || got.Profile != tc.want.Profile ||
				got.ClusterSpec != tc.want.ClusterSpec || got.VpcId != tc.want.VpcId ||
				got.ContainerCidr != tc.want.ContainerCidr || got.ServiceCidr != tc.want.ServiceCidr ||
				got.KubernetesVersion != tc.want.KubernetesVersion || got.SecurityGroupId != tc.want.SecurityGroupId ||
				got.NodeCidrMask != tc.want.NodeCidrMask || got.SnatEntry != tc.want.SnatEntry ||
				got.WorkerInstanceType != tc.want.WorkerInstanceType ||
				got.WorkerSystemDiskCategory != tc.want.WorkerSystemDiskCategory ||
				got.WorkerSystemDiskSize != tc.want.WorkerSystemDiskSize ||
				got.NumOfWorkers != tc.want.NumOfWorkers {
				t.Fatalf("buildAckClusterCreate() = %+v, want %+v", got, tc.want)
			}
			if !reflect.DeepEqual(got.VSwitchIds, tc.want.VSwitchIds) {
				t.Fatalf("VSwitchIds = %v, want %v", got.VSwitchIds, tc.want.VSwitchIds)
			}
			if !reflect.DeepEqual(got.Tags, tc.want.Tags) {
				t.Fatalf("Tags = %v, want %v", got.Tags, tc.want.Tags)
			}
		})
	}
}

func TestCsKubernetesBuildModifyRequest(t *testing.T) {
	for _, tc := range []struct {
		name            string
		config          map[string]interface{}
		wantDeletionPtr *bool
		wantResourceGrp string
	}{
		{
			name: "no updatable field set leaves deletion protection untouched",
			config: map[string]interface{}{
				"name":        "cluster",
				"vpc_id":      "vpc-1",
				"vswitch_ids": []interface{}{"vsw-1"},
			},
			wantDeletionPtr: nil,
			wantResourceGrp: "",
		},
		{
			name: "deletion protection change is a tri-state pointer",
			config: map[string]interface{}{
				"name":                "cluster",
				"vpc_id":              "vpc-1",
				"vswitch_ids":         []interface{}{"vsw-1"},
				"deletion_protection": true,
			},
			wantDeletionPtr: boolPtr(true),
			wantResourceGrp: "",
		},
		{
			name: "resource group is passed through when set",
			config: map[string]interface{}{
				"name":              "cluster",
				"vpc_id":            "vpc-1",
				"vswitch_ids":       []interface{}{"vsw-1"},
				"resource_group_id": "rg-42",
			},
			wantDeletionPtr: nil,
			wantResourceGrp: "rg-42",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, resourceAliCloudCSKubernetes().Schema, tc.config)
			got := buildAckClusterModify(d)
			if !reflect.DeepEqual(got.DeletionProtection, tc.wantDeletionPtr) {
				t.Fatalf("DeletionProtection = %v, want %v", got.DeletionProtection, tc.wantDeletionPtr)
			}
			if got.ResourceGroupId != tc.wantResourceGrp {
				t.Fatalf("ResourceGroupId = %q, want %q", got.ResourceGroupId, tc.wantResourceGrp)
			}
		})
	}
}

func boolPtr(v bool) *bool { return &v }

func TestCsKubernetesSetAttributesFromDetail(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceAliCloudCSKubernetes().Schema, map[string]interface{}{
		"name":        "will-be-overwritten",
		"vpc_id":      "vpc-1",
		"vswitch_ids": []interface{}{"vsw-old"},
	})
	cluster := &aliyunAckAPI.AckCluster{
		ClusterId:          "c1a2b3",
		Name:               "prod-cluster",
		ClusterType:        "Kubernetes",
		Profile:            "Managed",
		ClusterSpec:        "ack.pro.small",
		State:              "running",
		CurrentVersion:     "1.31.1-aliyun.1",
		MasterUrl:          "https://1.2.3.4:6443",
		VpcId:              "vpc-live",
		VswitchIds:         []string{"vsw-a", "vsw-b"},
		ContainerCidr:      "10.0.0.0/16",
		ServiceCidr:        "172.21.0.0/20",
		SubnetCidr:         "10.0.0.0/24",
		SecurityGroupId:    "sg-live",
		ResourceGroupId:    "rg-live",
		DeletionProtection: true,
		WorkerRamRoleName:  "aliyunmanagedcsrole",
		ZoneId:             "cn-shanghai-a",
		Size:               3,
		Tags:               []aliyunAckAPI.AckTag{{Key: "env", Value: "prod"}},
	}
	if err := setAckClusterAttributes(d, cluster); err != nil {
		t.Fatalf("setAckClusterAttributes() error = %v", err)
	}

	assertions := map[string]interface{}{
		"name":                 "prod-cluster",
		"cluster_type":         "Kubernetes",
		"profile":              "Managed",
		"cluster_spec":         "ack.pro.small",
		"state":                "running",
		"version":              "1.31.1-aliyun.1",
		"kubernetes_version":   "1.31.1-aliyun.1",
		"master_url":           "https://1.2.3.4:6443",
		"vpc_id":               "vpc-live",
		"container_cidr":       "10.0.0.0/16",
		"service_cidr":         "172.21.0.0/20",
		"subnet_cidr":          "10.0.0.0/24",
		"security_group_id":    "sg-live",
		"resource_group_id":    "rg-live",
		"deletion_protection":  true,
		"worker_ram_role_name": "aliyunmanagedcsrole",
		"zone_id":              "cn-shanghai-a",
		"node_count":           3,
	}
	for field, want := range assertions {
		if got := d.Get(field); !reflect.DeepEqual(got, want) {
			t.Fatalf("d.Get(%q) = %#v, want %#v", field, got, want)
		}
	}
	if got := d.Get("vswitch_ids").([]interface{}); !reflect.DeepEqual(got, []interface{}{"vsw-a", "vsw-b"}) {
		t.Fatalf("d.Get(vswitch_ids) = %v, want [vsw-a vsw-b]", got)
	}
	if got := d.Get("tags").(map[string]interface{}); !reflect.DeepEqual(got, map[string]interface{}{"env": "prod"}) {
		t.Fatalf("d.Get(tags) = %v, want map[env:prod]", got)
	}

	// A nil cluster must be a no-op, not a panic.
	if err := setAckClusterAttributes(d, nil); err != nil {
		t.Fatalf("setAckClusterAttributes(nil) error = %v", err)
	}
}

func TestCsKubernetesDiffAppliesSchemaDefaults(t *testing.T) {
	res := resourceAliCloudCSKubernetes()
	diff, err := res.Diff(nil, terraform.NewResourceConfigRaw(map[string]interface{}{
		"name":        "cluster",
		"vpc_id":      "vpc-1",
		"vswitch_ids": []interface{}{"vsw-1"},
	}), nil)
	if err != nil {
		t.Fatalf("Diff() error = %v", err)
	}
	for _, field := range []string{"cluster_type", "profile", "snat_entry", "node_cidr_mask"} {
		if _, ok := diff.Attributes[field]; !ok {
			t.Fatalf("Diff() must materialize default for %q, attributes = %#v", field, diff.Attributes)
		}
	}
}
