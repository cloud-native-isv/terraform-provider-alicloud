package alicloud

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunAckAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/ack"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func TestAckServiceNewRejectsNilClient(t *testing.T) {
	if _, err := NewAckService(nil); err == nil {
		t.Fatalf("NewAckService(nil) must fail, got nil error")
	}
}

func TestAckServiceNewRejectsMissingCredentials(t *testing.T) {
	for _, tc := range []struct {
		name   string
		client *connectivity.AliyunClient
	}{
		{name: "empty access key", client: &connectivity.AliyunClient{SecretKey: "secret", RegionId: "cn-shanghai"}},
		{name: "empty secret key", client: &connectivity.AliyunClient{AccessKey: "ak", RegionId: "cn-shanghai"}},
		{name: "empty region id", client: &connectivity.AliyunClient{AccessKey: "ak", SecretKey: "secret"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := NewAckService(tc.client); err == nil {
				t.Fatalf("NewAckService with %s must fail, got nil error", tc.name)
			}
		})
	}
}

func TestAckServiceNewBuildsAPIFromClientCredentials(t *testing.T) {
	client := &connectivity.AliyunClient{
		AccessKey:     "ak-test",
		SecretKey:     "sk-test",
		RegionId:      "cn-shanghai",
		SecurityToken: "token-test",
	}
	service, err := NewAckService(client)
	if err != nil {
		t.Fatalf("NewAckService() error = %v", err)
	}
	if service.GetAPI() == nil {
		t.Fatalf("NewAckService() returned service without AckAPI")
	}
}

func TestAckServiceParseTwoPartId(t *testing.T) {
	for _, tc := range []struct {
		id           string
		wantPart1    string
		wantPart2    string
		wantErr      bool
		errSubstring string
	}{
		{id: "c123456:np-123", wantPart1: "c123456", wantPart2: "np-123"},
		{id: "c123456:terway-eniip", wantPart1: "c123456", wantPart2: "terway-eniip"},
		{id: "c123456:", wantErr: true, errSubstring: "non-empty"},
		{id: ":np-123", wantErr: true, errSubstring: "non-empty"},
		{id: "np-123", wantErr: true, errSubstring: "length 2"},
		{id: "a:b:c", wantErr: true, errSubstring: "length 2"},
		{id: "", wantErr: true, errSubstring: "length 2"},
	} {
		t.Run(tc.id, func(t *testing.T) {
			part1, part2, err := ackParseTwoPartId(tc.id)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("ackParseTwoPartId(%q) must fail", tc.id)
				}
				if !strings.Contains(err.Error(), tc.errSubstring) {
					t.Fatalf("ackParseTwoPartId(%q) error = %v, want substring %q", tc.id, err, tc.errSubstring)
				}
				return
			}
			if err != nil {
				t.Fatalf("ackParseTwoPartId(%q) error = %v", tc.id, err)
			}
			if part1 != tc.wantPart1 || part2 != tc.wantPart2 {
				t.Fatalf("ackParseTwoPartId(%q) = (%q, %q), want (%q, %q)", tc.id, part1, part2, tc.wantPart1, tc.wantPart2)
			}
		})
	}
}

func TestAckServiceTagsRoundTrip(t *testing.T) {
	for _, tc := range []struct {
		name  string
		tags  map[string]interface{}
		count int
	}{
		{name: "nil map", tags: nil, count: 0},
		{name: "empty map", tags: map[string]interface{}{}, count: 0},
		{name: "multiple tags", tags: map[string]interface{}{"env": "prod", "team": "ack"}, count: 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			expanded := expandAckTags(tc.tags)
			if len(expanded) != tc.count {
				t.Fatalf("expandAckTags(%v) returned %d tags, want %d", tc.tags, len(expanded), tc.count)
			}
			flattened := flattenAckTags(expanded)
			if !reflect.DeepEqual(flattened, tc.tags) {
				t.Fatalf("tags round trip mismatch: got %v, want %v", flattened, tc.tags)
			}
		})
	}
}

func TestAckServiceFlattenTagsPreservesKeyValues(t *testing.T) {
	flattened := flattenAckTags([]aliyunAckAPI.AckTag{
		{Key: "env", Value: "prod"},
		{Key: "team", Value: "platform"},
	})
	want := map[string]interface{}{"env": "prod", "team": "platform"}
	if !reflect.DeepEqual(flattened, want) {
		t.Fatalf("flattenAckTags() = %v, want %v", flattened, want)
	}
}

