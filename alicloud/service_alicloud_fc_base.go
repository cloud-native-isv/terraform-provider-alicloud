package alicloud

import (
	"fmt"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunCommonAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/common"
	aliyunFCAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/fc/v3"
)

type FCService struct {
	client *connectivity.AliyunClient
	fcAPI  *aliyunFCAPI.FCAPI
}

// NewFCService creates a new FCService using cws-lib-go implementation
func NewFCService(client *connectivity.AliyunClient) (*FCService, error) {
	// Convert AliyunClient credentials to Credentials
	credentials := &aliyunCommonAPI.Credentials{
		AccessKey:     client.AccessKey,
		SecretKey:     client.SecretKey,
		RegionId:      client.RegionId,
		SecurityToken: client.SecurityToken,
	}

	// Create the cws-lib-go FCAPI
	fcAPI, err := aliyunFCAPI.NewFCAPI(credentials)
	if err != nil {
		return nil, fmt.Errorf("failed to create cws-lib-go FCAPI: %w", err)
	}

	return &FCService{
		client: client,
		fcAPI:  fcAPI,
	}, nil
}

// GetAPI returns the FCAPI instance for direct API access
func (service *FCService) GetAPI() *aliyunFCAPI.FCAPI {
	// add some customize logic for this API object
	return service.fcAPI
}

// Configuration-related API methods (async_invoke_config, concurrency_config, provision_config, vpc_binding)
// These are placeholders until the APIs are implemented in cws-lib-go

// DescribeFCAsyncInvokeConfig retrieves async invoke configuration
func (service *FCService) DescribeFCAsyncInvokeConfig(functionName string) (map[string]interface{}, error) {
	// Placeholder implementation - return empty for now
	// TODO: Implement using cws-lib-go when available
	return make(map[string]interface{}), fmt.Errorf("DescribeFCAsyncInvokeConfig not yet implemented in FC v3 API")
}

// DescribeFCConcurrencyConfig retrieves concurrency configuration
func (service *FCService) DescribeFCConcurrencyConfig(functionName string) (map[string]interface{}, error) {
	// Placeholder implementation - return empty for now
	// TODO: Implement using cws-lib-go when available
	return make(map[string]interface{}), fmt.Errorf("DescribeFCConcurrencyConfig not yet implemented in FC v3 API")
}

// DescribeFCProvisionConfig retrieves provision configuration
func (service *FCService) DescribeFCProvisionConfig(functionName string) (map[string]interface{}, error) {
	// Placeholder implementation - return empty for now
	// TODO: Implement using cws-lib-go when available
	return make(map[string]interface{}), fmt.Errorf("DescribeFCProvisionConfig not yet implemented in FC v3 API")
}

// DescribeFCVpcBinding retrieves VPC binding configuration
func (service *FCService) DescribeFCVpcBinding(functionName string) (map[string]interface{}, error) {
	// Placeholder implementation - return empty for now
	// TODO: Implement using cws-lib-go when available
	return make(map[string]interface{}), fmt.Errorf("DescribeFCVpcBinding not yet implemented in FC v3 API")
}
