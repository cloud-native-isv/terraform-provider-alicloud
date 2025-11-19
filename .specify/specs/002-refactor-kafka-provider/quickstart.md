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