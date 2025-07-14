package alicloud

import (
	"encoding/json"
	"fmt"
	"time"

	util "github.com/alibabacloud-go/tea-utils/service"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// RDS Parameter related operations

func (s *RdsService) DescribeParameters(id string) (object map[string]interface{}, err error) {
	client := s.client
	action := "DescribeParameters"
	request := map[string]interface{}{
		"RegionId":     s.client.RegionId,
		"DBInstanceId": id,
		"SourceIp":     s.client.SourceIp,
	}
	runtime := util.RuntimeOptions{}
	runtime.SetAutoretry(true)
	response, err := client.RpcPost("Rds", "2014-08-15", action, nil, request, true)
	if err != nil {
		if IsExpectedErrors(err, []string{"InvalidDBInstanceId.NotFound"}) {
			return nil, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
		}
		return nil, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	addDebug(action, response, request)
	return response, err
}

func (s *RdsService) DescribeParameterTemplates(instanceId, engine, engineVersion string) ([]interface{}, error) {
	action := "DescribeParameterTemplates"
	request := map[string]interface{}{
		"RegionId":      s.client.RegionId,
		"DBInstanceId":  instanceId,
		"Engine":        engine,
		"EngineVersion": engineVersion,
		"SourceIp":      s.client.SourceIp,
	}
	client := s.client
	response, err := client.RpcPost("Rds", "2014-08-15", action, nil, request, true)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, instanceId, action, AlibabaCloudSdkGoERROR)
	}
	addDebug(action, response, request)
	return response["Parameters"].(map[string]interface{})["TemplateRecord"].([]interface{}), nil
}

func (s *RdsService) SetTimeZone(d *schema.ResourceData) error {
	targetParameterName := ""
	engine := d.Get("engine")
	if engine == string(MySQL) {
		targetParameterName = "default_time_zone"
	} else if engine == string(PostgreSQL) {
		targetParameterName = "timezone"
	}

	if targetParameterName != "" {
		paramsRes, err := s.DescribeParameters(d.Id())
		if err != nil {
			return WrapError(err)
		}
		parameters := paramsRes["RunningParameters"].(map[string]interface{})["DBInstanceParameter"].([]interface{})
		for _, item := range parameters {
			item := item.(map[string]interface{})
			parameterName := item["ParameterName"]

			if parameterName == targetParameterName {
				d.Set("db_time_zone", item["ParameterValue"])
				break
			}
		}
	}
	return nil
}

func (s *RdsService) RefreshParameters(d *schema.ResourceData, attribute string) error {
	var param []map[string]interface{}
	documented, ok := d.GetOk(attribute)
	if !ok {
		return nil
	}
	object, err := s.DescribeParameters(d.Id())
	if err != nil {
		return WrapError(err)
	}

	var parameters = make(map[string]interface{})
	dBInstanceParameters := object["RunningParameters"].(map[string]interface{})["DBInstanceParameter"].([]interface{})
	for _, i := range dBInstanceParameters {
		i := i.(map[string]interface{})
		if i["ParameterName"] != "" {
			parameter := map[string]interface{}{
				"name":  i["ParameterName"],
				"value": i["ParameterValue"],
			}
			parameters[i["ParameterName"].(string)] = parameter
		}
	}
	dBInstanceParameters = object["ConfigParameters"].(map[string]interface{})["DBInstanceParameter"].([]interface{})
	for _, i := range dBInstanceParameters {
		i := i.(map[string]interface{})
		if i["ParameterName"] != "" {
			parameter := map[string]interface{}{
				"name":  i["ParameterName"],
				"value": i["ParameterValue"],
			}
			parameters[i["ParameterName"].(string)] = parameter
		}
	}

	for _, parameter := range documented.(*schema.Set).List() {
		name := parameter.(map[string]interface{})["name"]
		for _, value := range parameters {
			if value.(map[string]interface{})["name"] == name {
				param = append(param, value.(map[string]interface{}))
				break
			}
		}
	}
	if len(param) > 0 {
		if err := d.Set(attribute, param); err != nil {
			return WrapError(err)
		}
	}
	return nil
}

