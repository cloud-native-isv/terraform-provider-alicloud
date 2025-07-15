package alicloud

import (
	"context"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/ots"
	tablestoreSDK "github.com/aliyun/aliyun-tablestore-go-sdk/tablestore"
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/common"
	tablestoreAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/tablestore"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// Instance management functions

func (s *OtsService) CreateOtsInstance(d *schema.ResourceData) error {
	api, err := s.getTablestoreAPI()
	if err != nil {
		return WrapError(err)
	}

	instanceName := d.Get("name").(string)
	options := &tablestoreAPI.CreateTablestoreInstanceOptions{
		InstanceName: instanceName,
		ClusterType:  d.Get("instance_type").(string),
	}

	if description, ok := d.GetOk("description"); ok {
		options.InstanceDescription = description.(string)
	}

	if resourceGroupId, ok := d.GetOk("resource_group_id"); ok {
		options.ResourceGroupId = resourceGroupId.(string)
	}

	if networkTypeACL, ok := d.GetOk("network_type_acl"); ok {
		options.NetworkTypeACL = expandStringList(networkTypeACL.(*schema.Set).List())
	}

	if networkSourceACL, ok := d.GetOk("network_source_acl"); ok {
		options.NetworkSourceACL = expandStringList(networkSourceACL.(*schema.Set).List())
	}

	if tags, ok := d.GetOk("tags"); ok {
		options.Tags = convertToTablestoreTags(tags.(map[string]interface{}))
	}

	ctx := context.Background()
	if err := api.CreateInstance(ctx, options); err != nil {
		return WrapErrorf(err, DefaultErrorMsg, instanceName, "CreateInstance", AlibabaCloudSdkGoERROR)
	}

	d.SetId(instanceName)
	return nil
}

func (s *OtsService) DescribeOtsInstance(instanceName string) (*tablestoreAPI.TablestoreInstance, error) {
	api, err := s.getTablestoreAPI()
	if err != nil {
		return nil, WrapError(err)
	}

	ctx := context.Background()
	instance, err := api.GetInstance(ctx, instanceName)
	if err != nil {
		if IsExpectedErrors(err, []string{"NotExist", "InvalidInstanceName.NotFound"}) {
			return nil, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
		}
		return nil, WrapErrorf(err, DefaultErrorMsg, instanceName, "GetInstance", AlibabaCloudSdkGoERROR)
	}

	return instance, nil
}

func (s *OtsService) UpdateOtsInstance(d *schema.ResourceData, instanceName string) error {
	api, err := s.getTablestoreAPI()
	if err != nil {
		return WrapError(err)
	}

	options := &tablestoreAPI.UpdateTablestoreInstanceOptions{}
	update := false

	if d.HasChange("resource_group_id") {
		options.ResourceGroupId = d.Get("resource_group_id").(string)
		update = true
	}

	if d.HasChange("network_type_acl") {
		options.NetworkTypeACL = expandStringList(d.Get("network_type_acl").(*schema.Set).List())
		update = true
	}

	if d.HasChange("network_source_acl") {
		options.NetworkSourceACL = expandStringList(d.Get("network_source_acl").(*schema.Set).List())
		update = true
	}

	if update {
		ctx := context.Background()
		if err := api.UpdateInstance(ctx, instanceName, options); err != nil {
			return WrapErrorf(err, DefaultErrorMsg, instanceName, "UpdateInstance", AlibabaCloudSdkGoERROR)
		}
	}

	return nil
}

func (s *OtsService) DeleteOtsInstance(instanceName string) error {
	api, err := s.getTablestoreAPI()
	if err != nil {
		return WrapError(err)
	}

	ctx := context.Background()
	if err := api.DeleteInstance(ctx, instanceName); err != nil {
		if IsExpectedErrors(err, []string{"NotExist", "InvalidInstanceName.NotFound"}) {
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

func (s *OtsService) DescribeOtsInstanceTypes() ([]string, error) {
	// This would typically call an API to get available instance types
	// For now, return the known types
	return []string{"SSD", "HYBRID"}, nil
}

// Tag management functions

func (s *OtsService) TagOtsInstance(instanceName string, tags map[string]string) error {
	api, err := s.getTablestoreAPI()
	if err != nil {
		return WrapError(err)
	}

	var tablestoreTags []tablestoreAPI.TablestoreInstanceTag
	for key, value := range tags {
		tablestoreTags = append(tablestoreTags, tablestoreAPI.TablestoreInstanceTag{
			Key:   key,
			Value: value,
		})
	}

	options := &tablestoreAPI.TagResourcesOptions{
		ResourceType: "instance",
		ResourceIds:  []string{instanceName},
		Tags:         tablestoreTags,
	}

	ctx := context.Background()
	if err := api.TagResources(ctx, options); err != nil {
		return WrapErrorf(err, DefaultErrorMsg, instanceName, "TagResources", AlibabaCloudSdkGoERROR)
	}

	return nil
}

func (s *OtsService) UntagOtsInstance(instanceName string, tagKeys []string) error {
	api, err := s.getTablestoreAPI()
	if err != nil {
		return WrapError(err)
	}

	options := &tablestoreAPI.UntagResourcesOptions{
		ResourceType: "instance",
		ResourceIds:  []string{instanceName},
		TagKeys:      tagKeys,
	}

	ctx := context.Background()
	if err := api.UntagResources(ctx, options); err != nil {
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

// Helper functions

func convertToTablestoreTags(tags map[string]interface{}) []tablestoreAPI.TablestoreInstanceTag {
	var result []tablestoreAPI.TablestoreInstanceTag
	for key, value := range tags {
		result = append(result, tablestoreAPI.TablestoreInstanceTag{
			Key:   key,
			Value: value.(string),
		})
	}
	return result
}

func convertTablestoreTagsToMap(tags []tablestoreAPI.TablestoreInstanceTag) map[string]string {
	result := make(map[string]string)
	for _, tag := range tags {
		result[tag.Key] = tag.Value
	}
	return result
}

// List OTS instances for data source support
func (s *OtsService) ListOtsInstance() ([]*tablestoreAPI.TablestoreInstance, error) {
	api, err := s.getTablestoreAPI()
	if err != nil {
		return nil, WrapError(err)
	}

	ctx := context.Background()
	instances, err := api.ListAllInstances(ctx, nil)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, "ots_instances", "ListInstances", AlibabaCloudSdkGoERROR)
	}

	// Convert slice to pointer slice
	var result []*tablestoreAPI.TablestoreInstance
	for i := range instances {
		result = append(result, &instances[i])
	}

	return result, nil
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
