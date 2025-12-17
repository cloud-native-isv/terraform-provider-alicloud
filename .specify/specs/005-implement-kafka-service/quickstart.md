# Quickstart

## Using the new KafkaService methods

### Initialization

```go
client := meta.(*connectivity.AliyunClient)
kafkaService, err := NewKafkaService(client)
if err != nil {
    return WrapError(err)
}
```

### Creating an Order

```go
order := &kafka.KafkaOrder{
    RegionId:     client.RegionId,
    DeployType:   kafka.KafkaDeployTypeVPC,
    DiskSize:     500,
    DiskType:     kafka.KafkaDiskTypeUltra,
    SpecType:     kafka.KafkaSpecTypeProfessional,
    PartitionNum: 100,
    PaidType:     kafka.KafkaPaidTypePostPaid,
    // ... other fields
}

orderId, err := kafkaService.CreatePostPayOrder(order)
if err != nil {
    return err
}
```

### Starting an Instance

```go
req := &StartInstanceRequest{
    InstanceId: instanceId,
    RegionId:   client.RegionId,
    VpcId:      vpcId,
    VSwitchId:  vswitchId,
    // ... other fields
}

err := kafkaService.StartInstance(req)
if err != nil {
    return err
}
```
