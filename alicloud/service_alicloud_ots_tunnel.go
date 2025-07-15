package alicloud

import (
	"context"
	"fmt"
	"time"

	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/tablestore"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// Tunnel management functions

func (s *OtsService) CreateOtsTunnel(d *schema.ResourceData, instanceName, tableName string) error {
	api, err := s.getTablestoreAPI()
	if err != nil {
		return WrapError(err)
	}

	tunnelName := d.Get("tunnel_name").(string)
	tunnelTypeStr := d.Get("tunnel_type").(string)

	// Convert tunnel type string to enum
	var tunnelType tablestore.TunnelType
	switch tunnelTypeStr {
	case "BaseData":
		tunnelType = tablestore.TunnelType_BaseData
	case "Stream":
		tunnelType = tablestore.TunnelType_Stream
	case "BaseAndStream":
		tunnelType = tablestore.TunnelType_BaseAndStream
	default:
		tunnelType = tablestore.TunnelType_Stream
	}

	// Create tunnel options
	options := &tablestore.CreateTunnelOptions{
		TableName:    tableName,
		TunnelName:   tunnelName,
		TunnelType:   tunnelType,
		InstanceName: instanceName,
	}

	ctx := context.Background()
	result, err := api.CreateTunnel(ctx, options)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, tunnelName, "CreateTunnel", AlibabaCloudSdkGoERROR)
	}

	d.SetId(fmt.Sprintf("%s:%s:%s", instanceName, tableName, result.TunnelId))
	return nil
}

func (s *OtsService) DescribeOtsTunnel(instanceName, tableName, tunnelName string) (*tablestore.TunnelInfo, error) {
	api, err := s.getTablestoreAPI()
	if err != nil {
		return nil, WrapError(err)
	}

	options := &tablestore.DescribeTunnelOptions{
		TableName:  tableName,
		TunnelName: tunnelName,
	}

	ctx := context.Background()
	result, err := api.DescribeTunnel(ctx, options)
	if err != nil {
		if IsExpectedErrors(err, []string{"NotExist", "InvalidTunnelName.NotFound"}) {
			return nil, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
		}
		return nil, WrapErrorf(err, DefaultErrorMsg, tunnelName, "DescribeTunnel", AlibabaCloudSdkGoERROR)
	}

	return &result.TunnelInfo, nil
}

func (s *OtsService) DeleteOtsTunnel(instanceName, tableName, tunnelName string) error {
	api, err := s.getTablestoreAPI()
	if err != nil {
		return WrapError(err)
	}

	options := &tablestore.DeleteTunnelOptions{
		TableName:  tableName,
		TunnelName: tunnelName,
	}

	ctx := context.Background()
	err = api.DeleteTunnel(ctx, options)
	if err != nil {
		if IsExpectedErrors(err, []string{"NotExist", "InvalidTunnelName.NotFound"}) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, tunnelName, "DeleteTunnel", AlibabaCloudSdkGoERROR)
	}

	return nil
}

func (s *OtsService) WaitForOtsTunnel(instanceName, tableName, tunnelName string, status string, timeout int) error {
	deadline := time.Now().Add(time.Duration(timeout) * time.Second)
	for {
		tunnel, err := s.DescribeOtsTunnel(instanceName, tableName, tunnelName)
		if err != nil {
			if NotFoundError(err) {
				if status == string(Deleted) {
					return nil
				}
			} else {
				return WrapError(err)
			}
		}

		if tunnel != nil && tunnel.Stage.String() == status {
			return nil
		}

		if time.Now().After(deadline) {
			return WrapErrorf(err, WaitTimeoutMsg, tunnelName, GetFunc(1), timeout, tunnel.Stage.String(), status, ProviderERROR)
		}
		time.Sleep(DefaultIntervalShort * time.Second)
	}
}

func (s *OtsService) ListOtsTunnels(instanceName, tableName string) ([]*tablestore.TunnelInfo, error) {
	api, err := s.getTablestoreAPI()
	if err != nil {
		return nil, WrapError(err)
	}

	options := &tablestore.ListTunnelOptions{
		TableName: tableName,
	}

	ctx := context.Background()
	result, err := api.ListTunnel(ctx, options)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, tableName, "ListTunnel", AlibabaCloudSdkGoERROR)
	}

	// Convert slice to pointer slice
	var tunnels []*tablestore.TunnelInfo
	for i := range result.Tunnels {
		tunnels = append(tunnels, &result.Tunnels[i])
	}

	return tunnels, nil
}

// Stream management functions

func (s *OtsService) ListOtsStreams(instanceName, tableName string) (*tablestore.ListStreamResult, error) {
	api, err := s.getTablestoreAPI()
	if err != nil {
		return nil, WrapError(err)
	}

	options := &tablestore.ListStreamOptions{
		TableName: tableName,
	}

	ctx := context.Background()
	result, err := api.ListStream(ctx, options)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, tableName, "ListStream", AlibabaCloudSdkGoERROR)
	}

	return result, nil
}

func (s *OtsService) DescribeOtsStream(instanceName, tableName, streamId string) (*tablestore.DescribeStreamResult, error) {
	api, err := s.getTablestoreAPI()
	if err != nil {
		return nil, WrapError(err)
	}

	options := &tablestore.DescribeStreamOptions{
		StreamId: streamId,
	}

	ctx := context.Background()
	stream, err := api.DescribeStream(ctx, options)
	if err != nil {
		if IsExpectedErrors(err, []string{"NotExist", "InvalidStreamId.NotFound"}) {
			return nil, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
		}
		return nil, WrapErrorf(err, DefaultErrorMsg, streamId, "DescribeStream", AlibabaCloudSdkGoERROR)
	}

	return stream, nil
}
