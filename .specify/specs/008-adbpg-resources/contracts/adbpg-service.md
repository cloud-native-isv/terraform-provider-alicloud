# Contract: AdbpgService

**Spec**: [requirements.md](../requirements.md) | **Data Model**: [data-model.md](../data-model.md)

This document defines the service layer contract for `AdbpgService`. All methods listed here MUST be implemented in `alicloud/service_alicloud_adbpg.go`. Resource and data source layers MUST call these methods exclusively — no direct API or SDK calls.

## Service Struct

```go
type AdbpgService struct {
    client   *connectivity.AliyunClient
    adbpgAPI *adbpg.AdbpgAPI
}
```

## Constructor

### C-1: NewAdbpgService

```go
func NewAdbpgService(client *connectivity.AliyunClient) (*AdbpgService, error)
```

- MUST convert `client.AccessKey`, `client.SecretKey`, `client.RegionId`, `client.SecurityToken` to `common.Credentials`
- MUST call `adbpg.NewAdbpgAPI(credentials)` to create the API client
- MUST return wrapped error on failure

## Instance Methods

### C-2: DescribeAdbpgInstance

```go
func (s *AdbpgService) DescribeAdbpgInstance(instanceId string) (*adbpg.AdbpgInstanceDetail, error)
```

- Delegates to `s.adbpgAPI.DescribeInstance(instanceId)`
- Returns `IsNotFoundError` when instance does not exist

### C-3: CreateAdbpgInstance

```go
func (s *AdbpgService) CreateAdbpgInstance(input *adbpg.AdbpgInstanceCreate) (*adbpg.AdbpgInstanceDetail, error)
```

- Delegates to `s.adbpgAPI.CreateInstance(input)`
- Returns the created instance detail (API already calls DescribeInstance internally)

### C-4: DeleteAdbpgInstance

```go
func (s *AdbpgService) DeleteAdbpgInstance(instanceId string) error
```

- Delegates to `s.adbpgAPI.DeleteInstance(instanceId)`

### C-5: ModifyAdbpgInstanceDescription

```go
func (s *AdbpgService) ModifyAdbpgInstanceDescription(instanceId, description string) error
```

- Delegates to `s.adbpgAPI.ModifyInstanceDescription(instanceId, description)`

### C-6: ModifyAdbpgInstanceMaintainTime

```go
func (s *AdbpgService) ModifyAdbpgInstanceMaintainTime(instanceId, startTime, endTime string) error
```

- Delegates to `s.adbpgAPI.ModifyInstanceMaintainTime(instanceId, startTime, endTime)`

### C-7: AdbpgInstanceStateRefreshFunc

```go
func (s *AdbpgService) AdbpgInstanceStateRefreshFunc(instanceId string, failStates []string) resource.StateRefreshFunc
```

- Returns a `StateRefreshFunc` that calls `DescribeAdbpgInstance`
- Returns `(nil, "", nil)` when `IsNotFoundError`
- Returns `(object, status, WrapError)` when status matches a failState
- Returns `(object, status, nil)` otherwise

### C-8: WaitForAdbpgInstanceRunning

```go
func (s *AdbpgService) WaitForAdbpgInstanceRunning(instanceId string, timeout time.Duration) error
```

- Pending: `["Creating", "ClassChanging", "NetAddressCreating", "Restarting"]`
- Target: `["Running"]`
- FailStates: `[]` (locked state handled by checking LockMode in Read)
- Delay: 30 seconds
- MinTimeout: 10 seconds

### C-9: WaitForAdbpgInstanceDeleted

```go
func (s *AdbpgService) WaitForAdbpgInstanceDeleted(instanceId string, timeout time.Duration) error
```

- Pending: `["Deleting"]`
- Target: `[]` (empty — wait for resource to disappear)
- Delay: 30 seconds

## Account Methods

### C-10: ListAdbpgAccounts

```go
func (s *AdbpgService) ListAdbpgAccounts(instanceId string) ([]adbpg.AdbpgAccount, error)
```

- Delegates to `s.adbpgAPI.ListAccounts(instanceId)`

### C-11: DescribeAdbpgAccount

```go
func (s *AdbpgService) DescribeAdbpgAccount(id string) (*adbpg.AdbpgAccount, error)
```

- Decodes composite ID `instanceId:accountName`
- Calls `ListAdbpgAccounts(instanceId)` and finds the matching account by name
- Returns `IsNotFoundError` when account not found in list

### C-12: CreateAdbpgAccount

```go
func (s *AdbpgService) CreateAdbpgAccount(instanceId string, input *adbpg.AdbpgAccountCreate) error
```

- Delegates to `s.adbpgAPI.CreateAccount(instanceId, input)`

### C-13: DeleteAdbpgAccount

```go
func (s *AdbpgService) DeleteAdbpgAccount(instanceId, accountName string) error
```

