package alicloud

import (
	"fmt"
	"strings"

	aliyunSlsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/sls"
)

// InternalOperationLogStoreIndex defines the default index configuration for internal-operation_log logstore
var InternalOperationLogStoreIndex = &aliyunSlsAPI.LogStoreIndex{
	TTL: 30, // Default 30 days retention for index
	Line: &aliyunSlsAPI.IndexLine{
		Token:         []string{" ", "\t", "\n", "\r", ",", ";", "=", "(", ")", "[", "]", "{", "}", "?", "@", "&", "<", ">", "/", ":", "\\", "'", "\""},
		CaseSensitive: false,
		IncludeKeys:   []string{},
		ExcludeKeys:   []string{},
	},
	Keys: map[string]*aliyunSlsAPI.IndexKey{
		"APIVersion": {
			Type:     "text",
			Token:    []string{" "},
			DocValue: true,
		},
		"AccessKeyId": {
			Type:  "text",
			Token: []string{" "},
		},
		"BeginTime": {
			Type: "long",
		},
		"CallerType": {
			Type:     "text",
			Token:    []string{" "},
			DocValue: true,
		},
		"ClientIP": {
			Type:  "text",
			Token: []string{" "},
		},
		"Config": {
			Type:     "text",
			Token:    []string{" "},
			DocValue: true,
		},
		"ConsumerGroup": {
			Type:     "text",
			Token:    []string{" "},
			DocValue: true,
		},
		"Dashboard": {
			Type:     "text",
			Token:    []string{" "},
			DocValue: true,
		},
		"DataStatus": {
			Type:     "text",
			Token:    []string{" "},
			DocValue: true,
		},
		"EndTime": {
			Type: "long",
		},
		"ErrorCode": {
			Type:     "text",
			Token:    []string{",", " ", "'", "\"", ";", "=", "(", ")", "[", "]", "{", "}", "?", "@", "&", "<", ">", "/", ":", "\n", "\t", "\r"},
			DocValue: true,
		},
		"ErrorMsg": {
			Type:     "text",
			Token:    []string{",", " ", "'", "\"", ";", "=", "(", ")", "[", "]", "{", "}", "?", "@", "&", "<", ">", "/", ":", "\n", "\t", "\r"},
			DocValue: true,
		},
		"InFlow": {
			Type:     "long",
			DocValue: true,
		},
		"InputLines": {
			Type:     "long",
			DocValue: true,
		},
		"InvokerUid": {
			Type:     "text",
			Token:    []string{" "},
			DocValue: true,
		},
		"Job": {
			Type:     "text",
			Token:    []string{" "},
			DocValue: true,
		},
		"JobName": {
			Type:     "text",
			Token:    []string{" ", ",", ";", "=", "(", ")", "{", "}", "<", ">", "?", "#"},
			DocValue: true,
		},
		"JobType": {
			Type:     "text",
			Token:    []string{" ", ",", ";", "=", "(", ")", "{", "}", "<", ">", "?", "#"},
			DocValue: true,
		},
		"Latency": {
			Type:     "long",
			DocValue: true,
		},
		"LogStore": {
			Type:     "text",
			Token:    []string{" ", ",", ";", "=", "(", ")", "{", "}", "<", ">", "?", "#"},
			DocValue: true,
		},
		"MachineGroup": {
			Type:     "text",
			Token:    []string{" "},
			DocValue: true,
		},
		"Method": {
			Type:     "text",
			Token:    []string{" ", ",", ";", "=", "(", ")", "{", "}", "<", ">", "?", "#"},
			DocValue: true,
		},
		"NetInflow": {
			Type:     "long",
			DocValue: true,
		},
		"NetOutFlow": {
			Type:     "long",
			DocValue: true,
		},
		"NetworkOut": {
			Type:     "long",
			DocValue: true,
		},
		"Project": {
			Type:     "text",
			Token:    []string{" ", ",", ";", "=", "(", ")", "{", "}", "<", ">", "?", "/", "#"},
			DocValue: true,
		},
		"Query": {
			Type:  "text",
			Token: []string{" ", ",", ";", "=", "(", ")", "{", "}", "<", ">", "?", "#"},
		},
		"RequestId": {
			Type:  "text",
			Token: []string{" "},
		},
		"RequestLines": {
			Type:     "long",
			DocValue: true,
		},
		"ResponseLines": {
			Type:     "long",
			DocValue: true,
		},
		"Reverse": {
			Type:     "long",
			DocValue: true,
		},
		"Role": {
			Type:     "text",
			Token:    []string{" "},
			DocValue: true,
		},
		"RoleSessionName": {
			Type:     "text",
			Token:    []string{" "},
			DocValue: true,
		},
		"SavedSearch": {
			Type:     "text",
			Token:    []string{" "},
			DocValue: true,
		},
		"Shard": {
			Type:     "long",
			DocValue: true,
		},
		"SourceIP": {
			Type:     "text",
			Token:    []string{" "},
			DocValue: true,
		},
		"Status": {
			Type:     "long",
			DocValue: true,
		},
		"TermUnit": {
			Type:     "long",
			DocValue: true,
		},
		"Topic": {
			Type:     "text",
			Token:    []string{" "},
			DocValue: true,
		},
		"UserAgent": {
			Type:     "text",
			Token:    []string{" "},
			DocValue: true,
		},
	},
}

