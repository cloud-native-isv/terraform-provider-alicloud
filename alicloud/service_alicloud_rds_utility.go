package alicloud

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/PaesslerAG/jsonpath"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// RDS Tags and Utility related operations

func (s *RdsService) setInstanceTags(d *schema.ResourceData) error {
	if d.HasChange("tags") {
		added, removed := parsingTags(d)
		client := s.client
		var err error
		removedTagKeys := make([]string, 0)
		for _, v := range removed {
			if !ignoredTags(v, "") {
				removedTagKeys = append(removedTagKeys, v)
			}
		}
		if len(removedTagKeys) > 0 {
			action := "UnTagResources"
			request := map[string]interface{}{
				"RegionId":     s.client.RegionId,
				"ResourceType": "INSTANCE",
				"ResourceId.1": d.Id(),
				"SourceIp":     s.client.SourceIp,
			}
			for i, key := range removedTagKeys {
				request[fmt.Sprintf("TagKey.%d", i+1)] = key
			}
			wait := incrementalWait(1*time.Second, 2*time.Second)
			err = resource.Retry(10*time.Minute, func() *resource.RetryError {
				response, err := client.RpcPost("Rds", "2014-08-15", action, nil, request, false)
				if err != nil {
					if NeedRetry(err) {
						wait()
						return resource.RetryableError(err)
					}
					return resource.NonRetryableError(err)
				}
				addDebug(action, response, request)
				return nil
			})
			if err != nil {
				return WrapErrorf(err, DefaultErrorMsg, d.Id(), action, AlibabaCloudSdkGoERROR)
			}
		}

		if len(added) > 0 {
			action := "TagResources"
			request := map[string]interface{}{
				"RegionId":     s.client.RegionId,
				"ResourceType": "INSTANCE",
				"ResourceId.1": d.Id(),
			}
			count := 1
			for key, value := range added {
				request[fmt.Sprintf("Tag.%d.Key", count)] = key
				request[fmt.Sprintf("Tag.%d.Value", count)] = value
				count++
			}

			wait := incrementalWait(1*time.Second, 2*time.Second)
			err = resource.Retry(10*time.Minute, func() *resource.RetryError {
				response, err := client.RpcPost("Rds", "2014-08-15", action, nil, request, false)
				if err != nil {
					if NeedRetry(err) {
						wait()
						return resource.RetryableError(err)
					}
					return resource.NonRetryableError(err)
				}
				addDebug(action, response, request)
				return nil
			})
			if err != nil {
				return WrapErrorf(err, DefaultErrorMsg, d.Id(), action, AlibabaCloudSdkGoERROR)
			}
		}
		if err := s.WaitForDBInstance(d.Id(), Running, DefaultLongTimeout); err != nil {
			return WrapError(err)
		}
		d.SetPartial("tags")
	}

	return nil
}

func (s *RdsService) describeTags(d *schema.ResourceData) (tags []Tag, err error) {
	action := "DescribeTags"
	request := map[string]interface{}{
		"DBInstanceId": d.Id(),
		"RegionId":     s.client.RegionId,
		"SourceIp":     s.client.SourceIp,
	}
	client := s.client
	var response map[string]interface{}
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
		addDebug(action, response, request)
		return nil
	})
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, d.Id(), action, AlibabaCloudSdkGoERROR)
	}
	return s.respToTags(response["Items"].(map[string]interface{})["TagInfos"].([]interface{})), nil
}

func (s *RdsService) respToTags(tagSet []interface{}) (tags []Tag) {
	result := make([]Tag, 0, len(tagSet))
	for _, t := range tagSet {
		t := t.(map[string]interface{})
		tag := Tag{
			Key:   t["TagKey"].(string),
			Value: t["TagValue"].(string),
		}
		result = append(result, tag)
	}

	return result
}

func (s *RdsService) tagsToMap(tags []Tag) map[string]string {
	result := make(map[string]string)
	for _, t := range tags {
		if !s.ignoreTag(t) {
			result[t.Key] = t.Value
		}
	}

	return result
}

func (s *RdsService) ignoreTag(t Tag) bool {
	filter := []string{"^aliyun", "^acs:", "^http://", "^https://"}
	for _, v := range filter {
		log.Printf("[DEBUG] Matching prefix %v with %v\n", v, t.Key)
		ok, _ := regexp.MatchString(v, t.Key)
		if ok {
			log.Printf("[DEBUG] Found Alibaba Cloud specific t %s (val: %s), ignoring.\n", t.Key, t.Value)
			return true
		}
	}
	return false
}

func (s *RdsService) tagsToString(tags []Tag) string {
	v, _ := json.Marshal(s.tagsToMap(tags))

	return string(v)
}

// return multiIZ list of current region
func (s *RdsService) DescribeMultiIZByRegion() (izs []string, err error) {
	action := "DescribeRegions"
	request := map[string]interface{}{
		"RegionId": s.client.RegionId,
		"SourceIp": s.client.SourceIp,
	}
	client := s.client
	response, err := client.RpcPost("Rds", "2014-08-15", action, nil, request, true)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, "MultiIZByRegion", action, AlibabaCloudSdkGoERROR)
	}
	addDebug(action, response, request)
	regions := response["Regions"].(map[string]interface{})["RDSRegion"].([]interface{})
	zoneIds := []string{}
	for _, r := range regions {
		r := r.(map[string]interface{})
		if r["RegionId"] == string(s.client.Region) && strings.Contains(r["ZoneId"].(string), MULTI_IZ_SYMBOL) {
			zoneIds = append(zoneIds, r["ZoneId"].(string))
		}
	}

	return zoneIds, nil
}

func (s *RdsService) DescribeRdsParameterGroup(id string) (object map[string]interface{}, err error) {
	var response map[string]interface{}
	client := s.client
	action := "DescribeParameterGroup"
	request := map[string]interface{}{
		"RegionId":         s.client.RegionId,
		"ParameterGroupId": id,
	}
	response, err = client.RpcPost("Rds", "2014-08-15", action, nil, request, true)
	if err != nil {
		if IsExpectedErrors(err, []string{"ParamGroupsNotExistError"}) {
			err = WrapErrorf(NotFoundErr("RdsParameterGroup", id), NotFoundMsg, ProviderERROR)
			return object, err
		}
		err = WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
		return object, err
	}
	addDebug(action, response, request)
	v, err := jsonpath.Get("$.ParamGroup.ParameterGroup", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.ParamGroup.ParameterGroup", response)
	}
	if len(v.([]interface{})) < 1 {
		return object, WrapErrorf(NotFoundErr("RDS", id), NotFoundWithResponse, response)
	} else {
		if v.([]interface{})[0].(map[string]interface{})["ParameterGroupId"].(string) != id {
			return object, WrapErrorf(NotFoundErr("RDS", id), NotFoundWithResponse, response)
		}
	}
	object = v.([]interface{})[0].(map[string]interface{})
	return object, nil
}