- Delegates to `s.adbpgAPI.DeleteAccount(instanceId, accountName)`

### C-14: ResetAdbpgAccountPassword

```go
func (s *AdbpgService) ResetAdbpgAccountPassword(instanceId, accountName, newPassword string) error
```

- Delegates to `s.adbpgAPI.ResetAccountPassword(instanceId, accountName, newPassword)`

## Database Methods

### C-15: CreateAdbpgDatabase

```go
func (s *AdbpgService) CreateAdbpgDatabase(instanceId string, input *adbpg.AdbpgDatabaseCreate) error
```

- Delegates to `s.adbpgAPI.CreateDatabase(instanceId, input)`

### C-16: DeleteAdbpgDatabase

```go
func (s *AdbpgService) DeleteAdbpgDatabase(instanceId, dbName string) error
```

- Delegates to `s.adbpgAPI.DeleteDatabase(instanceId, dbName)`

### C-17: ListAdbpgDatabases

```go
func (s *AdbpgService) ListAdbpgDatabases(instanceId string) ([]adbpg.AdbpgDatabase, error)
```

- Delegates to `s.adbpgAPI.ListDatabases(instanceId)`

### C-18: DescribeAdbpgDatabase

```go
func (s *AdbpgService) DescribeAdbpgDatabase(id string) (*adbpg.AdbpgDatabase, error)
```

- Decodes composite ID `instanceId:dbName`
- Calls `ListAdbpgDatabases(instanceId)` and finds matching database by name
- Returns `IsNotFoundError` when not found

## Connection Methods

### C-19: AllocateAdbpgPublicConnection

```go
func (s *AdbpgService) AllocateAdbpgPublicConnection(instanceId, connectionStringPrefix string) error
```

- Delegates to `s.adbpgAPI.AllocatePublicConnection(instanceId, connectionStringPrefix)`

### C-20: ReleaseAdbpgPublicConnection

```go
func (s *AdbpgService) ReleaseAdbpgPublicConnection(instanceId, connectionString string) error
```

- Delegates to `s.adbpgAPI.ReleasePublicConnection(instanceId, connectionString)`

### C-21: DescribeAdbpgPublicConnection

```go
func (s *AdbpgService) DescribeAdbpgPublicConnection(instanceId string) (*adbpg.AdbpgNetInfo, error)
```

- Calls `s.adbpgAPI.DescribeInstanceNetInfo(instanceId)`
- Filters for entry where `IPType == "Public"`
- Returns `IsNotFoundError` when no public connection exists

## Security IP Methods

### C-22: ModifyAdbpgSecurityIps

```go
func (s *AdbpgService) ModifyAdbpgSecurityIps(instanceId, securityIPList, arrayName string) error
```

- Delegates to `s.adbpgAPI.ModifySecurityIps(instanceId, securityIPList, arrayName)`

## Backup Policy Methods

### C-23: DescribeAdbpgBackupPolicy

```go
func (s *AdbpgService) DescribeAdbpgBackupPolicy(instanceId string) (*adbpg.AdbpgBackupPolicy, error)
```

- Delegates to `s.adbpgAPI.DescribeBackupPolicy(instanceId)`

## SSL Methods

### C-24: DescribeAdbpgSSL

```go
func (s *AdbpgService) DescribeAdbpgSSL(instanceId string) (*adbpg.AdbpgSSLConfig, error)
```

- Delegates to `s.adbpgAPI.DescribeInstanceSSL(instanceId)`

### C-25: ModifyAdbpgSSL

```go
func (s *AdbpgService) ModifyAdbpgSSL(instanceId string, sslEnabled bool) error
```

- Delegates to `s.adbpgAPI.ModifyInstanceSSL(instanceId, sslEnabled)`

## Discovery Methods

### C-26: ListAdbpgAvailableResources

```go
func (s *AdbpgService) ListAdbpgAvailableResources(regionId string) ([]adbpg.AdbpgAvailableResource, error)
```

- Delegates to `s.adbpgAPI.ListAvailableResources(regionId)`

### C-27: ListAdbpgInstances

```go
func (s *AdbpgService) ListAdbpgInstances(query *adbpg.AdbpgInstanceQuery) ([]adbpg.AdbpgInstance, error)
```

- Delegates to `s.adbpgAPI.ListInstances(query)`

## Tag Methods

### C-28: TagAdbpgResources

```go
func (s *AdbpgService) TagAdbpgResources(resourceId string, tags []adbpg.Tag) error
```

- Delegates to `s.adbpgAPI.TagResources(resourceId, "instance", tags)`

### C-29: UntagAdbpgResources

```go
func (s *AdbpgService) UntagAdbpgResources(resourceId string, tagKeys []string) error
```

- Delegates to `s.adbpgAPI.UntagResources(resourceId, "instance", tagKeys)`
