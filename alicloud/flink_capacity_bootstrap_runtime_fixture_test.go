//go:build flink_capacity_bootstrap_runtime_fixture
// +build flink_capacity_bootstrap_runtime_fixture

package alicloud

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/internal/flinkworkspace"
	flink "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func TestFlinkCapacityBootstrapRuntimeFixtureMatchesIDOnlyProductionCreateShape(t *testing.T) {
	root := t.TempDir()
	ledger := filepath.Join(root, "ledger.json")
	writeFlinkCapacityBootstrapRuntimeLedger(t, ledger, flinkCapacityBootstrapTofuLedger{DelayedVisibility: true})
	service := &flinkCapacityBootstrapRuntimeService{ledgerPath: ledger}
	token := strings.Repeat("a", 64)
	fingerprint := strings.Repeat("b", 64)
	request := &flink.Workspace{
		Name: "bootstrap-runtime-parent", Region: "cn-test", ResourceGroupId: "rg-fixture", VpcId: "vpc-fixture",
		VSwitchIds: []string{"vsw-fixture"}, ChargeType: "PRE", ArchitectureType: "X86",
		Tags: []flink.Tag{
			{Key: flinkworkspace.CreateTokenTagKey, Value: token},
			{Key: flinkworkspace.CreateIntentTagKey, Value: fingerprint},
		},
	}
	created, err := service.CreateInstance(request, flinkworkspace.CreateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if created.Id == "" || created.ResourceId != "" || len(created.Tags) != 0 {
		t.Fatalf("fixture Create shape = Id:%q ResourceIdPresent:%t Tags:%d, want ID only", created.Id, created.ResourceId != "", len(created.Tags))
	}
	data := schema.TestResourceDataRaw(t, resourceAliCloudFlinkWorkspace().Schema, nil)
	_ = data.Set("terraform_create_token", token)
	_ = data.Set("create_intent_fingerprint", fingerprint)
	request.Id = created.Id
	if err := completeFlinkWorkspaceCreateWithCapacityBootstrapContext(data, request, true, time.Now()); err != nil {
		t.Fatal(err)
	}
	if _, ok := data.GetOk("capacity_bootstrap_context"); !ok {
		t.Fatal("fixture ID-only Create did not persist context before returning")
	}
	state := readFlinkCapacityBootstrapRuntimeLedger(t, ledger)
	if state.ParentIdentityLists != 0 {
		t.Fatalf("paid Create performed %d post-ID list call(s), want immediate return", state.ParentIdentityLists)
	}
}
