package alicloud

import (
	"fmt"
	"regexp"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/common"

	aliyunTablestoreAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/tablestore"
)

type OtsService struct {
	client        *connectivity.AliyunClient
	tablestoreAPI *aliyunTablestoreAPI.TablestoreAPI
}

func (s *OtsService) getTablestoreAPI() (*aliyunTablestoreAPI.TablestoreAPI, error) {
	credentials := &common.Credentials{
		AccessKey: s.client.AccessKey,
		SecretKey: s.client.SecretKey,
		RegionId:  s.client.RegionId,
	}
	return aliyunTablestoreAPI.NewTablestoreAPI(credentials)
}

// NewSlsService creates a new SlsService instance with initialized clients
func NewOtsService(client *connectivity.AliyunClient) (*OtsService, error) {
	credentials := &common.Credentials{
		AccessKey:     client.AccessKey,
		SecretKey:     client.SecretKey,
		RegionId:      client.RegionId,
		SecurityToken: client.SecurityToken,
	}

	tablestoreAPI, err := aliyunTablestoreAPI.NewTablestoreAPI(credentials)
	if err != nil {
		return nil, fmt.Errorf("failed to create Tablestore API client: %w", err)
	}

	return &OtsService{
		client:        client,
		tablestoreAPI: tablestoreAPI,
	}, nil
}


// Helper types for data source filtering
type InputDataSource struct {
	inputs  []interface{}
	filters []DataSourceFilter
}

type DataSourceFilter interface {
	Filter(input interface{}) bool
}

type ValuesFilter struct {
	allowedValues  []interface{}
	getSourceValue func(interface{}) interface{}
}

func (f *ValuesFilter) Filter(input interface{}) bool {
	value := f.getSourceValue(input)
	for _, allowed := range f.allowedValues {
		if value == allowed {
			return true
		}
	}
	return false
}

type RegxFilter struct {
	regx           *regexp.Regexp
	getSourceValue func(interface{}) interface{}
}

func (f *RegxFilter) Filter(input interface{}) bool {
	value := f.getSourceValue(input)
	if str, ok := value.(string); ok {
		return f.regx.MatchString(str)
	}
	return false
}

func (ds *InputDataSource) doFilters() []interface{} {
	var result []interface{}
	for _, input := range ds.inputs {
		include := true
		for _, filter := range ds.filters {
			if !filter.Filter(input) {
				include = false
				break
			}
		}
		if include {
			result = append(result, input)
		}
	}
	return result
}
