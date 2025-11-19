# Research Findings: Kafka Provider Refactoring

## Research Tasks

### Task 1: Research cws-lib-go Kafka API structure for Kafka provider refactoring
**Context**: Need to understand the exact API structure and available functions in cws-lib-go for Kafka resources to properly implement the service layer.

### Task 2: Research Kafka resource ID encoding patterns for provider implementation  
**Context**: Need to understand how IDs should be encoded/decoded for different Kafka resources (instances, topics, consumer groups, SASL users, SASL ACLs, IP attachments) to maintain consistency with existing patterns.

### Task 3: Research Kafka state transitions and valid states for proper state management
**Context**: Need to understand the valid states and transitions for Kafka resources to implement proper StateRefreshFunc and WaitFor functions.

## Findings

### Decision: cws-lib-go Kafka API Structure
**Rationale**: Based on the go.mod file analysis, the cws-lib-go provides Kafka API through `github.com/alibabacloud-go/alikafka-20190916/v3` which is mapped to `../cws-lib-go/lib/cloud/aliyun/sdk/alikafka-20190916`. The API follows the same pattern as the Flink API with strong typing and proper error handling. The service layer should use `s.GetAPI().GetKafkaClient()` or similar pattern to access the Kafka API client.

**Alternatives considered**: 
- Direct SDK usage (rejected - violates layering principle)
- HTTP API calls (rejected - violates layering and security principles)
- Third-party Kafka libraries (rejected - not aligned with provider architecture)

### Decision: Kafka Resource ID Encoding Patterns
**Rationale**: Based on the existing code analysis, Kafka resources use colon-separated composite IDs:
- Consumer Groups: `instanceId:consumerId`
- Topics: `instanceId:topicName` 
- SASL Users: `instanceId:username`
- SASL ACLs: `instanceId:username:resourceType:resourceName:patternType:operationType`
- IP Attachments: `instanceId:allowedType:portRange:ipAddress`

The service layer should implement `Encode*Id` and `Decode*Id` functions following the standard pattern established in the constitution.

**Alternatives considered**:
- Using simple instance IDs only (rejected - insufficient for multi-resource identification)
- JSON-encoded IDs (rejected - violates simplicity and readability principles)
- UUID-based IDs (rejected - breaks existing user configurations)

### Decision: Kafka State Transitions and Valid States
**Rationale**: Based on the existing implementation analysis:
- **Instances**: ServiceStatus 5 = running, ServiceStatus 10 = deleted/released
- **Topics**: Existence-based state (topic exists = ready, doesn't exist = deleted)
- **Consumer Groups**: Existence-based state
- **SASL Users**: Existence-based state  
- **SASL ACLs**: Existence-based state
- **IP Attachments**: Existence-based state

State management should use existence checks for most resources and ServiceStatus checks for instances. The WaitFor functions should use appropriate pending/target states based on these patterns.

**Alternatives considered**:
- Complex state machines for all resources (rejected - overcomplicated for most Kafka resources)
- Polling all resources with the same pattern (rejected - doesn't account for different resource characteristics)
- No state waiting (rejected - violates reliability requirements)

## Implementation Guidance

1. **Service Layer Structure**: Create a `KafkaService` struct with methods following the Flink pattern:
   - `Describe*` methods for reading resources
   - `Create*` methods for creating resources  
   - `Delete*` methods for deleting resources
   - `WaitFor*` methods for state management
   - `*StateRefreshFunc` methods for state polling

2. **API Client Access**: Use `s.GetAPI().GetKafkaClient()` or similar pattern to access the cws-lib-go Kafka client

3. **Error Handling**: Use standard error predicates (`IsNotFoundError`, `IsAlreadyExistError`, `NeedRetry`) with appropriate error wrapping

4. **ID Management**: Implement proper `Encode*Id`/`Decode*Id` functions for all composite resources

5. **State Management**: Use `BuildStateConf` with proper pending/target states and timeouts

6. **Pagination**: Encapsulate pagination logic in the service layer for list operations

All implementations must strictly follow the constitutional requirements for layering, error handling, strong typing, and build verification.