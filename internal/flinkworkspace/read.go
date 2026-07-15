package flinkworkspace

import aliyunFlinkAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"

func HAConfigWithFallback(workspace *aliyunFlinkAPI.Workspace, fallbackZoneID string) (map[string]interface{}, bool) {
	if workspace == nil {
		return nil, false
	}

	if ha := workspace.HighAvailability; ha != nil && ha.Enabled {
		zoneID := firstNonEmpty(ha.ZoneId, zoneFromVSwitchInfo(workspace.HaVSwitchInfo), fallbackZoneID)
		return haConfig(zoneID, ha.VSwitchIds, ha.ResourceSpec), true
	}
	if workspace.Ha {
		zoneID := firstNonEmpty(workspace.HaZoneId, zoneFromVSwitchInfo(workspace.HaVSwitchInfo), fallbackZoneID)
		return haConfig(zoneID, workspace.HaVSwitchIds, workspace.HaResourceSpec), true
	}

	return nil, false
}

func PrimaryZoneID(workspace *aliyunFlinkAPI.Workspace, fallbackZoneID string) string {
	if workspace == nil {
		return fallbackZoneID
	}
	return firstNonEmpty(workspace.ZoneId, zoneFromVSwitchInfo(workspace.VSwitchInfo), fallbackZoneID)
}

func haConfig(zoneID string, vSwitchIDs []string, spec *aliyunFlinkAPI.ResourceSpec) map[string]interface{} {
	config := map[string]interface{}{
		"zone_id":     zoneID,
		"vswitch_ids": vSwitchIDs,
	}
	if spec != nil {
		config["resource"] = map[string]interface{}{
			"cpu":    int(spec.Cpu),
			"memory": int(spec.MemoryGB),
		}
	}
	return config
}

func zoneFromVSwitchInfo(items []aliyunFlinkAPI.VSwitchInfo) string {
	for _, item := range items {
		if item.ZoneId != "" {
			return item.ZoneId
		}
	}
	return ""
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
