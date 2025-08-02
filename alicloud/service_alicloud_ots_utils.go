package alicloud

import (
	"regexp"

	aliyunTablestoreAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/tablestore"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

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
