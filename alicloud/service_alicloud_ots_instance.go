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
	return WrapErrorf(err, IdMsg, instanceName)
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
	return WrapErrorf(err, IdMsg, instanceName)
}

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

func (s *OtsService) DescribeOtsInstanceAttachment(instanceName string) (*tablestoreAPI.TablestoreInstanceAttachment, error) {
	api, err := s.getTablestoreAPI()
	if err != nil {
		return nil, WrapError(err)
	}

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
