package alicloud

import (
	"fmt"
	"regexp"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/common"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	aliyunTablestoreAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/tablestore"
)

type OtsService struct {
	client        *connectivity.AliyunClient
	tablestoreAPI *aliyunTablestoreAPI.TablestoreAPI
}

func (s *OtsService) getTablestoreAPI() (*aliyunTablestoreAPI.TablestoreAPI, error) {
	if s.tablestoreAPI != nil {
		return s.tablestoreAPI, nil
	}

	// Create new API instance if not exists
	tablestoreAPI, err := aliyunTablestoreAPI.NewTablestoreAPI(&common.ConnectionConfig{
		Credentials: common.Credentials{
			AccessKey:     s.client.AccessKey,
			SecretKey:     s.client.SecretKey,
			RegionId:      s.client.RegionId,
			SecurityToken: s.client.SecurityToken,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Tablestore API client: %w", err)
	}

	s.tablestoreAPI = tablestoreAPI
	return s.tablestoreAPI, nil
}

// NewSlsService creates a new SlsService instance with initialized clients
func NewOtsService(client *connectivity.AliyunClient) (*OtsService, error) {
	tablestoreAPI, err := aliyunTablestoreAPI.NewTablestoreAPI(&common.ConnectionConfig{
		Credentials: common.Credentials{
			AccessKey:     client.AccessKey,
			SecretKey:     client.SecretKey,
			RegionId:      client.RegionId,
			SecurityToken: client.SecurityToken,
		},
	})
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

// Helper functions for schema conversion

// Helper conversion functions

func convertSetToStringSlice(set *schema.Set) []string {
	if set == nil {
		return []string{}
	}

	list := set.List()
	result := make([]string, len(list))
	for i, v := range list {
		result[i] = v.(string)
	}
	return result
}

func convertStringSliceToSet(slice []string) *schema.Set {
	set := schema.NewSet(schema.HashString, []interface{}{})
	for _, str := range slice {
		set.Add(str)
	}
	return set
}

func convertMapToTablestoreInstanceTags(tagsMap map[string]interface{}) []aliyunTablestoreAPI.TablestoreInstanceTag {
	var tags []aliyunTablestoreAPI.TablestoreInstanceTag
	for key, value := range tagsMap {
		tags = append(tags, aliyunTablestoreAPI.TablestoreInstanceTag{
			Key:   key,
			Value: value.(string),
		})
	}
	return tags
}

func convertTablestoreInstanceTagsToMap(tags []aliyunTablestoreAPI.TablestoreInstanceTag) map[string]interface{} {
	tagsMap := make(map[string]interface{})
	for _, tag := range tags {
		tagsMap[tag.Key] = tag.Value
	}
	return tagsMap
}