// InternalDiagnosticLogStoreIndex defines the default index configuration for internal-diagnostic_log logstore
var InternalDiagnosticLogStoreIndex = &aliyunSlsAPI.LogStoreIndex{
	TTL: 30, // Default 30 days retention for index
	Line: &aliyunSlsAPI.IndexLine{
		Token:         []string{" ", "\t", "\n", "\r", ",", ";", "=", "(", ")", "[", "]", "{", "}", "?", "@", "&", "<", ">", "/", ":", "\\", "'", "\""},
		CaseSensitive: false,
		IncludeKeys:   []string{},
		ExcludeKeys:   []string{},
	},
	Keys: map[string]*aliyunSlsAPI.IndexKey{
		"_container_ip_": {
			Type:     "text",
			Token:    []string{" "},
			DocValue: true,
		},
		"_container_name_": {
			Type:     "text",
			Token:    []string{" "},
			DocValue: true,
		},
		"_etl_:connector_meta": {
			Type:  "json",
			Token: []string{",", " ", "'", "\"", ";", "=", "(", ")", "[", "]", "{", "}", "?", "@", "&", "<", ">", "/", ":", "\n", "\t", "\r"},
			JsonKeys: map[string]*aliyunSlsAPI.IndexKey{
				"action": {
					Type:     "text",
					Token:    []string{",", " ", "'", "\"", ";", "=", "(", ")", "[", "]", "{", "}", "?", "@", "&", "<", ">", "/", ":", "\n", "\t", "\r"},
					DocValue: true,
				},
				"connector": {
					Type:     "text",
					Token:    []string{",", " ", "'", "\"", ";", "=", "(", ")", "[", "]", "{", "}", "?", "@", "&", "<", ">", "/", ":", "\n", "\t", "\r"},
					DocValue: true,
				},
				"instance": {
					Type:     "text",
					Token:    []string{",", " ", "'", "\"", ";", "=", "(", ")", "[", "]", "{", "}", "?", "@", "&", "<", ">", "/", ":", "\n", "\t", "\r"},
					DocValue: true,
				},
				"task_id": {
					Type:     "long",
					DocValue: true,
				},
				"task_name": {
					Type:     "text",
					Token:    []string{",", " ", "'", "\"", ";", "=", "(", ")", "[", "]", "{", "}", "?", "@", "&", "<", ">", "/", ":", "\n", "\t", "\r"},
					DocValue: true,
				},
			},
		},
		"_etl_:connector_metrics": {
			Type:  "json",
			Token: []string{",", " ", "'", "\"", ";", "=", "(", ")", "[", "]", "{", "}", "?", "@", "&", "<", ">", "/", ":", "\n", "\t", "\r"},
			JsonKeys: map[string]*aliyunSlsAPI.IndexKey{
				"desc": {
					Type:     "text",
					Token:    []string{",", " ", "'", "\"", ";", "=", "(", ")", "[", "]", "{", "}", "?", "@", "&", "<", ">", "/", ":", "\n", "\t", "\r"},
					DocValue: true,
				},
				"error": {
					Type:     "text",
					Token:    []string{",", " ", "'", "\"", ";", "=", "(", ")", "[", "]", "{", "}", "?", "@", "&", "<", ">", "/", ":", "\n", "\t", "\r"},
					DocValue: true,
				},
				"events": {
					Type:     "long",
					DocValue: true,
				},
				"events_bytes": {
					Type:     "long",
					DocValue: true,
				},
				"extras": {
					Type:     "text",
					Token:    []string{",", " ", "'", "\"", ";", "=", "(", ")", "[", "]", "{", "}", "?", "@", "&", "<", ">", "/", ":", "\n", "\t", "\r"},
					DocValue: true,
				},
				"failed": {
					Type:     "long",
					DocValue: true,
				},
				"lags": {
					Type:     "double",
					DocValue: true,
				},
				"native_bytes": {
					Type:     "long",
					DocValue: true,
				},
				"pub_net_bytes": {
					Type:     "long",
					DocValue: true,
				},
				"rep_time": {
					Type:     "long",
					DocValue: true,
				},
				"req_count": {
					Type:     "long",
					DocValue: true,
				},
				"state": {
					Type:     "long",
					DocValue: true,
				},
			},
		},
		"_namespace_": {
			Type:     "text",
			Token:    []string{" "},
			DocValue: true,
		},
		"_pod_name_": {
			Type:     "text",
			Token:    []string{" "},
			DocValue: true,
		},
		"_pod_uid_": {
			Type:     "text",
			Token:    []string{" "},
			DocValue: true,
		},
		"alarm_count": {
			Type:     "long",
			DocValue: true,
		},
		"alarm_message": {
			Type:     "text",
			Token:    []string{",", " ", "'", "\"", ";", "=", "(", ")", "[", "]", "{", "}", "?", "@", "&", "<", ">", "/", ":", "\n", "\t", "\r"},
			DocValue: true,
		},
		"alarm_type": {
			Type:     "text",
			Token:    []string{",", " ", "'", "\"", ";"},
			DocValue: true,
		},
		"begin_time": {
			Type:     "long",
			DocValue: true,
		},
		"bytes_read": {
			Type:     "long",
			DocValue: true,
		},
		"config_name": {
			Type:     "text",
			Token:    []string{" "},
			DocValue: true,
		},
		"consumer_group": {
			Type:     "text",
			Token:    []string{",", " ", "'", "\"", ";", "=", "(", ")", "[", "]", "{", "}", "?", "@", "&", "<", ">", "/", ":", "\n", "\t", "\r"},
			DocValue: true,
		},
		"cpu": {
			Type:     "double",
			DocValue: true,
		},
		"create_time": {
			Type:     "long",
			DocValue: true,
		},
		"detail_metric": {
			Type:     "json",
			Token:    []string{",", " "},
			DocValue: true,
			JsonKeys: map[string]*aliyunSlsAPI.IndexKey{
				"config_count": {
					Type:     "long",
					Alias:    "config-count",
					DocValue: true,
				},
				"config_update_count": {
					Type:     "long",
					Alias:    "config-update-count",
					DocValue: true,
				},
				"event_tps": {
					Type:     "double",
					Alias:    "event-tps",
					DocValue: true,
				},
				"open_fd": {
					Type:     "long",
					Alias:    "open-fd",
					DocValue: true,
				},
				"reader_count": {
					Type:     "long",
					Alias:    "reader-count",
					DocValue: true,
				},
				"register_handler": {
					Type:     "long",
					Alias:    "register-handler",
					DocValue: true,
				},
				"send_bytes_ps": {
					Type:     "long",
					Alias:    "send-bytes-ps",
					DocValue: true,
				},
				"send_lines_ps": {
					Type:     "double",
					Alias:    "send-lines-ps",
					DocValue: true,
				},
				"send_net_bytes_ps": {
					Type:     "long",
					Alias:    "send-net-bytes-ps",
					DocValue: true,
				},
				"send_tps": {
					Type:     "double",
					Alias:    "send-tps",
					DocValue: true,
				},
				"start_time": {
					Type:     "text",
					Alias:    "start-time",
					Token:    []string{",", " ", "'", "\"", ";", "=", "(", ")", "[", "]", "{", "}", "?", "@", "&", "<", ">", "/", ":", "\n", "\t", "\r"},
					DocValue: true,
				},
			},
		},
		"eci_instance_id": {
			Type:     "text",
			Token:    []string{",", "", "'", "\"", ";", "\\", "$", "#", "!", "=", "(", ")", "[", "]", "{", "}", "?", "@", "&", "<", ">", "/", ":", "\n", "\t", "\r"},
			DocValue: true,
		},
		"ecs_instance_id": {
			Type:     "text",
			Token:    []string{",", " ", "'", "\"", ";", "\\", "$", "#", "!", "=", "(", ")", "[", "]", "{", "}", "?", "@", "&", "<", ">", "/", ":", "\n", "\t", "\r"},
			DocValue: true,
		},
		"end_time": {
			Type:     "long",
			DocValue: true,
		},
		"error_code": {
			Type:     "text",
			Token:    []string{" "},
			DocValue: true,
		},
		"error_line": {
			Type:     "text",
			Token:    []string{",", " ", "'", "\"", ";", "=", "(", ")", "[", "]", "{", "}", "?", "@", "&", "<", ">", "/", ":", "\n", "\t", "\r"},
			DocValue: true,
		},
		"error_message": {
			Type:     "text",
			Token:    []string{",", " ", "'", "\"", ";", "=", "(", ")", "[", "]", "{", "}", "?", "@", "&", "<", ">", "/", ":", "\n", "\t", "\r"},
			DocValue: true,
		},
		"fallbehind": {
			Type:     "long",
			DocValue: true,
		},
		"file_dev": {
			Type:     "long",
			DocValue: true,
		},
		"file_inode": {
			Type:     "long",
			DocValue: true,
		},
		"file_name": {
			Type:     "text",
			Token:    []string{",", " ", "'", "\"", ";", "=", "(", ")", "[", "]", "{", "}", "?", "@", "&", "<", ">", "/", ":", "\n", "\t", "\r"},
			DocValue: true,
		},
		"file_size": {
			Type:     "long",
			DocValue: true,
		},
		"history_data_failures": {
			Type:     "long",
			DocValue: true,
		},
		"hostname": {
			Type:     "text",
			Token:    []string{",", " "},
			DocValue: true,
		},
		"index_flow": {
			Type:     "long",
			DocValue: true,
		},
		"inflow": {
			Type:     "long",
			DocValue: true,
		},
		"instance_id": {
			Type:     "text",
			Token:    []string{" "},
			DocValue: true,
		},
		"ip": {
			Type:     "text",
			Token:    []string{",", " "},
			DocValue: true,
		},
		"job_name": {
			Type:     "text",
			Token:    []string{" "},
			DocValue: true,
		},
		"job_run_id": {
			Type:     "text",
			Token:    []string{" "},
			DocValue: true,
		},
		"job_type": {
			Type:     "text",
			Token:    []string{" "},
			DocValue: true,
		},
		"label.child_node_id": {
			Type:     "text",
			Token:    []string{",", " ", "'", "\"", ";", "\\", "$", "#", "!", "=", "(", ")", "[", "]", "{", "}", "?", "@", "&", "<", ">", "/", ":", "\n", "\t", "\r"},
			DocValue: true,
		},
		"label.child_plugin_id": {
			Type:     "text",
			Token:    []string{",", "", "'", "\"", ";", "\\", "$", "#", "!", "=", "(", ")", "[", "]", "{", "}", "?", "@", "&", "<", ">", "/", ":", "\n", "\t", "\r"},
			DocValue: true,
		},
		"label.config_name": {
			Type:     "text",
			Token:    []string{",", " ", "'", "\"", ";", "\\", "$", "#", "!", "=", "(", ")", "[", "]", "{", "}", "?", "@", "&", "<", ">", "/", ":", "\n", "\t", "\r"},
			DocValue: true,
		},
		"label.logstore": {
			Type:     "text",
			Token:    []string{",", "", "'", "\"", ";", "\\", "$", "#", "!", "=", "(", ")", "[", "]", "{", "}", "?", "@", "&", "<", ">", "/", ":", "\n", "\t", "\r"},
			DocValue: true,
		},
		"label.node_id": {
			Type:     "text",
			Token:    []string{",", " ", "'", "\"", ";", "\\", "$", "#", "!", "=", "(", ")", "[", "]", "{", "}", "?", "@", "&", "<", ">", "/", ":", "\n", "\t", "\r"},
			DocValue: true,
		},
		"label.plugin_id": {
			Type:     "text",
			Token:    []string{",", "", "'", "\"", ";", "\\", "$", "#", "!", "=", "(", ")", "[", "]", "{", "}", "?", "@", "&", "<", ">", "/", ":", "\n", "\t", "\r"},
			DocValue: true,
		},
		"label.plugin_name": {
			Type:     "text",
			Token:    []string{",", " ", "'", "\"", ";", "\\", "$", "#", "!", "=", "(", ")", "[", "]", "{", "}", "?", "@", "&", "<", ">", "/", ":", "\n", "\t", "\r"},
			DocValue: true,
		},
		"label.project": {
			Type:     "text",
			Token:    []string{",", "", "'", "\"", ";", "\\", "$", "#", "!", "=", "(", ")", "[", "]", "{", "}", "?", "@", "&", "<", ">", "/", ":", "\n", "\t", "\r"},
			DocValue: true,
		},
		"last_read_time": {
			Type:     "long",
			DocValue: true,
		},
		"load": {
			Type:     "double",
			DocValue: true,
		},
		"logstore": {
			Type:     "text",
			Token:    []string{" "},
			DocValue: true,
		},
		"logtail_version": {
			Type:     "text",
			Token:    []string{",", " "},
			DocValue: true,
		},
		"max_send_success_time": {
			Type:     "long",
			DocValue: true,
		},
		"max_unsend_time": {
			Type:     "long",
			DocValue: true,
		},
		"memory": {
			Type:     "long",
			DocValue: true,
		},
		"metric_type": {
			Type:     "text",
			Token:    []string{" "},
			DocValue: true,
		},
		"min_unsend_time": {
			Type:     "long",
			DocValue: true,
		},
		"network_out": {
			Type:     "long",
			DocValue: true,
		},
		"os": {
			Type:     "text",
			Token:    []string{",", " ", "'", "\"", ";", "=", "(", ")", "[", "]", "{", "}", "?", "@", "&", "<", ">", "/", ":", "\n", "\t", "\r"},
			DocValue: true,
		},
		"os_cpu": {
			Type:     "double",
			DocValue: true,
		},
		"outflow": {
			Type:     "long",
			DocValue: true,
		},
		"parse_failed": {
			Type:     "long",
			DocValue: true,
		},
		"parse_failures": {
			Type:     "long",
			DocValue: true,
		},
		"plugins": {
			Type:     "text",
			Token:    []string{" "},
			DocValue: true,
		},
		"project": {
			Type:     "text",
			Token:    []string{" "},
			DocValue: true,
		},
		"public_net_flow": {
			Type:     "long",
			DocValue: true,
		},
		"read_avg_delay": {
			Type:     "long",
			DocValue: true,
		},
		"read_count": {
			Type:     "long",
			DocValue: true,
		},
		"read_offset": {
			Type:     "long",
			DocValue: true,
		},
		"regex_match_failures": {
			Type:     "long",
			DocValue: true,
		},
		"rows_copied": {
			Type:     "long",
			DocValue: true,
		},
		"rows_read": {
			Type:     "long",
			DocValue: true,
		},
		"rows_skipped": {
			Type:     "long",
			DocValue: true,
		},
		"schedule_id": {
			Type:     "text",
			Token:    []string{" "},
			DocValue: true,
		},
		"schedule_time": {
			Type:     "long",
			DocValue: true,
		},
		"send_block_flag": {
			Type:     "text",
			Token:    []string{",", " ", "'", "\"", ";", "=", "(", ")", "[", "]", "{", "}", "?", "@", "&", "<", ">", "/", ":", "\n", "\t", "\r"},
			DocValue: true,
		},
		"send_discard_error": {
			Type:     "long",
			DocValue: true,
		},
		"send_failures": {
			Type:     "long",
			DocValue: true,
		},
		"send_network_error": {
			Type:     "long",
			DocValue: true,
		},
		"send_queue_size": {
			Type:     "long",
			DocValue: true,
		},
		"send_quota_error": {
			Type:     "long",
			DocValue: true,
		},
		"send_success_count": {
			Type:     "long",
			DocValue: true,
		},
		"sender_valid_flag": {
			Type:     "long",
			DocValue: true,
		},
		"shard": {
			Type:     "double",
			DocValue: true,
		},
		"source_ip": {
			Type:     "text",
			Token:    []string{","},
			DocValue: true,
		},
		"source_type": {
			Type:     "text",
			Token:    []string{",", " ", "'", "\"", ";", "=", "(", ")", "[", "]", "{", "}", "?", "@", "&", "<", ">", "/", ":", "\n", "\t", "\r"},
			DocValue: true,
		},
		"start_time": {
			Type:     "long",
			DocValue: true,
		},
		"status": {
			Type:     "text",
			Token:    []string{",", " ", "'", "\"", ";", "=", "(", ")", "[", "]", "{", "}", "?", "@", "&", "<", ">", "/", ":", "\n", "\t", "\r"},
			DocValue: true,
		},
		"storage_index": {
			Type:     "long",
			DocValue: true,
		},
		"storage_raw": {
			Type:     "long",
			DocValue: true,
		},
		"succeed_lines": {
			Type:     "long",
			DocValue: true,
		},
		"task_count": {
			Type:     "long",
			DocValue: true,
		},
		"time_format_failures": {
			Type:     "long",
			DocValue: true,
		},
		"total_bytes": {
			Type:     "long",
			DocValue: true,
		},
		"user_defined_id": {
			Type:     "text",
			Token:    []string{" "},
			DocValue: true,
		},
		"uuid": {
			Type:     "text",
			Token:    []string{","},
			DocValue: true,
		},
		"value.proc_discard_records_total": {
			Type:     "long",
			DocValue: true,
		},
		"value.proc_in_records_total": {
			Type:     "long",
			DocValue: true,
		},
		"value.proc_key_count_not_match_error_total": {
			Type:     "long",
			DocValue: true,
		},
		"value.proc_out_records_total": {
			Type:     "long",
			DocValue: true,
		},
		"value.proc_parse_error_total": {
			Type:     "long",
			DocValue: true,
		},
		"value.proc_time_ms": {
			Type:     "long",
			DocValue: true,
		},
		"version": {
			Type:     "text",
			Token:    []string{",", " "},
			DocValue: true,
		},
		"worker_count": {
			Type:     "long",
			DocValue: true,
		},
		"write_count": {
			Type:     "long",
			DocValue: true,
		},
	},
}

