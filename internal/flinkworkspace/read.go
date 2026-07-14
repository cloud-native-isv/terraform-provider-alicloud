package flinkworkspace

import aliyunFlinkAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"

// HAConfig flattens the HA fields returned by DescribeInstances into the
// shape used by the alicloud_flink_workspace schema.
func HAConfig(workspace *aliyunFlinkAPI.Workspace) (map[string]interface{}, bool) {
	if workspace == nil {
		return nil, false
	}

	if ha := workspace.HighAvailability; ha != nil && ha.Enabled {
		return haConfig(ha.ZoneId, ha.VSwitchIds, ha.ResourceSpec), true
	}
	if workspace.Ha {
		return haConfig(workspace.HaZoneId, workspace.HaVSwitchIds, workspace.HaResourceSpec), true
	}

	return nil, false
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