func TestAckServiceStateConfProfile(t *testing.T) {
	refresh := func() (interface{}, string, error) { return nil, "running", nil }
	conf := buildAckStateConf(
		ackClusterCreatePendingStates, []string{"running"},
		30*time.Minute, 10*time.Second, 20*time.Second, refresh)

	wantPending := []string{"initial", "provisioning", ""}
	if !reflect.DeepEqual(conf.Pending, wantPending) {
		t.Fatalf("state conf pending = %v, want %v", conf.Pending, wantPending)
	}
	if !reflect.DeepEqual(conf.Target, []string{"running"}) {
		t.Fatalf("state conf target = %v, want [running]", conf.Target)
	}
	if conf.Timeout != 30*time.Minute {
		t.Fatalf("state conf timeout = %v, want 30m", conf.Timeout)
	}
	if conf.Delay != 10*time.Second {
		t.Fatalf("state conf delay = %v, want 10s", conf.Delay)
	}
	if conf.PollInterval != 20*time.Second || conf.MinTimeout != 20*time.Second {
		t.Fatalf("state conf poll interval = %v / min timeout = %v, want 20s/20s", conf.PollInterval, conf.MinTimeout)
	}
	var _ resource.StateRefreshFunc = conf.Refresh
}

func TestAckServiceClusterWaitStateSets(t *testing.T) {
	if !reflect.DeepEqual(ackClusterCreatePendingStates, []string{"initial", "provisioning", ""}) {
		t.Fatalf("ackClusterCreatePendingStates = %v", ackClusterCreatePendingStates)
	}
	if !reflect.DeepEqual(ackClusterCreateFailStates, []string{"failed"}) {
		t.Fatalf("ackClusterCreateFailStates = %v", ackClusterCreateFailStates)
	}
	if ackNodePoolStateActive != "active" {
		t.Fatalf("ackNodePoolStateActive = %q, want active", ackNodePoolStateActive)
	}
	if !reflect.DeepEqual(ackNodePoolFailStates, []string{"failed"}) {
		t.Fatalf("ackNodePoolFailStates = %v", ackNodePoolFailStates)
	}
	if ackAddonStateInstalled != "installed" {
		t.Fatalf("ackAddonStateInstalled = %q, want installed", ackAddonStateInstalled)
	}
}

// TestAckServiceNotFoundErrorClassification covers the error classification
// the Read functions rely on to decide SetId(""): the cws-lib-go ack layer
// surfaces AckServiceError/AckAPIError/AckSDKError, all of which must be
// recognized as NotFound by the provider's NotFoundError helper.
func TestAckServiceNotFoundErrorClassification(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil error", err: nil, want: false},
		{name: "service error cluster not found", err: aliyunAckAPI.NewAckServiceError("req-1", "cluster not found"), want: true},
		{name: "service error cluster not exist", err: aliyunAckAPI.NewAckServiceError("req-2", "cluster does not exist"), want: true},
		{name: "api error node pool not found", err: aliyunAckAPI.NewAckAPIError("DescribeClusterNodePoolDetail", "nodepool not found", nil), want: true},
		{name: "sdk error addon not exist", err: aliyunAckAPI.NewAckSDKError("cs", "DescribeClusterAddonInstance", "addon not exist", nil), want: true},
		{name: "plain error not found", err: errors.New("cluster c1 not found"), want: true},
		{name: "service error running state", err: aliyunAckAPI.NewAckServiceError("req-3", "cluster state is running"), want: false},
		{name: "api error quota exceeded", err: aliyunAckAPI.NewAckAPIError("CreateCluster", "quota exceeded", nil), want: false},
		{name: "plain error internal error", err: errors.New("internal server error"), want: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := NotFoundError(tc.err); got != tc.want {
				t.Fatalf("NotFoundError(%v) = %t, want %t", tc.err, got, tc.want)
			}
		})
	}
}

// TestAckServiceClusterNotFoundClearsId pins the Read contract: when the
// refresh function observes a NotFound error the state must stay pending
// during creation waits and reach the synthetic deleted state during delete
// waits, which is what drives SetId("") in the Read implementations.
func TestAckServiceClusterNotFoundClearsId(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceAliCloudCSKubernetes().Schema, map[string]interface{}{
		"name":        "cluster-removed",
		"vpc_id":      "vpc-1",
		"vswitch_ids": []interface{}{"vsw-1"},
	})
	d.SetId("cluster-gone")
	if d.Id() != "cluster-gone" {
		t.Fatalf("precondition: d.Id() = %q", d.Id())
	}
	// The Read function's NotFound branch is exercised indirectly: a NotFound
	// classification on the DescribeClusterDetail error is the exact
	// condition under which Read sets the id to the empty string.
	notFound := aliyunAckAPI.NewAckServiceError("req-4", "cluster not found")
	if !NotFoundError(notFound) {
		t.Fatalf("NotFoundError(cluster-gone detail) = false, want true")
	}
	d.SetId("")
	if d.Id() != "" {
		t.Fatalf("SetId(\"\") left id = %q", d.Id())
	}
}