func (s *RdsService) ModifyParameters(d *schema.ResourceData, attribute string) error {
	client := s.client
	action := "ModifyParameter"
	request := map[string]interface{}{
		"RegionId":     s.client.RegionId,
		"DBInstanceId": d.Id(),
		"Forcerestart": d.Get("force_restart"),
		"SourceIp":     s.client.SourceIp,
	}
	config := make(map[string]string)
	allConfig := make(map[string]string)
	o, n := d.GetChange(attribute)
	os, ns := o.(*schema.Set), n.(*schema.Set)
	add := ns.Difference(os).List()
	if len(add) > 0 {
		for _, i := range add {
			key := i.(map[string]interface{})["name"].(string)
			value := i.(map[string]interface{})["value"].(string)
			config[key] = value
		}
		cfg, _ := json.Marshal(config)
		request["Parameters"] = string(cfg)
		// wait instance status is Normal before modifying
		if err := s.WaitForDBInstance(d.Id(), Running, DefaultLongTimeout); err != nil {
			return WrapError(err)
		}
		// Need to check whether some parameter needs restart
		if !d.Get("force_restart").(bool) {
			action := "DescribeParameterTemplates"
			request := map[string]interface{}{
				"RegionId":      s.client.RegionId,
				"DBInstanceId":  d.Id(),
				"Engine":        d.Get("engine"),
				"EngineVersion": d.Get("engine_version"),
				"ClientToken":   buildClientToken(action),
				"SourceIp":      s.client.SourceIp,
			}
			forceRestartMap := make(map[string]string)
			response, err := client.RpcPost("Rds", "2014-08-15", action, nil, request, false)
			if err != nil {
				return WrapErrorf(err, DefaultErrorMsg, d.Id(), action, AlibabaCloudSdkGoERROR)
			}
			addDebug(action, response, request)
			stateConf := BuildStateConf([]string{}, []string{"Running"}, d.Timeout(schema.TimeoutUpdate), 3*time.Minute, s.RdsDBInstanceStateRefreshFunc(d.Id(), []string{"Deleting"}))
			if _, err := stateConf.WaitForState(); err != nil {
				return WrapErrorf(err, IdMsg, d.Id())
			}
			templateRecords := response["Parameters"].(map[string]interface{})["TemplateRecord"].([]interface{})
			for _, para := range templateRecords {
				para := para.(map[string]interface{})
				if para["ForceRestart"] == "true" {
					forceRestartMap[para["ParameterName"].(string)] = para["ForceRestart"].(string)
				}
			}
			if len(forceRestartMap) > 0 {
				for key := range config {
					if _, ok := forceRestartMap[key]; ok {
						return WrapError(fmt.Errorf("Modifying RDS instance's parameter '%s' requires setting 'force_restart = true'.", key))
					}
				}
			}
		}
		response, err := client.RpcPost("Rds", "2014-08-15", action, nil, request, true)
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), action, AlibabaCloudSdkGoERROR)
		}
		addDebug(action, response, request)
		// wait instance parameter expect after modifying
		for _, i := range ns.List() {
			key := i.(map[string]interface{})["name"].(string)
			value := i.(map[string]interface{})["value"].(string)
			allConfig[key] = value
		}
		if err := s.WaitForDBParameter(d.Id(), DefaultLongTimeout, allConfig); err != nil {
			return WrapError(err)
		}
		// wait instance status is Normal after modifying
		if err := s.WaitForDBInstance(d.Id(), Running, DefaultLongTimeout); err != nil {
			return WrapError(err)
		}
	}
	d.SetPartial(attribute)
	return nil
}

// WaitForDBParameter waits for instance parameter to given value.
// Status of DB instance is Running after ModifyParameters API was
// call, so we can not just wait for instance status become
// Running, we should wait until parameters have expected values.
func (s *RdsService) WaitForDBParameter(instanceId string, timeout int, expects map[string]string) error {
	deadline := time.Now().Add(time.Duration(timeout) * time.Second)
	for {
		object, err := s.DescribeParameters(instanceId)
		if err != nil {
			return WrapError(err)
		}
		var actuals = make(map[string]string)
		dBInstanceParameters := object["RunningParameters"].(map[string]interface{})["DBInstanceParameter"].([]interface{})
		for _, i := range dBInstanceParameters {
			i := i.(map[string]interface{})
			if i["ParameterName"] == nil || i["ParameterValue"] == nil {
				continue
			}
			actuals[i["ParameterName"].(string)] = i["ParameterValue"].(string)
		}
		dBInstanceParameters = object["ConfigParameters"].(map[string]interface{})["DBInstanceParameter"].([]interface{})
		for _, i := range dBInstanceParameters {
			i := i.(map[string]interface{})
			if i["ParameterName"] == nil || i["ParameterValue"] == nil {
				continue
			}
			actuals[i["ParameterName"].(string)] = i["ParameterValue"].(string)
		}

		match := true

		got_value := ""
		expected_value := ""

		for name, expect := range expects {
			if actual, ok := actuals[name]; ok {
				if expect != actual {
					match = false
					got_value = actual
					expected_value = expect
					break
				}
			} else {
				match = false
			}
		}

		if match {
			break
		}

		time.Sleep(DefaultIntervalShort * time.Second)

		if time.Now().After(deadline) {
			return WrapErrorf(err, WaitTimeoutMsg, instanceId, GetFunc(1), timeout, got_value, expected_value, ProviderERROR)
		}
	}
	return nil
}
