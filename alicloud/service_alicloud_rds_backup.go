package alicloud

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/PaesslerAG/jsonpath"
	util "github.com/alibabacloud-go/tea-utils/service"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// RDS Backup related operations

func (s *RdsService) DescribeBackupPolicy(id string) (object map[string]interface{}, err error) {
	action := "DescribeBackupPolicy"
	request := map[string]interface{}{
		"RegionId":     s.client.RegionId,
		"DBInstanceId": id,
		"SourceIp":     s.client.SourceIp,
	}
	client := s.client
	response, err := client.RpcPost("Rds", "2014-08-15", action, nil, request, true)
	if err != nil {
		if IsExpectedErrors(err, []string{"InvalidDBInstanceId.NotFound"}) {
			return nil, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
		}
		return nil, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	addDebug(action, response, request)
	return response, nil
}

func (s *RdsService) DescribeRdsBackup(id string) (object map[string]interface{}, err error) {
	var response map[string]interface{}
	client := s.client
	action := "DescribeBackups"
	parts, err := ParseResourceId(id, 2)
	if err != nil {
		err = WrapError(err)
		return
	}
	request := map[string]interface{}{
		"SourceIp":     s.client.SourceIp,
		"BackupId":     parts[1],
		"DBInstanceId": parts[0],
	}
	wait := incrementalWait(3*time.Second, 3*time.Second)
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		response, err = client.RpcPost("Rds", "2014-08-15", action, nil, request, true)
		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	addDebug(action, response, request)
	if err != nil {
		return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	v, err := jsonpath.Get("$.Items.Backup", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.Items.Backup", response)
	}
	if len(v.([]interface{})) < 1 {
		return object, WrapErrorf(NotFoundErr("RDS", id), NotFoundWithResponse, response)
	}
	object = v.([]interface{})[0].(map[string]interface{})
	return object, nil
}

func (s *RdsService) DescribeBackupTasks(id string, backupJobId string) (object map[string]interface{}, err error) {
	var response map[string]interface{}
	client := s.client
	action := "DescribeBackupTasks"
	request := map[string]interface{}{
		"SourceIp":     s.client.SourceIp,
		"DBInstanceId": id,
		"BackupJobId":  backupJobId,
	}
	wait := incrementalWait(3*time.Second, 3*time.Second)
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		response, err = client.RpcPost("Rds", "2014-08-15", action, nil, request, true)
		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	addDebug(action, response, request)
	if err != nil {
		return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	v, err := jsonpath.Get("$.Items.BackupJob", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.Items.BackupJob", response)
	}
	if len(v.([]interface{})) < 1 {
		return object, WrapErrorf(NotFoundErr("RDS", id), NotFoundWithResponse, response)
	} else {
		if fmt.Sprint(v.([]interface{})[0].(map[string]interface{})["BackupJobId"]) != backupJobId {
			return object, WrapErrorf(NotFoundErr("RDS", id), NotFoundWithResponse, response)
		}
	}
	object = v.([]interface{})[0].(map[string]interface{})
	return object, nil
}

func (s *RdsService) DescribeInstanceCrossBackupPolicy(id string) (object map[string]interface{}, err error) {
	action := "DescribeInstanceCrossBackupPolicy"
	request := map[string]interface{}{
		"RegionId":     s.client.RegionId,
		"DBInstanceId": id,
		"SourceIp":     s.client.SourceIp,
	}
	client := s.client
	response, err := client.RpcPost("Rds", "2014-08-15", action, nil, request, true)
	if err != nil {
		if IsExpectedErrors(err, []string{"InvalidDBInstanceId.NotFound"}) {
			return nil, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
		}
		return nil, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	if v, ok := response["BackupEnabled"]; ok && v.(string) == "Disabled" {
		return nil, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
	}
	addDebug(action, response, request)
	return response, nil
}

func (s *RdsService) ModifyDBBackupPolicy(d *schema.ResourceData, updateForData, updateForLog bool) error {
	enableBackupLog := "1"
	backupPeriod := ""
	if v, ok := d.GetOk("preferred_backup_period"); ok && v.(*schema.Set).Len() > 0 {
		periodList := expandStringList(v.(*schema.Set).List())
		backupPeriod = fmt.Sprintf("%s", strings.Join(periodList[:], COMMA_SEPARATED))
	} else {
		periodList := expandStringList(d.Get("backup_period").(*schema.Set).List())
		backupPeriod = fmt.Sprintf("%s", strings.Join(periodList[:], COMMA_SEPARATED))
	}

	backupTime := "02:00Z-03:00Z"
	if v, ok := d.GetOk("preferred_backup_time"); ok && v.(string) != "02:00Z-03:00Z" {
		backupTime = v.(string)
	} else if v, ok := d.GetOk("backup_time"); ok && v.(string) != "" {
		backupTime = v.(string)
	}

	retentionPeriod := "7"
	if v, ok := d.GetOk("backup_retention_period"); ok && v.(int) != 7 {
		retentionPeriod = strconv.Itoa(v.(int))
	} else if v, ok := d.GetOk("retention_period"); ok && v.(int) != 0 {
		retentionPeriod = strconv.Itoa(v.(int))
	}

	logBackupRetentionPeriod := ""
	if v, ok := d.GetOk("log_backup_retention_period"); ok && v.(int) != 0 {
		logBackupRetentionPeriod = strconv.Itoa(v.(int))
	} else if v, ok := d.GetOk("log_retention_period"); ok && v.(int) != 0 {
		logBackupRetentionPeriod = strconv.Itoa(v.(int))
	}

	localLogRetentionHours := ""
	if v, ok := d.GetOk("local_log_retention_hours"); ok {
		localLogRetentionHours = strconv.Itoa(v.(int))
	}

	localLogRetentionSpace := ""
	if v, ok := d.GetOk("local_log_retention_space"); ok {
		localLogRetentionSpace = strconv.Itoa(v.(int))
	}

	highSpaceUsageProtection := d.Get("high_space_usage_protection").(string)

	if !d.Get("enable_backup_log").(bool) {
		enableBackupLog = "0"
	}

	if d.HasChange("log_backup_retention_period") {
		if d.Get("log_backup_retention_period").(int) > d.Get("backup_retention_period").(int) {
			logBackupRetentionPeriod = retentionPeriod
		}
	}

	logBackupFrequency := ""
	if v, ok := d.GetOk("log_backup_frequency"); ok {
		logBackupFrequency = v.(string)
	}

	compressType := ""
	if v, ok := d.GetOk("compress_type"); ok {
		compressType = v.(string)
	}

	archiveBackupRetentionPeriod := "0"
	if v, ok := d.GetOk("archive_backup_retention_period"); ok {
		archiveBackupRetentionPeriod = strconv.Itoa(v.(int))
	}

	archiveBackupKeepCount := "1"
	if v, ok := d.GetOk("archive_backup_keep_count"); ok {
		archiveBackupKeepCount = strconv.Itoa(v.(int))
	}

	archiveBackupKeepPolicy := "0"
	if v, ok := d.GetOk("archive_backup_keep_policy"); ok {
		archiveBackupKeepPolicy = v.(string)
	}

	releasedKeepPolicy := ""
	if v, ok := d.GetOk("released_keep_policy"); ok {
		releasedKeepPolicy = v.(string)
	}

	category := ""
	if v, ok := d.GetOk("category"); ok {
		category = v.(string)
	}
	backupInterval := "-1"
	if v, ok := d.GetOk("backup_interval"); ok {
		backupInterval = v.(string)
	}
	enableIncrementDataBackup := false
	if v, ok := d.GetOkExists("enable_increment_data_backup"); ok {
		enableIncrementDataBackup = v.(bool)
	}
	backupMethod := "Physical"
	if v, ok := d.GetOk("backup_method"); ok {
		backupMethod = v.(string)
	}
	logBackupLocalRetentionNumber := 60
	if v, ok := d.GetOk("log_backup_local_retention_number"); ok {
		logBackupLocalRetentionNumber = v.(int)
	}
	runtime := util.RuntimeOptions{}
	runtime.SetAutoretry(true)
	instance, err := s.DescribeDBInstance(d.Id())
	if err != nil {
		return WrapError(err)
	}
	if updateForData {
		client := s.client
		action := "ModifyBackupPolicy"
		request := map[string]interface{}{
			"RegionId":              s.client.RegionId,
			"DBInstanceId":          d.Id(),
			"PreferredBackupPeriod": backupPeriod,
			"PreferredBackupTime":   backupTime,
			"BackupRetentionPeriod": retentionPeriod,
			"CompressType":          compressType,
			"BackupPolicyMode":      "DataBackupPolicy",
			"SourceIp":              s.client.SourceIp,
			"ReleasedKeepPolicy":    releasedKeepPolicy,
			"Category":              category,
		}
		if instance["Engine"] == "SQLServer" && instance["Category"] == "AlwaysOn" {
			if v, ok := d.GetOk("backup_priority"); ok {
				request["BackupPriority"] = v.(int)
			}

		}
		if instance["Engine"] == "SQLServer" && instance["DBInstanceStorageType"] != "local_ssd" {
			request["EnableIncrementDataBackup"] = enableIncrementDataBackup
			request["BackupMethod"] = backupMethod
		}
		if instance["Engine"] == "SQLServer" && logBackupFrequency == "LogInterval" {
			request["LogBackupFrequency"] = logBackupFrequency
		}
		if instance["Engine"] == "MySQL" && instance["DBInstanceStorageType"] == "local_ssd" {
			request["ArchiveBackupRetentionPeriod"] = archiveBackupRetentionPeriod
			request["ArchiveBackupKeepCount"] = archiveBackupKeepCount
			request["ArchiveBackupKeepPolicy"] = archiveBackupKeepPolicy
		}
		if (instance["Engine"] == "MySQL" || instance["Engine"] == "PostgreSQL") && instance["DBInstanceStorageType"] != "local_ssd" {
			request["BackupInterval"] = backupInterval
		}

		response, err := client.RpcPost("Rds", "2014-08-15", action, nil, request, false)
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), action, AlibabaCloudSdkGoERROR)
		}
		addDebug(action, response, request)
		if err := s.WaitForDBInstance(d.Id(), Running, DefaultTimeoutMedium); err != nil {
			return WrapError(err)
		}
	}

	// At present, the sql server database does not support setting logBackupRetentionPeriod
	if updateForLog && instance["Engine"] != "SQLServer" {
		client := s.client
		action := "ModifyBackupPolicy"
		request := map[string]interface{}{
			"RegionId":                      s.client.RegionId,
			"DBInstanceId":                  d.Id(),
			"EnableBackupLog":               enableBackupLog,
			"LocalLogRetentionHours":        localLogRetentionHours,
			"LocalLogRetentionSpace":        localLogRetentionSpace,
			"HighSpaceUsageProtection":      highSpaceUsageProtection,
			"BackupPolicyMode":              "LogBackupPolicy",
			"LogBackupRetentionPeriod":      logBackupRetentionPeriod,
			"LogBackupLocalRetentionNumber": logBackupLocalRetentionNumber,
			"SourceIp":                      s.client.SourceIp,
		}
		response, err := client.RpcPost("Rds", "2014-08-15", action, nil, request, false)
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), action, AlibabaCloudSdkGoERROR)
		}
		addDebug(action, response, request)
		if err := s.WaitForDBInstance(d.Id(), Running, DefaultTimeoutMedium); err != nil {
			return WrapError(err)
		}
	}
	return nil
}

func (s *RdsService) RdsBackupStateRefreshFunc(id string, backupJobId string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeBackupTasks(id, backupJobId)
		if err != nil {
			if IsNotFoundError(err) {
				// Set this to nil as if we didn't find anything.
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		for _, failState := range failStates {
			if object["BackupStatus"] == failState {
				return object, object["BackupStatus"].(string), WrapError(Error(FailedToReachTargetStatus, object["BackupStatus"]))
			}
		}
		return object, object["BackupStatus"].(string), nil
	}
}
