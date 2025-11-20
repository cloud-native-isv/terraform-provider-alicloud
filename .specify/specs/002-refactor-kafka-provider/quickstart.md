# Quickstart: Kafka Provider Refactoring Implementation

## Overview

This quickstart guide provides the essential steps to implement the Kafka provider refactoring according to the approved specification and design.

## Implementation Steps

### 1. Service Layer Implementation

Create/Update `service_alicloud_alikafka.go` with the following structure:

```go
package alicloud

import (
	"time"
	
	aliyunKafkaAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/kafka"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

type KafkaService struct {
	client *connectivity.AliyunClient
}

// Instance methods
func (s *KafkaService) DescribeKafkaInstance(id string) (*aliyunKafkaAPI.Instance, error) {
	return s.GetAPI().GetKafkaInstance(id)
}

func (s *KafkaService) CreateKafkaInstance(request *aliyunKafkaAPI.CreateInstanceRequest) (*aliyunKafkaAPI.Instance, error) {
	return s.GetAPI().CreateKafkaInstance(request)
}

func (s *KafkaService) DeleteKafkaInstance(id string) error {
	return s.GetAPI().DeleteKafkaInstance(id)
}

func (s *KafkaService) KafkaInstanceStateRefreshFunc(id string, failStates []string) resource.StateRefreshFunc {
	// Implementation based on Flink pattern
}

func (s *KafkaService) WaitForKafkaInstanceCreating(id string, timeout time.Duration) error {
	// Implementation using BuildStateConf
}

func (s *KafkaService) WaitForKafkaInstanceDeleting(id string, timeout time.Duration) error {
	// Implementation using StateChangeConf
}

// Similar methods for Topic, ConsumerGroup, SaslUser, SaslAcl, AllowedIp
// Follow the same pattern as above
```

### 2. ID Encoding/Decoding Functions

Implement proper ID encoding/decoding functions:

```go
// EncodeConsumerGroupId creates composite ID for consumer group
func EncodeConsumerGroupId(instanceId, consumerId string) string {
	return fmt.Sprintf("%s:%s", instanceId, consumerId)
}

// DecodeConsumerGroupId parses composite ID into components
func DecodeConsumerGroupId(id string) (string, string, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid consumer group ID format, expected instanceId:consumerId, got %s", id)
	}
	return parts[0], parts[1], nil
}

// Similar functions for Topic, SaslUser, SaslAcl, AllowedIp
```

### 3. Resource Implementation Pattern

Each resource file should follow this pattern:

```go
func resourceAliCloudAlikafkaInstanceCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	kafkaService, err := NewKafkaService(client)
	if err != nil {
		return WrapError(err)
	}

	// Build request from Terraform schema
	request := &aliyunKafkaAPI.CreateInstanceRequest{
		RegionId: client.RegionId,
		Name:     d.Get("name").(string),
		// ... other fields
	}

	// Create with retry
	var instance *aliyunKafkaAPI.Instance
	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		resp, err := kafkaService.CreateKafkaInstance(request)
		if err != nil {
			if NeedRetry(err) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		instance = resp
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_alikafka_instance", "CreateKafkaInstance", AlibabaCloudSdkGoERROR)
	}

	d.SetId(instance.InstanceId)

	// Wait for creation to complete
	err = kafkaService.WaitForKafkaInstanceCreating(d.Id(), d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	// Final read to sync state
	return resourceAliCloudAlikafkaInstanceRead(d, meta)
}
```

### 4. Data Source Implementation Pattern

Data sources should follow this pattern:

```go
func dataSourceAliCloudAlikafkaInstancesRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	kafkaService, err := NewKafkaService(client)
	if err != nil {
		return WrapError(err)
	}

	// Get all instances (pagination handled in service layer)
	instances, err := kafkaService.ListKafkaInstances()
	if err != nil {
		return WrapError(err)
	}

	// Filter and map results
	var instanceMaps []map[string]interface{}
	ids := make([]string, 0)

	for _, instance := range instances {
		instanceMap := map[string]interface{}{
			"id": instance.InstanceId,
			"name": instance.Name,
			// ... other fields
		}
		instanceMaps = append(instanceMaps, instanceMap)
		ids = append(ids, instance.InstanceId)
	}

	d.SetId("alikafka_instances")
	d.Set("ids", ids)
	d.Set("instances", instanceMaps)

	return nil
}
```

### 5. Error Handling

Use standard error handling patterns:

```go
// In Read methods
object, err := kafkaService.DescribeKafkaInstance(d.Id())
if err != nil {
	if !d.IsNewResource() && IsNotFoundError(err) {
		log.Printf("[DEBUG] Resource alicloud_alikafka_instance DescribeKafkaInstance Failed!!! %s", err)
		d.SetId("")
		return nil
	}
	return WrapError(err)
}

// In Create/Delete methods
if IsNotFoundError(err) {
	// Handle appropriately for context
}
if IsAlreadyExistError(err) {
	// Handle appropriately for context  
}
if NeedRetry(err) {
	return resource.RetryableError(err)
}
```

