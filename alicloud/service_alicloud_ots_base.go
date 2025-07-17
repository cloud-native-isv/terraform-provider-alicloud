package alicloud

import (
	"fmt"
	"regexp"

	consoleSDK "github.com/alibabacloud-go/tablestore-20201209/client"
	"github.com/alibabacloud-go/tea/tea"
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

// convertInstanceToCreateRequest converts TablestoreInstance to SDK create request
func convertInstanceToCreateRequest(instance *aliyunTablestoreAPI.TablestoreInstance) *consoleSDK.CreateInstanceRequest {
	request := &consoleSDK.CreateInstanceRequest{
		InstanceName: tea.String(instance.InstanceName),
		ClusterType:  tea.String(instance.ClusterType),
	}

	if instance.InstanceDescription != "" {
		request.InstanceDescription = tea.String(instance.InstanceDescription)
	}
	if instance.Network != "" {
		request.Network = tea.String(instance.Network)
	}
	if instance.Policy != "" {
		request.Policy = tea.String(instance.Policy)
	}
	if instance.ResourceGroupId != "" {
		request.ResourceGroupId = tea.String(instance.ResourceGroupId)
	}

	// Convert string arrays
	if len(instance.NetworkSourceACL) > 0 {
		networkSourceACL := make([]*string, len(instance.NetworkSourceACL))
		for i, str := range instance.NetworkSourceACL {
			networkSourceACL[i] = tea.String(str)
		}
		request.NetworkSourceACL = networkSourceACL
	}
	if len(instance.NetworkTypeACL) > 0 {
		networkTypeACL := make([]*string, len(instance.NetworkTypeACL))
		for i, str := range instance.NetworkTypeACL {
			networkTypeACL[i] = tea.String(str)
		}
		request.NetworkTypeACL = networkTypeACL
	}

	// Convert tags
	if len(instance.Tags) > 0 {
		tags := make([]*consoleSDK.CreateInstanceRequestTags, len(instance.Tags))
		for i, tag := range instance.Tags {
			tags[i] = &consoleSDK.CreateInstanceRequestTags{
				Key:   tea.String(tag.Key),
				Value: tea.String(tag.Value),
			}
		}
		request.Tags = tags
	}

	return request
}

// convertInstanceToUpdateRequest converts TablestoreInstance to SDK update request
func convertInstanceToUpdateRequest(instance *aliyunTablestoreAPI.TablestoreInstance) *consoleSDK.UpdateInstanceRequest {
	request := &consoleSDK.UpdateInstanceRequest{
		InstanceName: tea.String(instance.InstanceName),
	}

	if instance.AliasName != "" {
		request.AliasName = tea.String(instance.AliasName)
	}
	if instance.InstanceDescription != "" {
		request.InstanceDescription = tea.String(instance.InstanceDescription)
	}
	if instance.Network != "" {
		request.Network = tea.String(instance.Network)
	}

	// Convert string arrays
	if len(instance.NetworkSourceACL) > 0 {
		networkSourceACL := make([]*string, len(instance.NetworkSourceACL))
		for i, str := range instance.NetworkSourceACL {
			networkSourceACL[i] = tea.String(str)
		}
		request.NetworkSourceACL = networkSourceACL
	}
	if len(instance.NetworkTypeACL) > 0 {
		networkTypeACL := make([]*string, len(instance.NetworkTypeACL))
		for i, str := range instance.NetworkTypeACL {
			networkTypeACL[i] = tea.String(str)
		}
		request.NetworkTypeACL = networkTypeACL
	}

	return request
}

// convertTagsToSDK converts business tags to SDK format
func convertTagsToSDK(tags []aliyunTablestoreAPI.TablestoreInstanceTag) []*consoleSDK.TagResourcesRequestTags {
	if tags == nil {
		return nil
	}

	result := make([]*consoleSDK.TagResourcesRequestTags, 0, len(tags))
	for _, tag := range tags {
		result = append(result, &consoleSDK.TagResourcesRequestTags{
			Key:   tea.String(tag.Key),
			Value: tea.String(tag.Value),
		})
	}

	return result
}

// convertSchemaToTablestoreInstance converts Terraform schema data to TablestoreInstance
func convertSchemaToTablestoreInstance(d *schema.ResourceData) *aliyunTablestoreAPI.TablestoreInstance {
	instance := &aliyunTablestoreAPI.TablestoreInstance{
		InstanceName: d.Get("name").(string),
		ClusterType:  d.Get("cluster_type").(string),
	}

	if v, ok := d.GetOk("alias_name"); ok {
		instance.AliasName = v.(string)
	}
	if v, ok := d.GetOk("description"); ok {
		instance.InstanceDescription = v.(string)
	}
	if v, ok := d.GetOk("network"); ok {
		instance.Network = v.(string)
	}
	if v, ok := d.GetOk("policy"); ok {
		instance.Policy = v.(string)
	}
	if v, ok := d.GetOk("resource_group_id"); ok {
		instance.ResourceGroupId = v.(string)
	}

	// Convert network ACLs
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
func convertTablestoreInstanceToSchema(d *schema.ResourceData, instance *aliyunTablestoreAPI.TablestoreInstance) error {
	d.Set("name", instance.InstanceName)
	d.Set("alias_name", instance.AliasName)
	d.Set("description", instance.InstanceDescription)
	d.Set("cluster_type", instance.ClusterType)
	d.Set("storage_type", instance.StorageType)
	d.Set("status", instance.InstanceStatus)
	d.Set("network", instance.Network)
	d.Set("region_id", instance.RegionId)
	d.Set("resource_group_id", instance.ResourceGroupId)
	d.Set("payment_type", instance.PaymentType)
	d.Set("policy", instance.Policy)
	d.Set("policy_version", instance.PolicyVersion)
	d.Set("is_multi_az", instance.IsMultiAZ)
	d.Set("table_quota", instance.TableQuota)
	d.Set("vcu_quota", instance.VCUQuota)
	d.Set("elastic_vcu_upper_limit", instance.ElasticVCUUpperLimit)
	d.Set("create_time", instance.CreateTime.Format("2006-01-02T15:04:05Z"))
	d.Set("user_id", instance.UserId)

	// Set network ACLs
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
