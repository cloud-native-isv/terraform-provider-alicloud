package alicloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccAliCloudLogtailConfig_Coexistence_SameObject(t *testing.T) {
	// P2: Verify that both resources can manage the same underlying object
	// and "Last Successful Write Wins" strategy applies.

	projectName := "tf-test-coexist-" + RandString(8)
	logstoreName := "tf-test-logstore-" + RandString(8)
	configName := "tf-test-config-" + RandString(8)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAliCloudLogtailConfigDestroy, // Checks both types if implementation allows
		Steps: []resource.TestStep{
			// 1. Create Legacy Resource
			{
				Config: testAccAliCloudLogtailConfigBasicConfig(projectName, logstoreName, configName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAliCloudLogtailConfigExists("alicloud_logtail_config.foo", nil),
					resource.TestCheckResourceAttr("alicloud_logtail_config.foo", "name", configName),
					resource.TestCheckResourceAttr("alicloud_logtail_config.foo", "input_type", "file"),
				),
			},
			// 2. Create Pipeline Resource pointing to SAME object (Adoption)
			// This tests the "Import on Create" logic we added.
			{
				Config: testAccAliCloudLogtailPipelineConfig_Coexistence_Adopt(projectName, logstoreName, configName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAliCloudLogtailPipelineConfigExists("alicloud_logtail_pipeline_config.bar", nil),
					resource.TestCheckResourceAttr("alicloud_logtail_pipeline_config.bar", "name", configName),
					// Should match the pipeline config definition
					resource.TestCheckResourceAttr("alicloud_logtail_pipeline_config.bar", "inputs.0.type", "input_file"),
				),
			},
			// 3. Update Legacy Resource (Overwrite Pipeline)
			{
				Config: testAccAliCloudLogtailConfigBasicConfig_Update(projectName, logstoreName, configName, "updated.log"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAliCloudLogtailConfigExists("alicloud_logtail_config.foo", nil),
					resource.TestCheckResourceAttr("alicloud_logtail_config.foo", "input_detail", `{"discardUnmatch":false,"enableRawLog":true,"filePattern":"updated.log","logPath":"/log","logType":"common_reg_log","maxDepth":1000,"topicFormat":"none"}`),
				),
			},
		},
	})
}

func testAccAliCloudLogtailPipelineConfig_Coexistence_Adopt(project, logstore, name string) string {
	return fmt.Sprintf(`
resource "alicloud_log_project" "foo" {
  name        = "%s"
  description = "tf-test"
}

resource "alicloud_log_store" "foo" {
  project          = alicloud_log_project.foo.name
  name             = "%s"
  retention_period = 3000
  shard_count      = 1
}

# Pipeline Config taking over (Same Name)
resource "alicloud_logtail_pipeline_config" "bar" {
  project = alicloud_log_project.foo.name
  name    = "%s" 
  
  inputs {
    type = "input_file"
    config_json = "{\"FilePaths\":[\"/log/access.log\"]}"
  }
  
  flushers {
    type = "flusher_sls"
    config_json = "{\"Logstore\":\"%s\", \"Region\":\"cn-hangzhou\"}"
  }
}
`, project, logstore, name, logstore)
}

func testAccAliCloudLogtailConfigBasicConfig_Update(project, logstore, name, pattern string) string {
	return fmt.Sprintf(`
resource "alicloud_log_project" "foo" {
  name        = "%s"
  description = "tf-test"
}

resource "alicloud_log_store" "foo" {
  project          = alicloud_log_project.foo.name
  name             = "%s"
  retention_period = 3000
  shard_count      = 1
}

resource "alicloud_logtail_config" "foo" {
  project      = alicloud_log_project.foo.name
  logstore     = alicloud_log_store.foo.name
  name         = "%s"
  input_type   = "file"
  output_type  = "LogService"
  input_detail = <<EOF
  {
	"logPath": "/log",
	"filePattern": "%s",
	"logType": "common_reg_log",
	"topicFormat": "none",
	"discardUnmatch": false,
	"enableRawLog": true,
	"maxDepth": 1000
  }
  EOF
}
`, project, logstore, name, pattern)
}
