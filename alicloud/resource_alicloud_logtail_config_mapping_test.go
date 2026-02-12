package alicloud

import (
	"testing"

	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/sls"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func TestExpandSlsLogtailPipelineConfig_ProductionSample(t *testing.T) {
	r := resourceAliCloudLogtailConfig()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]interface{}{
		"project": "test-project",
		"name":    "kangaroo-pai-file-fabricmanager-proxy",
		"inputs": []interface{}{
			map[string]interface{}{
				"type":        "input_file",
				"config_json": `{"AllowingIncludedByMultiConfigs":true,"EnableContainerDiscovery":false,"FileEncoding":"utf8","FilePaths":["/logtail_host/var/log/fabricmanager-proxy/**/fabric.log"],"MaxDirSearchDepth":10,"TailSizeKB":10485760}`,
			},
		},
		"processors": []interface{}{
			map[string]interface{}{
				"type":        "processor_parse_json_native",
				"config_json": `{"SourceKey":"content"}`,
			},
		},
		"flushers": []interface{}{
			map[string]interface{}{
				"type":        "flusher_sls",
				"config_json": `{"Endpoint":"cn-shanghai-b-intranet.log.aliyuncs.com","Logstore":"kangaroo-pai-fabricmanager-proxy","Region":"cn-shanghai-b","TelemetryType":"logs"}`,
			},
		},
		"aggregators": []interface{}{},
		"global_json": `{"TopicType":"default"}`,
		"task_json":   `{}`,
		"log_sample":  "",
	})

	cfg, err := expandSlsLogtailPipelineConfig(d)
	if err != nil {
		t.Fatalf("expand failed: %v", err)
	}

	if cfg.Name != "kangaroo-pai-file-fabricmanager-proxy" {
		t.Fatalf("unexpected name: %s", cfg.Name)
	}
	if len(cfg.Inputs) != 1 || cfg.Inputs[0].Type != "input_file" {
		t.Fatalf("unexpected inputs: %+v", cfg.Inputs)
	}
	if len(cfg.Processors) != 1 || cfg.Processors[0].Type != "processor_parse_json_native" {
		t.Fatalf("unexpected processors: %+v", cfg.Processors)
	}
	if len(cfg.Flushers) != 1 || cfg.Flushers[0].Type != "flusher_sls" {
		t.Fatalf("unexpected flushers: %+v", cfg.Flushers)
	}
	if len(cfg.Aggregators) != 0 {
		t.Fatalf("aggregators should be empty")
	}
}

func TestFlattenSlsLogtailPipelineConfig_ProductionSample(t *testing.T) {
	r := resourceAliCloudLogtailConfig()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]interface{}{})

	libCfg := &sls.LogtailPipelineConfig{
		ConfigName: "kangaroo-pai-file-fabricmanager-proxy",
		Inputs: []map[string]interface{}{{
			"Type":                           "input_file",
			"AllowingIncludedByMultiConfigs": true,
		}},
		Processors: []map[string]interface{}{{
			"Type":      "processor_parse_json_native",
			"SourceKey": "content",
		}},
		Flushers: []map[string]interface{}{{
			"Type":          "flusher_sls",
			"Endpoint":      "cn-shanghai-b-intranet.log.aliyuncs.com",
			"Logstore":      "kangaroo-pai-fabricmanager-proxy",
			"Region":        "cn-shanghai-b",
			"TelemetryType": "logs",
		}},
		Aggregators:    []map[string]interface{}{},
		Global:         map[string]interface{}{"TopicType": "default"},
		Task:           map[string]interface{}{},
		LogSample:      "",
		CreateTime:     1770809720,
		LastModifyTime: 1770869527,
	}

	domain := FromLibConfig(libCfg, "test-project")
	if err := flattenSlsLogtailPipelineConfig(d, domain); err != nil {
		t.Fatalf("flatten failed: %v", err)
	}

	if got := d.Get("name").(string); got != "kangaroo-pai-file-fabricmanager-proxy" {
		t.Fatalf("unexpected state name: %s", got)
	}
	if got := d.Get("global_json").(string); got != `{"TopicType":"default"}` {
		t.Fatalf("unexpected state global_json: %s", got)
	}
	if got := d.Get("task_json").(string); got != `{}` {
		t.Fatalf("unexpected state task_json: %s", got)
	}

	inputs := d.Get("inputs").([]interface{})
	if len(inputs) != 1 {
		t.Fatalf("unexpected inputs size: %d", len(inputs))
	}
	in0 := inputs[0].(map[string]interface{})
	if in0["type"].(string) != "input_file" {
		t.Fatalf("unexpected input type: %v", in0["type"])
	}
}