### 6. Validation Checklist

Before submitting the implementation:

- [ ] All Resources/DataSources call Service layer only
- [ ] Service layer uses cws-lib-go API functions only  
- [ ] No `map[string]interface{}` usage in new code
- [ ] Proper ID encoding/decoding functions implemented
- [ ] State management uses WaitFor* functions with proper timeouts
- [ ] Error handling uses standard predicates and wrapping
- [ ] All existing acceptance tests pass
- [ ] `make` command compiles successfully
- [ ] Code files under 1000 lines (split if necessary)

## Key Files to Modify

1. `alicloud/service_alicloud_alikafka.go` - Service layer implementation
2. `alicloud/resource_alicloud_alikafka_instance.go` - Instance resource
3. `alicloud/resource_alicloud_alikafka_topic.go` - Topic resource  
4. `alicloud/resource_alicloud_alikafka_consumer_group.go` - Consumer group resource
5. `alicloud/resource_alicloud_alikafka_sasl_user.go` - SASL user resource
6. `alicloud/resource_alicloud_alikafka_sasl_acl.go` - SASL ACL resource
7. `alicloud/resource_alicloud_alikafka_instance_allowed_ip_attachment.go` - IP attachment resource
8. `alicloud/data_source_alicloud_alikafka_instances.go` - Instances data source
9. `alicloud/data_source_alicloud_alikafka_topics.go` - Topics data source
10. `alicloud/data_source_alicloud_alikafka_consumer_groups.go` - Consumer groups data source
11. `alicloud/data_source_alicloud_alikafka_sasl_users.go` - SASL users data source
12. `alicloud/data_source_alicloud_alikafka_sasl_acls.go` - SASL ACLs data source

## Testing Strategy

1. **Unit Tests**: Ensure all service layer methods have proper unit tests
2. **Acceptance Tests**: Run all existing Kafka acceptance tests to verify backward compatibility
3. **Manual Testing**: Test create/read/update/delete operations for each resource type
4. **Error Scenarios**: Test error handling for quota limits, invalid configurations, etc.
5. **Performance Testing**: Verify resource creation times are within 10% of baseline

## Common Pitfalls to Avoid

- **Direct SDK calls**: Never call alikafka SDK directly from Resources/DataSources
- **Weak typing**: Avoid `map[string]interface{}` - use cws-lib-go strong types
- **Missing state waits**: Always use WaitFor* functions after Create/Delete operations
- **Incorrect ID handling**: Use proper Encode/Decode functions for composite IDs
- **Inconsistent error handling**: Always use standard error predicates and wrapping
- **Pagination issues**: Handle pagination in service layer, not in Resources/DataSources

## Updated Implementation Guidance

### Service Layer Structure

Based on analysis of the existing code and cws-lib-go API, the service layer should be structured as follows:

1. **KafkaService struct** that encapsulates the client connection
2. **GetAPI() method** that returns the cws-lib-go Kafka API client
3. **Describe* methods** for each resource type that retrieve resource details
4. **Create* methods** for each resource type that create new resources
5. **Delete* methods** for each resource type that delete resources
6. **State refresh functions** that implement the resource.StateRefreshFunc interface
7. **WaitFor* methods** that use BuildStateConf to wait for resource state changes

### Resource Implementation Details

Each resource should follow these specific patterns:

#### Create Method
1. Initialize the KafkaService
2. Build the request object using cws-lib-go types
3. Use resource.Retry with proper error handling
4. Set the resource ID from the response
5. Wait for the resource to be in the desired state
6. Call the Read method to sync the state

#### Read Method
1. Initialize the KafkaService
2. Call the appropriate Describe method
3. Handle IsNotFoundError for non-new resources
4. Set all schema fields including computed properties
5. Return proper error wrapping

#### Delete Method
1. Initialize the KafkaService
2. Call the appropriate Delete method
3. Handle IsNotFoundError as successful deletion
4. Use StateChangeConf to wait for actual deletion completion
5. Proper timeout and delay configuration

### Data Source Implementation Details

Data sources should:
1. Initialize the KafkaService
2. Call the appropriate List method
3. Filter results based on input parameters
4. Map the cws-lib-go types to Terraform schema
5. Set the appropriate schema fields

### Error Handling Best Practices

1. **Use WrapError/WrapErrorf** for all error wrapping
2. **Use error predicates** (IsNotFoundError, IsAlreadyExistError, NeedRetry) rather than IsExpectedErrors
3. **Include context** in error messages for easier troubleshooting
4. **Handle retryable errors** appropriately with resource.RetryableError
5. **Log debug information** when appropriate

### ID Management Best Practices

1. **Use existing Encode*/Decode* functions** consistently
2. **Follow the established patterns** for composite IDs
3. **Ensure backward compatibility** with existing resource IDs
4. **Validate ID formats** in Decode functions

### State Management Best Practices

1. **Use WaitFor* functions** for async operations
2. **Don't call Read directly in Create** methods
3. **Implement proper timeout configuration**
4. **Use BuildStateConf** with appropriate pending/target states
5. **Handle fail states** appropriately in state refresh functions

