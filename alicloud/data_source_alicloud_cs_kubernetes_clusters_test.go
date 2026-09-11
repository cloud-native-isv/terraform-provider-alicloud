package alicloud

import (
	"reflect"
	"testing"

	aliyunAckAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/ack"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func TestCsKubernetesClustersSchemaContract(t *testing.T) {
	dataSource := dataSourceAliCloudCSKubernetesClusters()
	if err := dataSource.InternalValidate(nil, false); err != nil {
		t.Fatalf("InternalValidate() error = %v", err)
	}

	nameRegex := dataSource.Schema["name_regex"]
	if !nameRegex.Optional || nameRegex.ValidateFunc == nil {
		t.Fatalf("name_regex must be Optional with regexp validation: %#v", nameRegex)
	}
	if _, errs := nameRegex.ValidateFunc("^prod-.*$", "name_regex"); len(errs) != 0 {
		t.Fatalf("valid name_regex rejected: %v", errs)
	}
	if _, errs := nameRegex.ValidateFunc("(unclosed", "name_regex"); len(errs) == 0 {
		t.Fatalf("invalid name_regex must be rejected")
	}

	enableDetails := dataSource.Schema["enable_details"]
	if !enableDetails.Optional || enableDetails.Default != false {
		t.Fatalf("enable_details schema = %#v", enableDetails)
	}
	for _, name := range []string{"cluster_type", "profile", "output_file"} {
		if _, ok := dataSource.Schema[name]; !ok {
			t.Fatalf("field %q missing from schema", name)
		}
	}
	for _, name := range []string{"ids", "names", "clusters"} {
		field, ok := dataSource.Schema[name]
		if !ok || !field.Computed {
			t.Fatalf("field %q must be Computed: %#v", name, field)
		}
	}
}

func TestCsKubernetesClustersBuildQuery(t *testing.T) {
	for _, tc := range []struct {
		name   string
		config map[string]interface{}
		want   *aliyunAckAPI.AckClusterQuery
	}{
		{
			name:   "no filter yields empty query",
			config: map[string]interface{}{},
			want:   &aliyunAckAPI.AckClusterQuery{},
		},
		{
			name: "cluster type is a server-side filter",
			config: map[string]interface{}{
				"cluster_type": "Kubernetes",
			},
			want: &aliyunAckAPI.AckClusterQuery{ClusterType: "Kubernetes"},
		},
		{
			name: "profile filter",
			config: map[string]interface{}{
				"profile": "Managed",
			},
			want: &aliyunAckAPI.AckClusterQuery{Profile: "Managed"},
		},
		{
			name: "both filters",
			config: map[string]interface{}{
				"cluster_type": "Kubernetes",
				"profile":      "Default",
			},
			want: &aliyunAckAPI.AckClusterQuery{ClusterType: "Kubernetes", Profile: "Default"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, dataSourceAliCloudCSKubernetesClusters().Schema, tc.config)
			if got := buildAckClusterQuery(d); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("buildAckClusterQuery() = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func ackClustersFixture() []aliyunAckAPI.AckCluster {
	return []aliyunAckAPI.AckCluster{
		{ClusterId: "c-prod-1", Name: "prod-shanghai", State: "running", ClusterType: "Kubernetes", Profile: "Managed"},
		{ClusterId: "c-dev-1", Name: "dev-shanghai", State: "running", ClusterType: "Kubernetes", Profile: "Managed"},
		{ClusterId: "c-edge-1", Name: "edge-beijing", State: "running", ClusterType: "Kubernetes", Profile: "Default"},
	}
}

func TestCsKubernetesClustersNameRegexFilter(t *testing.T) {
	for _, tc := range []struct {
		name      string
		nameRegex string
		wantIds   []string
		wantErr   bool
	}{
		{name: "empty regex passes everything through", nameRegex: "", wantIds: []string{"c-prod-1", "c-dev-1", "c-edge-1"}},
		{name: "prefix match", nameRegex: "^prod-", wantIds: []string{"c-prod-1"}},
		{name: "substring match", nameRegex: "shanghai", wantIds: []string{"c-prod-1", "c-dev-1"}},
		{name: "no match", nameRegex: "^staging-", wantIds: []string{}},
		{name: "invalid regex", nameRegex: "(unclosed", wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := filterAckClustersByNameRegex(ackClustersFixture(), tc.nameRegex)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("filterAckClustersByNameRegex(%q) must fail", tc.nameRegex)
				}
				return
			}
			if err != nil {
				t.Fatalf("filterAckClustersByNameRegex(%q) error = %v", tc.nameRegex, err)
			}
			ids := make([]string, 0, len(got))
			for _, cluster := range got {
				ids = append(ids, cluster.ClusterId)
			}
			if !reflect.DeepEqual(ids, tc.wantIds) {
				t.Fatalf("filtered ids = %v, want %v", ids, tc.wantIds)
			}
		})
	}
}

func TestCsKubernetesClustersFlatten(t *testing.T) {
	clusters := []aliyunAckAPI.AckCluster{
		{
			ClusterId:       "c-prod-1",
			Name:            "prod-shanghai",
			State:           "running",
			ClusterSpec:     "ack.pro.small",
			ClusterType:     "Kubernetes",
			Profile:         "Managed",
			CurrentVersion:  "1.31.1-aliyun.1",
			VpcId:           "vpc-1",
			VswitchIds:      []string{"vsw-1", "vsw-2"},
			ResourceGroupId: "rg-1",
			Tags:            []aliyunAckAPI.AckTag{{Key: "env", Value: "prod"}},
		},
	}
	flattened := flattenAckClusters(clusters)
	if len(flattened) != 1 {
		t.Fatalf("flattenAckClusters() returned %d blocks, want 1", len(flattened))
	}
	block := flattened[0].(map[string]interface{})
	want := map[string]interface{}{
		"cluster_id":        "c-prod-1",
		"name":              "prod-shanghai",
		"state":             "running",
		"cluster_spec":      "ack.pro.small",
		"cluster_type":      "Kubernetes",
		"profile":           "Managed",
		"version":           "1.31.1-aliyun.1",
		"vpc_id":            "vpc-1",
		"vswitch_ids":       []interface{}{"vsw-1", "vsw-2"},
		"resource_group_id": "rg-1",
		"tags":              map[string]interface{}{"env": "prod"},
	}
	for field, wantValue := range want {
		if got := block[field]; !reflect.DeepEqual(got, wantValue) {
			t.Fatalf("clusters block field %q = %#v, want %#v", field, got, wantValue)
		}
	}

	if got := flattenAckClusters(nil); len(got) != 0 {
		t.Fatalf("flattenAckClusters(nil) = %v, want empty", got)
	}
}
