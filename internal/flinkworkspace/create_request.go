package flinkworkspace

import (
	"encoding/json"
	"fmt"

	aliyunFlinkAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
)

// CreateOptions contains CreateInstance inputs that are not represented by
// cws-lib-go's Workspace type.
type CreateOptions struct {
	AutoRenew        bool
	Duration         int
	PricingCycle     string
	Extra            string
	MonitorType      string
	PromotionCode    string
	UsePromotionCode bool
}

// BuildCreateInstanceBody builds the form body expected by the
// foasconsole/2021-10-28 CreateInstance RPC API.
func BuildCreateInstanceBody(workspace *aliyunFlinkAPI.Workspace, options CreateOptions) (map[string]interface{}, error) {
	if workspace == nil {
		return nil, fmt.Errorf("workspace cannot be nil")
	}

	body := map[string]interface{}{
		"Region":           workspace.Region,
		"InstanceName":     workspace.Name,
		"ChargeType":       workspace.ChargeType,
		"ArchitectureType": workspace.ArchitectureType,
		"AutoRenew":        options.AutoRenew,
		"Duration":         options.Duration,
		"UsePromotionCode": options.UsePromotionCode,
	}

	setString(body, "VpcId", workspace.VpcId)
	setString(body, "ResourceGroupId", workspace.ResourceGroupId)
	setString(body, "PricingCycle", options.PricingCycle)
	setString(body, "Extra", options.Extra)
	setString(body, "MonitorType", options.MonitorType)
	setString(body, "PromotionCode", options.PromotionCode)

	if err := setJSON(body, "VSwitchIds", workspace.VSwitchIds); err != nil {
		return nil, err
	}
	if workspace.ResourceSpec != nil {
		if err := setJSON(body, "ResourceSpec", resourceSpec(workspace.ResourceSpec)); err != nil {
			return nil, err
		}
	}
	if workspace.Storage != nil && workspace.Storage.Oss != nil && workspace.Storage.Oss.Bucket != "" {
		storage := map[string]interface{}{
			"Oss": map[string]string{"Bucket": workspace.Storage.Oss.Bucket},
		}
		if workspace.Storage.FullyManaged {
			storage["FullyManaged"] = true
		}
		if err := setJSON(body, "Storage", storage); err != nil {
			return nil, err
		}
	}

	ha := workspace.HighAvailability
	if ha != nil && ha.Enabled {
		body["Ha"] = true
		if err := setJSON(body, "HaVSwitchIds", ha.VSwitchIds); err != nil {
			return nil, err
		}
		if ha.ResourceSpec != nil {
			if err := setJSON(body, "HaResourceSpec", resourceSpec(ha.ResourceSpec)); err != nil {
				return nil, err
			}
		}
	}

	return body, nil
}

// InstanceIDFromCreateResponse returns the workspace ID from a CreateInstance
// response body.
func InstanceIDFromCreateResponse(response map[string]interface{}) (string, error) {
	orderInfo, ok := response["OrderInfo"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("CreateInstance response does not contain OrderInfo")
	}
	instanceID, ok := orderInfo["InstanceId"].(string)
	if !ok || instanceID == "" {
		return "", fmt.Errorf("CreateInstance response does not contain InstanceId")
	}
	return instanceID, nil
}

func resourceSpec(spec *aliyunFlinkAPI.ResourceSpec) map[string]int32 {
	return map[string]int32{
		"Cpu":      int32(spec.Cpu),
		"MemoryGB": int32(spec.MemoryGB),
	}
}

func setString(body map[string]interface{}, key, value string) {
	if value != "" {
		body[key] = value
	}
}

func setJSON(body map[string]interface{}, key string, value interface{}) error {
	encoded, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("encode %s: %w", key, err)
	}
	body[key] = string(encoded)
	return nil
}
