package alicloud

import (
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/ots"
	tablestoreAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/tablestore"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// Instance management functions

func (s *OtsService) CreateOtsInstance(instance *tablestoreAPI.TablestoreInstance) error {
	api, err := s.getTablestoreAPI()
	if err != nil {
		return WrapError(err)
	}

	if err := api.CreateInstance(instance); err != nil {
		return WrapErrorf(err, DefaultErrorMsg, instance.InstanceName, "CreateInstance", AlibabaCloudSdkGoERROR)
	}

	return nil
}

func (s *OtsService) DescribeOtsInstance(instanceName string) (*tablestoreAPI.TablestoreInstance, error) {
	api, err := s.getTablestoreAPI()
	if err != nil {
		return nil, WrapError(err)
	}

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
	api, err := s.getTablestoreAPI()
	if err != nil {
		return WrapError(err)
	}

	if err := api.UpdateInstance(instance); err != nil {
		return WrapErrorf(err, DefaultErrorMsg, instance.InstanceName, "UpdateInstance", AlibabaCloudSdkGoERROR)
	}

	return nil
}

func (s *OtsService) DeleteOtsInstance(instanceName string) error {
	api, err := s.getTablestoreAPI()
	if err != nil {
		return WrapError(err)
	}

	if err := api.DeleteInstance(instanceName); err != nil {
		if NotFoundError(err) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, instanceName, "DeleteInstance", AlibabaCloudSdkGoERROR)
	}

	return nil
}

func (s *OtsService) WaitForOtsInstance(instanceName string, status string, timeout int) error {
	deadline := time.Now().Add(time.Duration(timeout) * time.Second)
	for {
		instance, err := s.DescribeOtsInstance(instanceName)
		if err != nil {
			if NotFoundError(err) {
				if status == string(Deleted) {
					return nil
				}
			} else {
				return WrapError(err)
			}
		}

		if instance != nil && instance.InstanceStatus == status {
			return nil
		}

		if time.Now().After(deadline) {
			return WrapErrorf(err, WaitTimeoutMsg, instanceName, GetFunc(1), timeout, instance.InstanceStatus, status, ProviderERROR)
		}
		time.Sleep(DefaultIntervalShort * time.Second)
	}
}

func (s *OtsService) OtsInstanceStateRefreshFunc(instanceName string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeOtsInstance(instanceName)
		if err != nil {
			if NotFoundError(err) {
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		for _, failState := range failStates {
			if object.InstanceStatus == failState {
				return object, object.InstanceStatus, WrapError(Error(FailedToReachTargetStatus, object.InstanceStatus))
			}
		}
		return object, object.InstanceStatus, nil
	}
}

func (s *OtsService) WaitForOtsInstanceCreating(instanceName string, timeout time.Duration) error {
	stateConf := BuildStateConf(
		[]string{"Creating"}, // pending states
		[]string{"normal"},   // target states
		timeout,
		5*time.Second,
		s.OtsInstanceStateRefreshFunc(instanceName, []string{"forbidden", "deleting"}),
	)

	_, err := stateConf.WaitForState()
	return WrapErrorf(err, IdMsg, instanceName)
}

func (s *OtsService) WaitForOtsInstanceDeleting(instanceName string, timeout time.Duration) error {
	stateConf := BuildStateConf(
		[]string{"deleting"}, // pending states
		[]string{""},         // target states (not found)
		timeout,
		5*time.Second,
		s.OtsInstanceStateRefreshFunc(instanceName, []string{}),
	)

	_, err := stateConf.WaitForState()
	return WrapErrorf(err, IdMsg, instanceName)
}

func (s *OtsService) DescribeOtsInstanceTypes() ([]string, error) {
	// This would typically call an API to get available instance types
	// For now, return the known types
	return []string{"SSD", "HYBRID"}, nil
}

// Tag management functions

func (s *OtsService) TagOtsInstance(instanceName string, tags []tablestoreAPI.TablestoreInstanceTag) error {
	api, err := s.getTablestoreAPI()
	if err != nil {
		return WrapError(err)
	}

	if err := api.TagResources([]string{instanceName}, tags); err != nil {
		return WrapErrorf(err, DefaultErrorMsg, instanceName, "TagResources", AlibabaCloudSdkGoERROR)
	}

	return nil
}

func (s *OtsService) UntagOtsInstance(instanceName string, tagKeys []string) error {
	api, err := s.getTablestoreAPI()
	if err != nil {
		return WrapError(err)
	}

	if err := api.UntagResources([]string{instanceName}, tagKeys, false); err != nil {
		return WrapErrorf(err, DefaultErrorMsg, instanceName, "UntagResources", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// Instance VPC Attachment management functions

func (s *OtsService) DescribeOtsInstanceAttachment(instanceName string) (*ots.VpcInfo, error) {
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
	if len(response.VpcInfos.VpcInfo) == 0 {
		return nil, WrapErrorf(Error(GetNotFoundMessage("OtsInstanceAttachment", instanceName)), NotFoundMsg, ProviderERROR)
	}

	return &response.VpcInfos.VpcInfo[0], nil
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

// Additional helper functions for instance attachment

func (s *OtsService) BindOtsInstanceToVpc(instanceName, vpcName, vpcId, vswitchId string) error {
	request := ots.CreateBindInstance2VpcRequest()
	request.RegionId = s.client.RegionId
	request.InstanceName = instanceName
	request.InstanceVpcName = vpcName
	request.VirtualSwitchId = vswitchId
	request.VpcId = vpcId

	raw, err := s.client.WithOtsClient(func(otsClient *ots.Client) (interface{}, error) {
		return otsClient.BindInstance2Vpc(request)
	})
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, instanceName, request.GetActionName(), AlibabaCloudSdkGoERROR)
	}
	addDebug(request.GetActionName(), raw, request.RpcRequest, request)

	return nil
}

func (s *OtsService) UnbindOtsInstanceFromVpc(instanceName, vpcName string) error {
	request := ots.CreateUnbindInstance2VpcRequest()
	request.RegionId = s.client.RegionId
	request.InstanceName = instanceName
	request.InstanceVpcName = vpcName

	raw, err := s.client.WithOtsClient(func(otsClient *ots.Client) (interface{}, error) {
		return otsClient.UnbindInstance2Vpc(request)
	})
	if err != nil {
		if NotFoundError(err) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, instanceName, request.GetActionName(), AlibabaCloudSdkGoERROR)
	}
	addDebug(request.GetActionName(), raw, request.RpcRequest, request)

	return nil
}

// List OTS instances for data source support
func (s *OtsService) ListOtsInstance() ([]tablestoreAPI.TablestoreInstance, error) {
	api, err := s.getTablestoreAPI()
	if err != nil {
		return nil, WrapError(err)
	}

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
