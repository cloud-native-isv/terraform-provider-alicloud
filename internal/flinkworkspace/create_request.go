package flinkworkspace

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"

	aliyunFlinkAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
)

const (
	CreateTokenTagKey     = "terraform-create-token"
	CreateIntentTagKey    = "terraform-create-intent-v1"
	CapacityIntentLegacy  = "LEGACY"
	CapacityIntentInitial = "INITIAL"
)

type CreateOptions = aliyunFlinkAPI.WorkspaceCreateOptions

// WorkspaceCreateToken returns a stable provider-owned identity token used to
// recover a paid CreateInstance whose response was lost. Capacity is excluded
// deliberately: a user may correct capacity before recovering the same
// identity, and that must not purchase another workspace.
func WorkspaceCreateToken(workspace *aliyunFlinkAPI.Workspace) string {
	if workspace == nil {
		return ""
	}
	identity := fmt.Sprintf("alicloud-flink-workspace-v1\x00%s\x00%s", workspace.Region, workspace.Name)
	return fmt.Sprintf("%x", sha256.Sum256([]byte(identity)))
}

// WorkspaceCreateIntentFingerprint returns a stable hash of the effective
// paid CreateInstance intent. Identity remains the responsibility of
// WorkspaceCreateToken; this fingerprint deliberately excludes name/region
// and includes every API-unobservable purchase option plus billing, topology,
// and capacity mode.
func WorkspaceCreateIntentFingerprint(workspace *aliyunFlinkAPI.Workspace, options CreateOptions, capacityMode string) string {
	if workspace == nil {
		return ""
	}
	var canonical bytes.Buffer
	writeFlinkIntentString(&canonical, "alicloud-flink-workspace-create-intent-v1")
	writeFlinkIntentString(&canonical, workspace.ChargeType)
	writeFlinkIntentString(&canonical, capacityMode)
	writeFlinkIntentResourceSpec(&canonical, workspace.ResourceSpec)
	haEnabled := workspace.HighAvailability != nil && workspace.HighAvailability.Enabled
	writeFlinkIntentBool(&canonical, haEnabled)
	if haEnabled {
		writeFlinkIntentResourceSpec(&canonical, workspace.HighAvailability.ResourceSpec)
	} else {
		writeFlinkIntentResourceSpec(&canonical, nil)
	}
	writeFlinkIntentOptionalBool(&canonical, options.AutoRenew)
	writeFlinkIntentOptionalInt32(&canonical, options.Duration)
	writeFlinkIntentString(&canonical, options.PricingCycle)
	writeFlinkIntentString(&canonical, options.Extra)
	writeFlinkIntentString(&canonical, options.PromotionCode)
	writeFlinkIntentOptionalBool(&canonical, options.UsePromotionCode)
	return fmt.Sprintf("%x", sha256.Sum256(canonical.Bytes()))
}

func writeFlinkIntentString(buffer *bytes.Buffer, value string) {
	_ = binary.Write(buffer, binary.BigEndian, uint64(len(value)))
	_, _ = buffer.WriteString(value)
}

func writeFlinkIntentBool(buffer *bytes.Buffer, value bool) {
	if value {
		_ = buffer.WriteByte(1)
		return
	}
	_ = buffer.WriteByte(0)
}

func writeFlinkIntentOptionalBool(buffer *bytes.Buffer, value *bool) {
	writeFlinkIntentBool(buffer, value != nil)
	if value != nil {
		writeFlinkIntentBool(buffer, *value)
	}
}

func writeFlinkIntentOptionalInt32(buffer *bytes.Buffer, value *int32) {
	writeFlinkIntentBool(buffer, value != nil)
	if value != nil {
		_ = binary.Write(buffer, binary.BigEndian, *value)
	}
}

func writeFlinkIntentResourceSpec(buffer *bytes.Buffer, spec *aliyunFlinkAPI.ResourceSpec) {
	writeFlinkIntentBool(buffer, spec != nil)
	if spec == nil {
		return
	}
	_ = binary.Write(buffer, binary.BigEndian, math.Float64bits(spec.Cpu))
	_ = binary.Write(buffer, binary.BigEndian, math.Float64bits(spec.MemoryGB))
}
