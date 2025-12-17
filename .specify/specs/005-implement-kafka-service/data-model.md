# Data Model

## Request Structs

### StartInstanceRequest

```go
type StartInstanceRequest struct {
    InstanceId           string
    RegionId             string
    VpcId                string
    VSwitchId            string
    ZoneId               string
    DeployModule         string
    IsEipInner           bool
    IsSetUserAndPassword bool
    Username             string
    Password             string
    Name                 string
    CrossZone            bool
    SecurityGroup        string
    ServiceVersion       string
    Config               string
    KMSKeyId             string
    Notifier             string
    UserPhoneNum         string
    SelectedZones        string
    IsForceSelectedZones bool
    VSwitchIds           []string
}
```

### ModifyInstanceNameRequest

```go
type ModifyInstanceNameRequest struct {
    InstanceId   string
    RegionId     string
    InstanceName string
}
```

### UpgradeInstanceVersionRequest

```go
type UpgradeInstanceVersionRequest struct {
    InstanceId    string
    RegionId      string
    TargetVersion string
}
```

### KafkaOrder (from cws-lib-go)

Used for `CreatePostPayOrder` and `CreatePrePayOrder`.

```go
type KafkaOrder struct {
    RegionId        string
    DeployType      KafkaDeployType
    DiskSize        int32
    DiskType        KafkaDiskType
    IoMax           int32
    IoMaxSpec       string
    SpecType        KafkaSpecType
    PartitionNum    int32
    TopicQuota      int32
    EipMax          int32
    PaidType        KafkaPaidType
    Duration        int32
    ResourceGroupId string
    Tags            map[string]string
}
```
