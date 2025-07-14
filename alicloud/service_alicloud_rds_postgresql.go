package alicloud

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// RDS PostgreSQL-specific operations

func (s *RdsService) DescribePGHbaConfig(id string) (object map[string]interface{}, err error) {
	var response map[string]interface{}
	client := s.client
	action := "DescribePGHbaConfig"
	request := map[string]interface{}{
		"RegionId":     s.client.RegionId,
		"SourceIp":     s.client.SourceIp,
		"DBInstanceId": id,
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
		if IsExpectedErrors(err, []string{"InvalidDBInstanceId.NotFound"}) {
			return object, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
		}
		return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	return response, nil
}

func (s *RdsService) RefreshPgHbaConf(d *schema.ResourceData, attribute string) error {
	response, err := s.DescribePGHbaConfig(d.Id())
	if err != nil {
		return WrapError(err)
	}
	runningHbaItems := make([]interface{}, 0)
	if v, exist := response["RunningHbaItems"].(map[string]interface{})["HbaItem"]; exist {
		runningHbaItems = v.([]interface{})
	}

	var items []map[string]interface{}

	documented, ok := d.GetOk(attribute)
	if !ok {
		return nil
	}

	for _, item := range documented.(*schema.Set).List() {
		item := item.(map[string]interface{})
		for _, item2 := range runningHbaItems {
			item2 := item2.(map[string]interface{})
			if item["priority_id"] == formatInt(item2["PriorityId"]) {
				mapping := map[string]interface{}{
					"type":        item2["Type"],
					"database":    item2["Database"],
					"priority_id": formatInt(item2["PriorityId"]),
					"address":     item2["Address"],
					"user":        item2["User"],
					"method":      item2["Method"],
					"option":      item2["Option"],
					"mask":        item2["Mask"],
				}
				if item2["mask"] != nil && item2["mask"] != "" {
					mapping["mask"] = item2["mask"]
				}
				if item2["option"] != nil && item2["option"] != "" {
					mapping["option"] = item2["option"]
				}
				items = append(items, mapping)
			}
		}
	}
	if len(items) > 0 {
		if err := d.Set(attribute, items); err != nil {
			return WrapError(err)
		}
	}
	return nil
}

func (s *RdsService) ModifyPgHbaConfig(d *schema.ResourceData, attribute string) error {
	client := s.client
	var err error
	action := "ModifyPGHbaConfig"
	request := map[string]interface{}{
		"RegionId":     s.client.RegionId,
		"DBInstanceId": d.Id(),
		"SourceIp":     s.client.SourceIp,
	}
	request["OpsType"] = "update"
	pgHbaConfig := d.Get("pg_hba_conf")
	count := 1
	for _, i := range pgHbaConfig.(*schema.Set).List() {
		i := i.(map[string]interface{})
		request[fmt.Sprint("HbaItem.", count, ".Type")] = i["type"]
		if i["mask"] != nil && i["mask"] != "" {
			request[fmt.Sprint("HbaItem.", count, ".Mask")] = i["mask"]
		}
		request[fmt.Sprint("HbaItem.", count, ".Database")] = i["database"]
		request[fmt.Sprint("HbaItem.", count, ".PriorityId")] = i["priority_id"]
		request[fmt.Sprint("HbaItem.", count, ".Address")] = i["address"]
		request[fmt.Sprint("HbaItem.", count, ".User")] = i["user"]
		request[fmt.Sprint("HbaItem.", count, ".Method")] = i["method"]
		if i["option"] != nil && i["mask"] != "" {
			request[fmt.Sprint("HbaItem.", count, ".Option")] = i["option"]
		}
		count = count + 1
	}
	var response map[string]interface{}
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		response, err = client.RpcPost("Rds", "2014-08-15", action, nil, request, false)
		if err != nil {
			if IsExpectedErrors(err, []string{"InternalError"}) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		addDebug(action, response, request)
		return nil
	})
	if err != nil {
		return WrapError(err)
	}
	if err := s.WaitForDBInstance(d.Id(), Running, DefaultLongTimeout); err != nil {
		return WrapError(err)
	}

	desResponse, err := s.DescribePGHbaConfig(d.Id())
	if err != nil {
		return WrapError(err)
	}
	if desResponse["LastModifyStatus"] == "failed" {
		return WrapError(Error("%v", desResponse["ModifyStatusReason"].(string)))
	}
	d.SetPartial(attribute)
	return nil
}
