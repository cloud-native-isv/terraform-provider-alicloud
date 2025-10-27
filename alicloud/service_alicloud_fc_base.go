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
func (service *FCService) DescribeFCAsyncInvokeConfig(functionName string) (*aliyunFCAPI.AsyncConfig, error) {
	if functionName == "" {
		return nil, fmt.Errorf("function name cannot be empty")
	}
	return service.GetAPI().GetAsyncInvokeConfig(functionName)
}

// CreateFCAsyncInvokeConfig creates async invoke configuration
func (service *FCService) CreateFCAsyncInvokeConfig(functionName string, config *aliyunFCAPI.AsyncConfig) (*aliyunFCAPI.AsyncConfig, error) {
	if functionName == "" {
		return nil, fmt.Errorf("function name cannot be empty")
	}
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	return service.GetAPI().PutAsyncInvokeConfig(functionName, config)
}

// UpdateFCAsyncInvokeConfig updates async invoke configuration
func (service *FCService) UpdateFCAsyncInvokeConfig(functionName string, config *aliyunFCAPI.AsyncConfig) (*aliyunFCAPI.AsyncConfig, error) {
	if functionName == "" {
		return nil, fmt.Errorf("function name cannot be empty")
	}
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	return service.GetAPI().PutAsyncInvokeConfig(functionName, config)
}

// DeleteFCAsyncInvokeConfig deletes async invoke configuration
func (service *FCService) DeleteFCAsyncInvokeConfig(functionName string) error {
	if functionName == "" {
		return fmt.Errorf("function name cannot be empty")
	}
	return service.GetAPI().DeleteAsyncInvokeConfig(functionName)
}

// DescribeFCConcurrencyConfig retrieves concurrency configuration
func (service *FCService) DescribeFCConcurrencyConfig(functionName string) (*aliyunFCAPI.ConcurrencyConfig, error) {
	if functionName == "" {
		return nil, fmt.Errorf("function name cannot be empty")
	}
	return service.GetAPI().GetConcurrencyConfig(functionName)
}

// CreateFCConcurrencyConfig creates concurrency configuration
func (service *FCService) CreateFCConcurrencyConfig(functionName string, config *aliyunFCAPI.ConcurrencyConfig) (*aliyunFCAPI.ConcurrencyConfig, error) {
	if functionName == "" {
		return nil, fmt.Errorf("function name cannot be empty")
	}
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	return service.GetAPI().PutConcurrencyConfig(functionName, config)
}

// UpdateFCConcurrencyConfig updates concurrency configuration
func (service *FCService) UpdateFCConcurrencyConfig(functionName string, config *aliyunFCAPI.ConcurrencyConfig) (*aliyunFCAPI.ConcurrencyConfig, error) {
	if functionName == "" {
		return nil, fmt.Errorf("function name cannot be empty")
	}
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	return service.GetAPI().PutConcurrencyConfig(functionName, config)
}

// DeleteFCConcurrencyConfig deletes concurrency configuration
func (service *FCService) DeleteFCConcurrencyConfig(functionName string) error {
	if functionName == "" {
		return fmt.Errorf("function name cannot be empty")
	}
	return service.GetAPI().DeleteConcurrencyConfig(functionName)
}

// DescribeFCProvisionConfig retrieves provision configuration
func (service *FCService) DescribeFCProvisionConfig(functionName string, qualifier string) (*aliyunFCAPI.ProvisionConfig, error) {
	if functionName == "" {
		return nil, fmt.Errorf("function name cannot be empty")
	}
	if qualifier == "" {
		return nil, fmt.Errorf("qualifier cannot be empty")
	}
	return service.GetAPI().GetProvisionConfig(functionName, qualifier)
}

// CreateFCProvisionConfig creates provision configuration
func (service *FCService) CreateFCProvisionConfig(functionName string, qualifier string, config *aliyunFCAPI.ProvisionConfig) (*aliyunFCAPI.ProvisionConfig, error) {
	if functionName == "" {
		return nil, fmt.Errorf("function name cannot be empty")
	}
	if qualifier == "" {
		return nil, fmt.Errorf("qualifier cannot be empty")
	}
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	return service.GetAPI().PutProvisionConfig(functionName, qualifier, config)
}

// UpdateFCProvisionConfig updates provision configuration
func (service *FCService) UpdateFCProvisionConfig(functionName string, qualifier string, config *aliyunFCAPI.ProvisionConfig) (*aliyunFCAPI.ProvisionConfig, error) {
	if functionName == "" {
		return nil, fmt.Errorf("function name cannot be empty")
	}
	if qualifier == "" {
		return nil, fmt.Errorf("qualifier cannot be empty")
	}
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	return service.GetAPI().PutProvisionConfig(functionName, qualifier, config)
}

// DeleteFCProvisionConfig deletes provision configuration
func (service *FCService) DeleteFCProvisionConfig(functionName string, qualifier string) error {
	if functionName == "" {
		return fmt.Errorf("function name cannot be empty")
	}
	if qualifier == "" {
		return fmt.Errorf("qualifier cannot be empty")
	}
	return service.GetAPI().DeleteProvisionConfig(functionName, qualifier)
}

// DescribeFCVpcBinding retrieves VPC binding configuration
func (service *FCService) DescribeFCVpcBinding(functionName string) ([]*string, error) {
	if functionName == "" {
		return nil, fmt.Errorf("function name cannot be empty")
	}
	return service.GetAPI().ListVpcBindings(functionName)
}

// CreateFCVpcBinding creates VPC binding configuration
func (service *FCService) CreateFCVpcBinding(functionName string, vpcId string) error {
	if functionName == "" {
		return fmt.Errorf("function name cannot be empty")
	}
	if vpcId == "" {
		return fmt.Errorf("VPC ID cannot be empty")
	}
	return service.GetAPI().CreateVpcBinding(functionName, vpcId)
}

// DeleteFCVpcBinding deletes VPC binding configuration
func (service *FCService) DeleteFCVpcBinding(functionName string, vpcId string) error {
	if functionName == "" {
		return fmt.Errorf("function name cannot be empty")
	}
	if vpcId == "" {
		return fmt.Errorf("VPC ID cannot be empty")
	}
	return service.GetAPI().DeleteVpcBinding(functionName, vpcId)
}