## Code Examples

### Service Layer Method Implementation
```go
// DescribeKafkaInstance gets Kafka instance details
func (s *KafkaService) DescribeKafkaInstance(id string) (*aliyunKafkaAPI.KafkaInstance, error) {
	kafkaInstance, err := s.GetAPI().GetInstance(context.TODO(), id)
	if err != nil {
		if IsNotFoundError(err) {
			return nil, WrapErrorf(err, NotFoundMsg, id)
		}
		return nil, WrapErrorf(err, DefaultErrorMsg, id, "DescribeKafkaInstance", AlibabaCloudSdkGoERROR)
	}

	// ServiceStatus equals 10 means the instance is released, don't return the instance
	if kafkaInstance.Status == "10" {
		return nil, WrapErrorf(NotFoundErr("KafkaInstance", id), NotFoundMsg, ProviderERROR)
	}

	return kafkaInstance, nil
}
```

### Resource Create Method Implementation
```go
func resourceAliCloudKafkaInstanceCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	kafkaService, err := NewKafkaService(client)
	if err != nil {
		return WrapError(err)
	}

	// Create instance request using cws-lib-go types
	instanceRequest := &aliyunKafkaAPI.KafkaInstance{
		Name:            d.Get("name").(string),
		RegionId:        client.RegionId,
		ZoneId:          d.Get("zone_id").(string),
		DiskType:        d.Get("disk_type").(string),
		DiskSize:        int32(d.Get("disk_size").(int)),
		DeployType:      int32(d.Get("deploy_type").(int)),
		IoMax:           int32(d.Get("io_max").(int)),
		SpecType:        d.Get("spec_type").(string),
		Version:         d.Get("version").(string),
		VpcId:           d.Get("vpc_id").(string),
		VSwitchId:       d.Get("vswitch_id").(string),
		SecurityGroupId: d.Get("security_group_id").(string),
	}

	// Create the instance with retry mechanism
	var instance *aliyunKafkaAPI.KafkaInstance
	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		resp, err := kafkaService.CreateKafkaInstance(instanceRequest)
		if err != nil {
			if NeedRetry(err) {
				time.Sleep(5 * time.Second)
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		instance = resp
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_kafka_instance", "CreateKafkaInstance", AlibabaCloudSdkGoERROR)
	}

	if instance == nil || instance.InstanceId == "" {
		return WrapError(Error("Failed to get instance ID from Kafka instance"))
	}

	d.SetId(instance.InstanceId)

	// Wait for the instance to be in running state using service layer function
	if err := kafkaService.WaitForKafkaInstanceCreating(d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	// Finally call Read to sync state
	return resourceAliCloudKafkaInstanceRead(d, meta)
}
```

### Data Source Read Method Implementation
```go
func dataSourceAliCloudKafkaInstancesRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	kafkaService, err := NewKafkaService(client)
	if err != nil {
		return WrapError(err)
	}

	// Get all Kafka instances (pagination handled in service layer)
	instances, err := kafkaService.ListKafkaInstances()
	if err != nil {
		return WrapError(err)
	}

	// Filter and map results
	var instanceMaps []map[string]interface{}
	ids := make([]string, 0)
	names := make([]string, 0)

	for _, instance := range instances {
		// Apply filters if specified
		if v, ok := d.GetOk("ids"); ok && len(v.([]interface{})) > 0 {
			found := false
			for _, id := range v.([]interface{}) {
				if instance.InstanceId == id.(string) {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		if v, ok := d.GetOk("names"); ok && len(v.([]interface{})) > 0 {
			found := false
			for _, name := range v.([]interface{}) {
				if instance.Name == name.(string) {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Map fields from cws-lib-go types to Terraform schema
		instanceMap := map[string]interface{}{
			"id":             instance.InstanceId,
			"name":           instance.Name,
			"status":         instance.Status,
			"region_id":      instance.RegionId,
			"zone_id":        instance.ZoneId,
			"spec_type":      instance.SpecType,
			"disk_type":      instance.DiskType,
			"disk_size":      instance.DiskSize,
			"io_max":         instance.IoMax,
			"io_max_spec":    instance.IoMaxSpec,
			"version":        instance.Version,
			"endpoint":       instance.EndPoint,
			"create_time":    instance.CreateTime,
			"expire_time":    instance.ExpireTime,
			"ssl_endpoint":   instance.SslEndPoint,
			"sasl_endpoint":  instance.SaslEndPoint,
		}

		instanceMaps = append(instanceMaps, instanceMap)
		ids = append(ids, instance.InstanceId)
		names = append(names, instance.Name)
	}

	// Set data source return values
	d.SetId("kafka_instances")
	if err := d.Set("ids", ids); err != nil {
		return WrapError(err)
	}
	if err := d.Set("names", names); err != nil {
		return WrapError(err)
	}
	if err := d.Set("instances", instanceMaps); err != nil {
		return WrapError(err)
	}

	return nil
}
```