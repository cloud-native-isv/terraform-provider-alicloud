# Data Model: Alikafka Instance

## Enitity: KafkaInstance

Represents an Alibaba Cloud Alikafka (Message Queue for Apache Kafka) Instance.

### Fields

| Field Name | Type | Description | Source (API) | Terraform Attribute |
|---|---|---|---|---|
| InstanceId | String | Unique identifier | InstanceId | `id` |
| Name | String | Instance name | InstanceName | `name` |
| VpcId | String | VPC ID | VpcId | `vpc_id` |
| VSwitchId | String | VSwitch ID | VSwitchId | `vswitch_id` |
| ZoneId | String | Zone ID | ZoneId | `zone_id` |
| SecurityGroup | String | Security Group ID | SecurityGroup | `security_group` |
| EndPoint | String | Default Endpoint | EndPoint | `end_point` |
| DomainEndpoint | String | Domain Endpoint | DomainEndpoint | `domain_endpoint` |
| SslEndpoint | String | SSL Endpoint | SslEndPoint | `ssl_endpoint` |
| SslDomainEndpoint | String | SSL Domain Endpoint | SslDomainEndpoint | `ssl_domain_endpoint` |
| SaslDomainEndpoint | String | SASL Domain Endpoint | SaslDomainEndpoint | `sasl_domain_endpoint` |
| ServiceVersion | String | Instance Service Version | ServiceVersion | `service_version` |
| Config | String | Instance Configuration | Config | `config` |
| ResourceGroupId | String | Resource Group ID | ResourceGroupId | `resource_group_id` |
| Tags | Map[string]string | Instance Tags | Tags | `tags` |

### Relationships

- Belongs to a Region.
- Belongs to a Resource Group.
- Associated with a VPC and VSwitch.

### State Transitions

- **Creating** -> **Running** (Available)
- **Running** -> **Deleting** -> **Deleted**

## Validation Rules

- `DeployType` must be 4 or 5.
- `DiskType` must be "0" or "1".
- `PaidType` must be "PrePaid" or "PostPaid".
