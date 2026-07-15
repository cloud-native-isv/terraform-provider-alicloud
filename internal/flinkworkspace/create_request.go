package flinkworkspace

import (
	"crypto/sha256"
	"fmt"

	aliyunFlinkAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
)

const CreateTokenTagKey = "terraform-create-token"

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