func (s *SlsService) CreateProjectLogging(projectName string, logging *aliyunSlsAPI.LogProjectLogging) error {
	addDebug("SlsService.CreateProjectLogging", fmt.Sprintf("Starting create project logging for project: %s", projectName), nil)

	if s.aliyunSlsAPI == nil {
		addDebug("SlsService.CreateProjectLogging", "aliyunSlsAPI client is not initialized", nil)
		return fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	// Ensure logging project exists
	if logging != nil && logging.LoggingProject != "" {
		addDebug("SlsService.CreateProjectLogging", fmt.Sprintf("Checking/creating logging project: %s", logging.LoggingProject), nil)

		// Create project with region-aware configuration
		createProject := &aliyunSlsAPI.LogProject{
			ProjectName:        logging.LoggingProject,
			Description:        fmt.Sprintf("service logging for %s", projectName),
			DataRedundancyType: aliyunSlsAPI.DataRedundancyTypeLRS,
			RecycleBinEnabled:  false,
		}

		if _, err := s.CreateProjectIfNotExist(createProject); err != nil {
			addDebug("SlsService.CreateProjectLogging", fmt.Sprintf("Failed to create logging project: %s", logging.LoggingProject), err)
			return WrapErrorf(err, DefaultErrorMsg, logging.LoggingProject, "CreateProjectIfNotExist", AlibabaCloudSdkGoERROR)
		}
		addDebug("SlsService.CreateProjectLogging", fmt.Sprintf("Successfully ensured logging project exists: %s", logging.LoggingProject), nil)

		// Ensure all logstores in logging details exist
		if logging.LoggingDetails != nil {
			addDebug("SlsService.CreateProjectLogging", fmt.Sprintf("Processing %d logstore configurations", len(logging.LoggingDetails)), nil)
			for _, detail := range logging.LoggingDetails {
				if detail.Logstore != "" {
					addDebug("SlsService.CreateProjectLogging", fmt.Sprintf("Checking/creating logstore: %s in project: %s", detail.Logstore, logging.LoggingProject), nil)

					// Create logstore with default configuration
					createLogStore := &aliyunSlsAPI.LogStore{
						LogstoreName:  detail.Logstore,
						Ttl:           30,   // Default 30 days retention
						ShardCount:    8,    // Default 8 shards
						AutoSplit:     true, // Enable auto split
						MaxSplitShard: 64,   // Default max split shard count
						AppendMeta:    true, // Enable append meta
					}
					createdLogStore, err := s.CreateLogStoreIfNotExist(logging.LoggingProject, createLogStore)
					if err != nil {
						addDebug("SlsService.CreateProjectLogging", fmt.Sprintf("Failed to create logstore: %s in project: %s", detail.Logstore, logging.LoggingProject), err)
						return WrapErrorf(err, DefaultErrorMsg, detail.Logstore, "CreateLogStoreIfNotExist", AlibabaCloudSdkGoERROR)
					}

					// Always check and create default index for internal logstores
					// regardless of whether the logstore was newly created or already existed
					switch detail.Logstore {
					case "internal-diagnostic_log":
						addDebug("SlsService.CreateProjectLogging", fmt.Sprintf("Checking/creating default index for internal-diagnostic_log logstore: %s", detail.Logstore), nil)
						defaultIndex := InternalDiagnosticLogStoreIndex

						// Print detailed structure of the default index configuration
						// addDebugStructure("InternalDiagnosticLogStoreIndex", defaultIndex)
						// addDebugStructure("InternalDiagnosticLogStoreIndex(in sdk)", aliyunSlsAPI.ConvertLogStoreIndexToSDKIndex(defaultIndex))

						// Create or update the index for the logstore
						if err := s.aliyunSlsAPI.CreateOrUpdateLogStoreIndex(logging.LoggingProject, detail.Logstore, defaultIndex); err != nil {
							// Log warning but don't fail the whole operation if index creation fails
							// This ensures the logging configuration is still created even if index fails
							// Users can manually create indexes later if needed
							addDebug("SlsService.CreateProjectLogging", fmt.Sprintf("Warning: Failed to create or update index for logstore %s: %v", detail.Logstore, err), err)
							fmt.Printf("Warning: Failed to create or update index for logstore %s: %v\n", detail.Logstore, err)
						} else {
							addDebug("SlsService.CreateProjectLogging", fmt.Sprintf("Successfully created or updated index for logstore: %s", detail.Logstore), nil)
						}
					case "internal-operation_log":
						addDebug("SlsService.CreateProjectLogging", fmt.Sprintf("Checking/creating default index for internal-operation_log logstore: %s", detail.Logstore), nil)
						defaultIndex := InternalOperationLogStoreIndex

						// Print detailed structure of the default index configuration
						// addDebugStructure("InternalOperationLogStoreIndex", defaultIndex)
						// addDebugStructure("InternalOperationLogStoreIndex(in sdk)", aliyunSlsAPI.ConvertLogStoreIndexToSDKIndex(defaultIndex))

						// Create or update the index for the logstore
						if err := s.aliyunSlsAPI.CreateOrUpdateLogStoreIndex(logging.LoggingProject, detail.Logstore, defaultIndex); err != nil {
							// Log warning but don't fail the whole operation if index creation fails
							// This ensures the logging configuration is still created even if index fails
							// Users can manually create indexes later if needed
							addDebug("SlsService.CreateProjectLogging", fmt.Sprintf("Warning: Failed to create or update index for logstore %s: %v", detail.Logstore, err), err)
							fmt.Printf("Warning: Failed to create or update index for logstore %s: %v\n", detail.Logstore, err)
						} else {
							addDebug("SlsService.CreateProjectLogging", fmt.Sprintf("Successfully created or updated index for logstore: %s", detail.Logstore), nil)
						}
					}

					// Log the result based on whether logstore was newly created
					if createdLogStore != nil {
						addDebug("SlsService.CreateProjectLogging", fmt.Sprintf("Logstore %s was newly created", detail.Logstore), nil)
					} else {
						addDebug("SlsService.CreateProjectLogging", fmt.Sprintf("Logstore %s already existed", detail.Logstore), nil)
					}
				}
			}
			addDebug("SlsService.CreateProjectLogging", "Completed processing all logstore configurations", nil)
		}
	}

	addDebug("SlsService.CreateProjectLogging", fmt.Sprintf("Calling API to create project logging for: %s", projectName), logging)
	err := s.aliyunSlsAPI.CreateLogProjectLogging(projectName, logging)
	if err != nil {
		addDebug("SlsService.CreateProjectLogging", fmt.Sprintf("Failed to create project logging configuration for: %s", projectName), err)
		return WrapErrorf(err, DefaultErrorMsg, projectName, "CreateLogging", AlibabaCloudSdkGoERROR)
	}

	addDebug("SlsService.CreateProjectLogging", fmt.Sprintf("Successfully created project logging for: %s", projectName), nil)
	return nil
}

func (s *SlsService) UpdateProjectLogging(projectName string, logging *aliyunSlsAPI.LogProjectLogging) error {
	addDebug("SlsService.UpdateProjectLogging", fmt.Sprintf("Starting update project logging for project: %s", projectName), nil)

	if s.aliyunSlsAPI == nil {
		addDebug("SlsService.UpdateProjectLogging", "aliyunSlsAPI client is not initialized", nil)
		return fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	// Ensure logging project exists
	if logging != nil && logging.LoggingProject != "" {
		addDebug("SlsService.UpdateProjectLogging", fmt.Sprintf("Checking/creating logging project: %s", logging.LoggingProject), nil)

		// Create project with region-aware configuration
		createProject := &aliyunSlsAPI.LogProject{
			ProjectName:        logging.LoggingProject,
			Description:        fmt.Sprintf("service logging for %s", projectName),
			DataRedundancyType: aliyunSlsAPI.DataRedundancyTypeLRS,
			RecycleBinEnabled:  false,
		}

		if _, err := s.CreateProjectIfNotExist(createProject); err != nil {
			addDebug("SlsService.UpdateProjectLogging", fmt.Sprintf("Failed to create logging project: %s", logging.LoggingProject), err)
			return WrapErrorf(err, DefaultErrorMsg, logging.LoggingProject, "CreateProjectIfNotExist", AlibabaCloudSdkGoERROR)
		}
		addDebug("SlsService.UpdateProjectLogging", fmt.Sprintf("Successfully ensured logging project exists: %s", logging.LoggingProject), nil)

		// Ensure all logstores in logging details exist
		if logging.LoggingDetails != nil {
			addDebug("SlsService.UpdateProjectLogging", fmt.Sprintf("Processing %d logstore configurations", len(logging.LoggingDetails)), nil)
			for _, detail := range logging.LoggingDetails {
				if detail.Logstore != "" {
					addDebug("SlsService.UpdateProjectLogging", fmt.Sprintf("Checking/creating logstore: %s in project: %s", detail.Logstore, logging.LoggingProject), nil)
					// Create logstore with default configuration
					createLogStore := &aliyunSlsAPI.LogStore{
						LogstoreName:  detail.Logstore,
						Ttl:           30,   // Default 30 days retention
						ShardCount:    8,    // Default 8 shards
						AutoSplit:     true, // Enable auto split
						MaxSplitShard: 64,   // Default max split shard count
						AppendMeta:    true, // Enable append meta
					}
					createdLogStore, err := s.CreateLogStoreIfNotExist(logging.LoggingProject, createLogStore)
					if err != nil {
						addDebug("SlsService.UpdateProjectLogging", fmt.Sprintf("Failed to create logstore: %s in project: %s", detail.Logstore, logging.LoggingProject), err)
						return WrapErrorf(err, DefaultErrorMsg, detail.Logstore, "CreateLogStoreIfNotExist", AlibabaCloudSdkGoERROR)
					}

					// Always check and create default index for internal logstores
					// regardless of whether the logstore was newly created or already existed
					switch detail.Logstore {
					case "internal-diagnostic_log":
						addDebug("SlsService.UpdateProjectLogging", fmt.Sprintf("Checking/creating default index for internal-diagnostic_log logstore: %s", detail.Logstore), nil)
						defaultIndex := InternalDiagnosticLogStoreIndex

						// Create or update the index for the logstore
						if err := s.aliyunSlsAPI.CreateOrUpdateLogStoreIndex(logging.LoggingProject, detail.Logstore, defaultIndex); err != nil {
							// Log warning but don't fail the whole operation if index creation fails
							// This ensures the logging configuration is still created even if index fails
							// Users can manually create indexes later if needed
							addDebug("SlsService.UpdateProjectLogging", fmt.Sprintf("Warning: Failed to create or update index for logstore %s: %v", detail.Logstore, err), err)
							fmt.Printf("Warning: Failed to create or update index for logstore %s: %v\n", detail.Logstore, err)
						} else {
							addDebug("SlsService.UpdateProjectLogging", fmt.Sprintf("Successfully created or updated index for logstore: %s", detail.Logstore), nil)
						}
					case "internal-operation_log":
						addDebug("SlsService.UpdateProjectLogging", fmt.Sprintf("Checking/creating default index for internal-operation_log logstore: %s", detail.Logstore), nil)
						defaultIndex := InternalOperationLogStoreIndex

						// Create or update the index for the logstore
						if err := s.aliyunSlsAPI.CreateOrUpdateLogStoreIndex(logging.LoggingProject, detail.Logstore, defaultIndex); err != nil {
							// Log warning but don't fail the whole operation if index creation fails
							// This ensures the logging configuration is still created even if index fails
							// Users can manually create indexes later if needed
							addDebug("SlsService.UpdateProjectLogging", fmt.Sprintf("Warning: Failed to create or update index for logstore %s: %v", detail.Logstore, err), err)
							fmt.Printf("Warning: Failed to create or update index for logstore %s: %v\n", detail.Logstore, err)
						} else {
							addDebug("SlsService.UpdateProjectLogging", fmt.Sprintf("Successfully created or updated index for logstore: %s", detail.Logstore), nil)
						}
					}

					// Log the result based on whether logstore was newly created
					if createdLogStore != nil {
						addDebug("SlsService.UpdateProjectLogging", fmt.Sprintf("Logstore %s was newly created", detail.Logstore), nil)
					} else {
						addDebug("SlsService.UpdateProjectLogging", fmt.Sprintf("Logstore %s already existed", detail.Logstore), nil)
					}
					addDebug("SlsService.UpdateProjectLogging", fmt.Sprintf("Successfully ensured logstore exists: %s", detail.Logstore), nil)
				}
			}
			addDebug("SlsService.UpdateProjectLogging", "Completed processing all logstore configurations", nil)
		}
	}

	addDebug("SlsService.UpdateProjectLogging", fmt.Sprintf("Calling API to update project logging for: %s", projectName), logging)
	err := s.aliyunSlsAPI.UpdateLogProjectLogging(projectName, logging)
	if err != nil {
		addDebug("SlsService.UpdateProjectLogging", fmt.Sprintf("API call failed for project: %s", projectName), err)
		return WrapErrorf(err, DefaultErrorMsg, projectName, "UpdateLogging", AlibabaCloudSdkGoERROR)
	}

	addDebug("SlsService.UpdateProjectLogging", fmt.Sprintf("Successfully updated project logging for: %s", projectName), nil)
	return nil
}

func (s *SlsService) DeleteProjectLogging(projectName string) error {
	addDebug("SlsService.DeleteProjectLogging", fmt.Sprintf("Starting delete project logging for project: %s", projectName), nil)

	if s.aliyunSlsAPI == nil {
		addDebug("SlsService.DeleteProjectLogging", "aliyunSlsAPI client is not initialized", nil)
		return fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	addDebug("SlsService.DeleteProjectLogging", fmt.Sprintf("Calling API to delete project logging for: %s", projectName), nil)
	err := s.aliyunSlsAPI.DeleteLogProjectLogging(projectName)
	if err != nil {
		addDebug("SlsService.DeleteProjectLogging", fmt.Sprintf("API call failed for project: %s", projectName), err)
		return WrapErrorf(err, DefaultErrorMsg, projectName, "DeleteLogging", AlibabaCloudSdkGoERROR)
	}

	addDebug("SlsService.DeleteProjectLogging", fmt.Sprintf("Successfully deleted project logging for: %s", projectName), nil)
	return nil
}

func (s *SlsService) GetProjectLogging(projectName string) (*aliyunSlsAPI.LogProjectLogging, error) {
	addDebug("SlsService.GetProjectLogging", fmt.Sprintf("Starting get project logging for project: %s", projectName), nil)

	if s.aliyunSlsAPI == nil {
		addDebug("SlsService.GetProjectLogging", "aliyunSlsAPI client is not initialized", nil)
		return nil, fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	addDebug("SlsService.GetProjectLogging", fmt.Sprintf("Calling API to get project logging for: %s", projectName), nil)
	logging, err := s.aliyunSlsAPI.GetLogProjectLogging(projectName)
	if err != nil {
		if strings.Contains(err.Error(), "LoggingNotExist") {
			addDebug("SlsService.GetProjectLogging", fmt.Sprintf("Project logging not found for: %s", projectName), err)
			return nil, WrapErrorf(NotFoundErr("LogProjectLogging", projectName), NotFoundMsg, "")
		}
		addDebug("SlsService.GetProjectLogging", fmt.Sprintf("API call failed for project: %s", projectName), err)
		return nil, WrapErrorf(err, DefaultErrorMsg, projectName, "GetLogging", AlibabaCloudSdkGoERROR)
	}

	addDebug("SlsService.GetProjectLogging", fmt.Sprintf("Successfully retrieved project logging for: %s", projectName), logging)
	return logging, nil
}
