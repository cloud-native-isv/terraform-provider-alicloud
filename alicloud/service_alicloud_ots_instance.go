package alicloud

import (
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/ots"
	tablestoreAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/tablestore"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// convertSchemaToTablestoreInstance converts Terraform schema data to TablestoreInstance
func convertSchemaToTablestoreInstance(d *schema.ResourceData) *tablestoreAPI.TablestoreInstance {
	instance := &tablestoreAPI.TablestoreInstance{
		InstanceName: d.Get("name").(string),
	}

	// Set instance specification (required field)
	if v, ok := d.GetOk("instance_specification"); ok {
		instance.InstanceSpecification = v.(string)
	}

	if v, ok := d.GetOk("alias_name"); ok {
		instance.AliasName = v.(string)
	}
	if v, ok := d.GetOk("description"); ok {
		instance.InstanceDescription = v.(string)
	}
	if v, ok := d.GetOk("policy"); ok {
		instance.Policy = v.(string)
	}
	if v, ok := d.GetOk("resource_group_id"); ok {
		instance.ResourceGroupId = v.(string)
	}

	// Convert network ACLs - only ACL fields are supported now (Network field is deprecated)
	if v, ok := d.GetOk("network_source_acl"); ok {
		if networkSourceAcl, ok := v.(*schema.Set); ok {
			instance.NetworkSourceACL = convertSetToStringSlice(networkSourceAcl)
		}
	}
	if v, ok := d.GetOk("network_type_acl"); ok {
		if networkTypeAcl, ok := v.(*schema.Set); ok {
			instance.NetworkTypeACL = convertSetToStringSlice(networkTypeAcl)
		}
	}

	// Convert tags
	if v, ok := d.GetOk("tags"); ok {
		if tagsMap, ok := v.(map[string]interface{}); ok {
			instance.Tags = convertMapToTablestoreInstanceTags(tagsMap)
		}
	}

	return instance
}

// convertTablestoreInstanceToSchema converts TablestoreInstance to Terraform schema data
func convertTablestoreInstanceToSchema(d *schema.ResourceData, instance *tablestoreAPI.TablestoreInstance) error {
	d.Set("name", instance.InstanceName)
	d.Set("instance_specification", instance.InstanceSpecification)
	d.Set("alias_name", instance.AliasName)
	d.Set("description", instance.InstanceDescription)
	d.Set("status", instance.InstanceStatus)
	d.Set("region_id", instance.RegionId)
	d.Set("resource_group_id", instance.ResourceGroupId)
	d.Set("payment_type", instance.PaymentType)
	d.Set("policy", instance.Policy)
	d.Set("policy_version", instance.PolicyVersion)
	d.Set("is_multi_az", instance.IsMultiAZ)
	d.Set("table_quota", instance.TableQuota)
	d.Set("vcu_quota", instance.VCUQuota)
	d.Set("elastic_vcu_upper_limit", instance.ElasticVCUUpperLimit)

	// 添加缺失的保留CU相关字段
	d.Set("is_reserved_cu_instance", instance.IsReservedCUInstance)
	d.Set("reserved_read_cu", instance.ReservedReadCU)
	d.Set("reserved_write_cu", instance.ReservedWriteCU)

	// 时间和用户ID字段
	if !instance.CreateTime.IsZero() {
		d.Set("create_time", instance.CreateTime.Format("2006-01-02T15:04:05Z"))
	}
	d.Set("user_id", instance.UserId)

	// Set network ACLs - only ACL fields are supported now (Network field is deprecated)
	if err := d.Set("network_source_acl", convertStringSliceToSet(instance.NetworkSourceACL)); err != nil {
		return err
	}
	if err := d.Set("network_type_acl", convertStringSliceToSet(instance.NetworkTypeACL)); err != nil {
		return err
	}

	// Set tags
	if err := d.Set("tags", convertTablestoreInstanceTagsToMap(instance.Tags)); err != nil {
		return err
	}

	return nil
}

// Instance management functions

func (s *OtsService) CreateOtsInstance(instance *tablestoreAPI.TablestoreInstance) error {
	api := s.GetAPI()

	if err := api.CreateInstance(instance); err != nil {
		return WrapErrorf(err, DefaultErrorMsg, instance.InstanceName, "CreateInstance", AlibabaCloudSdkGoERROR)
	}

	return nil
}

func (s *OtsService) DescribeOtsInstance(instanceName string) (*tablestoreAPI.TablestoreInstance, error) {
	api := s.GetAPI()

	instance, err := api.GetInstance(instanceName)
	if err != nil {
		if NotFoundError(err) {
			return nil, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
		}
		return nil, WrapErrorf(err, DefaultErrorMsg, instanceName, "GetInstance", AlibabaCloudSdkGoERROR)
	}

	return instance, nil
}

func (s *OtsService) UpdateOtsInstance(instance *tablestoreAPI.TablestoreInstance) error {
	api := s.GetAPI()

	if err := api.UpdateInstance(instance); err != nil {
		return WrapErrorf(err, DefaultErrorMsg, instance.InstanceName, "UpdateInstance", AlibabaCloudSdkGoERROR)
	}

	return nil
}

func (s *OtsService) DeleteOtsInstance(instanceName string) error {
	api := s.GetAPI()

	if err := api.DeleteInstance(instanceName); err != nil {
		if NotFoundError(err) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, instanceName, "DeleteInstance", AlibabaCloudSdkGoERROR)
	}

	return nil
}

