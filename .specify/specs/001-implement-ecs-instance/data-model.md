# Data Model: alicloud_ecs_instance

## Terraform Resource Schema

| Field | Type | Required/Optional | Description | Notes |
|---|---|---|---|---|
| `image_id` | String | Required | Image ID to launch. | ForceNew |
| `instance_type` | String | Required | Instance Type (e.g., ecs.g6.large). | ForceNew |
| `security_groups` | List(String) | Required | List of Security Group IDs. | |
| `vswitch_id` | String | Required | VSwitch ID for the instance. | ForceNew |
| `instance_name` | String | Optional | Display name of the instance. | |
| `description` | String | Optional | Description of the instance. | |
| `tags` | Map(String) | Optional | Tags to assign. | |
| `status` | String | Computed | Current status (Running/Stopped). | |
| `public_ip` | String | Computed | Public IP address (if assigned). | |
| `private_ip` | String | Computed | Private IP address. | |

## Service Layer Interface

```go
// EcsService Interface

// CreateInstance creates a new ECS instance using CreateInstance API
func (s *EcsService) CreateInstance(request *ecs.CreateInstanceRequest) (*ecs.Instance, error)

// DescribeInstance retrieves ID, Status, and Network details
func (s *EcsService) DescribeInstance(instanceId string) (*ecs.Instance, error)

// DeleteInstance terminates an instance
func (s *EcsService) DeleteInstance(instanceId string) error

// WaitForInstanceRunning waits until status is Running
func (s *EcsService) WaitForInstanceRunning(instanceId string, timeout time.Duration) error

// WaitForNetworkReady waits until primary network interface is active
func (s *EcsService) WaitForNetworkReady(instanceId string, timeout time.Duration) error
```