func (s *OtsService) OtsInstanceStateRefreshFunc(instanceName string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeOtsInstance(instanceName)
		if err != nil {
			if NotFoundError(err) {
				return nil, tablestoreAPI.InstanceStatusNotFound.String(), nil
			}
			return nil, tablestoreAPI.InstanceStatusFailed.String(), WrapError(err)
		}

		currentStatus := object.InstanceStatus
		for _, failState := range failStates {
			if currentStatus == failState {
				return object, object.InstanceStatus, WrapError(Error(FailedToReachTargetStatus, object.InstanceStatus))
			}
		}
		return object, object.InstanceStatus, nil
	}
}

func (s *OtsService) WaitForOtsInstanceCreating(instanceName string, timeout time.Duration) error {
	stateConf := BuildStateConf(
		[]string{},
		[]string{tablestoreAPI.InstanceStatusRunning.String()},
		timeout,
		5*time.Second,
		s.OtsInstanceStateRefreshFunc(instanceName, []string{tablestoreAPI.InstanceStatusFailed.String()}),
	)

	_, err := stateConf.WaitForState()
	if err != nil {
		return WrapErrorf(err, IdMsg, instanceName)
	}
	return nil
}

func (s *OtsService) WaitForOtsInstanceDeleting(instanceName string, timeout time.Duration) error {
	stateConf := BuildStateConf(
		[]string{tablestoreAPI.InstanceStatusDeleting.String()},
		[]string{tablestoreAPI.InstanceStatusNotFound.String()},
		timeout,
		5*time.Second,
		s.OtsInstanceStateRefreshFunc(instanceName, []string{tablestoreAPI.InstanceStatusFailed.String()}),
	)

	_, err := stateConf.WaitForState()
	if err != nil {
		return WrapErrorf(err, IdMsg, instanceName)
	}
	return nil
}

func (s *OtsService) TagOtsInstance(instanceName string, tags []tablestoreAPI.TablestoreInstanceTag) error {
	api := s.GetAPI()

	if err := api.TagResources([]string{instanceName}, tags); err != nil {
		return WrapErrorf(err, DefaultErrorMsg, instanceName, "TagResources", AlibabaCloudSdkGoERROR)
	}

	return nil
}

func (s *OtsService) UntagOtsInstance(instanceName string, tagKeys []string) error {
	api := s.GetAPI()

	if err := api.UntagResources([]string{instanceName}, tagKeys, false); err != nil {
		return WrapErrorf(err, DefaultErrorMsg, instanceName, "UntagResources", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// Instance VPC Attachment management functions

func (s *OtsService) DescribeOtsInstanceAttachment(instanceName string) (*tablestoreAPI.TablestoreInstanceAttachment, error) {
	api := s.GetAPI()

	// Get all VPC attachments for the instance
	attachments, err := api.ListInstanceAttachments(instanceName)
	if err != nil {
		if NotFoundError(err) {
			return nil, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
		}
		return nil, WrapErrorf(err, DefaultErrorMsg, instanceName, "ListInstanceAttachments", AlibabaCloudSdkGoERROR)
	}

	// Return the first attachment if any exists
	// Note: In the original implementation, it returned the first VPC attachment
	// If you need a specific attachment, you should use GetInstanceAttachment with vpc name
	if len(attachments) == 0 {
		return nil, WrapErrorf(Error(GetNotFoundMessage("OtsInstanceAttachment", instanceName)), NotFoundMsg, ProviderERROR)
	}

	return &attachments[0], nil
}

func (s *OtsService) WaitForOtsInstanceVpc(instanceName string, status Status, timeout int) error {
	deadline := time.Now().Add(time.Duration(timeout) * time.Second)
	for {
		object, err := s.DescribeOtsInstanceAttachment(instanceName)
		if err != nil {
			if NotFoundError(err) {
				if status == Deleted {
					return nil
				}
			} else {
				return WrapError(err)
			}
		}
		if status != Deleted {
			return nil
		}
		if time.Now().After(deadline) {
			return WrapErrorf(err, WaitTimeoutMsg, instanceName, GetFunc(1), timeout, object, status, ProviderERROR)
		}
		time.Sleep(DefaultIntervalShort * time.Second)
	}
}

// List OTS instances for data source support
func (s *OtsService) ListOtsInstance() ([]tablestoreAPI.TablestoreInstance, error) {
	api := s.GetAPI()

	instances, err := api.ListAllInstances(nil)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, "ots_instances", "ListInstances", AlibabaCloudSdkGoERROR)
	}

	return instances, nil
}

// List OTS instance VPC attachments for data source support
func (s *OtsService) ListOtsInstanceVpc(instanceName string) ([]*ots.VpcInfo, error) {
	request := ots.CreateListVpcInfoByInstanceRequest()
	request.RegionId = s.client.RegionId
	request.InstanceName = instanceName

	raw, err := s.client.WithOtsClient(func(otsClient *ots.Client) (interface{}, error) {
		return otsClient.ListVpcInfoByInstance(request)
	})
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, instanceName, request.GetActionName(), AlibabaCloudSdkGoERROR)
	}
	addDebug(request.GetActionName(), raw, request.RpcRequest, request)

	response, _ := raw.(*ots.ListVpcInfoByInstanceResponse)

	// Convert to pointer slice
	var result []*ots.VpcInfo
	for i := range response.VpcInfos.VpcInfo {
		result = append(result, &response.VpcInfos.VpcInfo[i])
	}

	return result, nil
}
