package alicloud

import (
	"fmt"
	"time"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"

	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/kafka"
)

func validateAliKafkaBillingCombination(instanceType kafka.KafkaInstanceSeries, billingType kafka.KafkaBillingType) error {
	if instanceType == kafka.InstanceSeriesServerless && billingType == kafka.BillingTypePrePay {
		return fmt.Errorf("unsupported instance/billing combination: %s + %s; allowed combinations: Reserved + PrePaid, Reserved + PostPaid, Serverless + PostPaid", instanceType, billingType)
	}
	return nil
}

func buildAliKafkaInstanceCreationConfig(instance *kafka.KafkaInstance, instanceType kafka.KafkaInstanceSeries, billingType kafka.KafkaBillingType) kafka.InstanceCreationConfig {
	config := kafka.InstanceCreationConfig{
		RegionId:        instance.RegionId,
		ResourceGroupId: instance.ResourceGroupId,
		Tags:            instance.Tags,
		InstanceType:    instanceType,
		BillingType:     billingType,
	}

	if instance.Name != nil {
		config.Name = *instance.Name
	}
	if instance.Description != nil {
		config.Description = *instance.Description
	}
	if instance.ZoneId != "" {
		config.ZoneId = instance.ZoneId
	}
	if instance.VpcId != "" {
		config.VpcId = instance.VpcId
	}
	if instance.VSwitchId != "" {
		config.VSwitchId = instance.VSwitchId
	}
	if instance.SpecType != nil {
		config.SpecType = *instance.SpecType
	}
	if instance.DiskType != nil {
		config.DiskType = fmt.Sprintf("%d", *instance.DiskType)
	}
	if instance.DiskSize != nil {
		config.DiskSize = *instance.DiskSize
	}
	if instance.PartitionNum != nil {
		config.PartitionNum = *instance.PartitionNum
	}
	if instance.IoMaxSpec != nil {
		config.IoMaxSpec = *instance.IoMaxSpec
	}
	if instance.DeployType != nil {
		config.DeployType = int(*instance.DeployType)
	}
	if instance.EipMax != nil {
		config.EipMax = *instance.EipMax
	}
	if instance.Duration != nil {
		config.Duration = *instance.Duration
	}

	return config
}

// WaitForAliKafkaInstanceCreating waits for the Kafka instance to reach pending after create.
func (s *KafkaService) WaitForAliKafkaInstanceCreating(id string, timeout time.Duration) error {
	stateConf := BuildStateConf(
		[]string{
			fmt.Sprint(kafka.KafkaViewInstanceStatusCreated),
		}, // pending states during create
		[]string{fmt.Sprint(kafka.KafkaViewInstanceStatusCreated)}, // target state: PendingDeploy
		timeout,
		5*time.Second,
		s.AliKafkaInstancePropertyRefreshFunc(id, "view_instance_status_code"),
	)

	_, err := stateConf.WaitForState()
	if err == nil {
		return nil
	}
	return WrapErrorf(err, IdMsg, id)
}

// WaitForAliKafkaInstanceUpdating waits for the Kafka instance to complete an update operation
func (s *KafkaService) WaitForAliKafkaInstanceUpdating(id string, timeout time.Duration) error {
	stateConf := BuildStateConf(
		[]string{
			fmt.Sprint(kafka.KafkaViewInstanceStatusDeploying),
			fmt.Sprint(kafka.KafkaViewInstanceStatusStarting),
			fmt.Sprint(kafka.KafkaViewInstanceStatusUpgrading),
			fmt.Sprint(kafka.KafkaViewInstanceStatusMigrating),
			fmt.Sprint(kafka.KafkaViewInstanceStatusAutoScaling),
		}, // pending states during update
		[]string{fmt.Sprint(kafka.KafkaViewInstanceStatusRunning)},
		timeout,
		5*time.Second,
		s.AliKafkaInstancePropertyRefreshFunc(id, "view_instance_status_code"),
	)

	_, err := stateConf.WaitForState()
	if err == nil {
		return nil
	}
	return WrapErrorf(err, IdMsg, id)
}

// WaitForAliKafkaInstanceStopping waits for the Kafka instance to be stopped (state that indicates stopped)
func (s *KafkaService) WaitForAliKafkaInstanceStopping(id string, timeout time.Duration) error {
	stateConf := BuildStateConf(
		[]string{
			fmt.Sprint(kafka.KafkaViewInstanceStatusRunning),
			fmt.Sprint(kafka.KafkaViewInstanceStatusStopping),
		}, // pending states during stop
		[]string{
			fmt.Sprint(kafka.KafkaViewInstanceStatusStopped),
		}, // target state: Stopped
		timeout,
		5*time.Second,
		s.AliKafkaInstancePropertyRefreshFunc(id, "view_instance_status_code"),
	)

	_, err := stateConf.WaitForState()
	if err == nil {
		return nil
	}
	return WrapErrorf(err, IdMsg, id)
}

// WaitForAliKafkaInstanceStarting waits for the Kafka instance to reach running after start/deploy.
func (s *KafkaService) WaitForAliKafkaInstanceStarting(id string, timeout time.Duration) error {
	stateConf := BuildStateConf(
		[]string{
			fmt.Sprint(kafka.KafkaViewInstanceStatusCreated),
			fmt.Sprint(kafka.KafkaViewInstanceStatusDeploying),
			fmt.Sprint(kafka.KafkaViewInstanceStatusStarting),
			fmt.Sprint(kafka.KafkaViewInstanceStatusAutoScaling),
		}, // pending states during start
		[]string{fmt.Sprint(kafka.KafkaViewInstanceStatusRunning)},
		timeout,
		5*time.Second,
		s.AliKafkaInstancePropertyRefreshFunc(id, "view_instance_status_code"),
	)

	_, err := stateConf.WaitForState()
	if err == nil {
		return nil
	}
	return WrapErrorf(err, IdMsg, id)
}

// WaitForAliKafkaInstanceDeleting waits for the Kafka instance to be released after delete.
func (s *KafkaService) WaitForAliKafkaInstanceDeleting(id string, timeout time.Duration) error {
	stateConf := BuildStateConf(
		[]string{
			fmt.Sprint(kafka.KafkaViewInstanceStatusRunning),
			fmt.Sprint(kafka.KafkaViewInstanceStatusStopping),
			fmt.Sprint(kafka.KafkaViewInstanceStatusReleasing),
			fmt.Sprint(kafka.KafkaViewInstanceStatusStopped),
		}, // pending states during delete
		[]string{},
		timeout,
		5*time.Second,
		s.AliKafkaInstancePropertyRefreshFunc(id, "view_instance_status_code"),
	)

	_, err := stateConf.WaitForState()
	if err == nil {
		return nil
	}
	return WrapErrorf(err, IdMsg, id)
}

// CreatePostPayOrder creates a post-paid Kafka instance order using cws-lib-go API
func (s *KafkaService) CreatePostPayOrder(order *kafka.KafkaOrder) (string, error) {
	order.PaidType = kafka.KafkaPaidTypePostPay
	return s.kafkaApi.CreateOrder(order)
}

// CreatePrePayOrder creates a pre-paid Kafka instance order using cws-lib-go API
func (s *KafkaService) CreatePrePayOrder(order *kafka.KafkaOrder) (string, error) {
	order.PaidType = kafka.KafkaPaidTypePrePay
	return s.kafkaApi.CreateOrder(order)
}

// StartAlikafkaInstance 启动Kafka实例
func (s *KafkaService) StartAlikafkaInstance(request *StartInstanceRequest) error {
	var isEipInner *bool
	if request.IsEipInner {
		isEipInner = &request.IsEipInner
	}
	var isSetUserAndPassword *bool
	if request.IsSetUserAndPassword {
		isSetUserAndPassword = &request.IsSetUserAndPassword
	}
	var crossZone *bool
	if request.CrossZone {
		crossZone = &request.CrossZone
	}
	var isForceSelectedZones *bool
	if request.IsForceSelectedZones {
		isForceSelectedZones = &request.IsForceSelectedZones
	}

	config := kafka.StartInstanceConfig{
		InstanceId:           request.InstanceId,
		RegionId:             request.RegionId,
		VpcId:                request.VpcId,
		VSwitchId:            request.VSwitchId,
		ZoneId:               request.ZoneId,
		DeployModule:         request.DeployModule,
		IsEipInner:           isEipInner,
		IsSetUserAndPassword: isSetUserAndPassword,
		Username:             request.Username,
		Password:             request.Password,
		Name:                 request.Name,
		CrossZone:            crossZone,
		SecurityGroup:        request.SecurityGroup,
		ServiceVersion:       request.ServiceVersion,
		Config:               request.Config,
		KMSKeyId:             request.KMSKeyId,
		Notifier:             request.Notifier,
		UserPhoneNum:         request.UserPhoneNum,
		SelectedZones:        request.SelectedZones,
		IsForceSelectedZones: isForceSelectedZones,
		VSwitchIds:           request.VSwitchIds,
	}
	return s.kafkaApi.StartInstance(config)
}

// StopAlikafkaInstance stops a Kafka instance
func (s *KafkaService) StopAlikafkaInstance(request *StopInstanceRequest) error {
	return s.kafkaApi.StopInstance(request.InstanceId, request.RegionId)
}

// ModifyAlikafkaInstanceName 修改Kafka实例名称
func (s *KafkaService) ModifyAlikafkaInstanceName(request *ModifyInstanceNameRequest) error {
	return s.kafkaApi.ModifyInstanceName(request.InstanceId, request.RegionId, request.InstanceName)
}

// UpgradeInstanceVersion 升级Kafka实例版本
func (s *KafkaService) UpgradeInstanceVersion(request *UpgradeInstanceVersionRequest) error {
	return s.kafkaApi.UpgradeInstanceVersion(request.RegionId, request.InstanceId, request.TargetVersion)
}

// UpgradePostPayOrder upgrades a post-paid Kafka instance order using cws-lib-go API
func (s *KafkaService) UpgradePostPayOrder(order *kafka.KafkaOrder) (string, error) {
	order.PaidType = kafka.KafkaPaidTypePostPay
	return s.kafkaApi.UpgradeOrder(order)
}

// UpgradePrePayOrder upgrades a pre-paid Kafka instance order using cws-lib-go API
func (s *KafkaService) UpgradePrePayOrder(order *kafka.KafkaOrder) (string, error) {
	order.PaidType = kafka.KafkaPaidTypePrePay
	return s.kafkaApi.UpgradeOrder(order)
}

// UpdateInstanceConfig updates the configuration of a Kafka instance
func (s *KafkaService) UpdateInstanceConfig(instanceId string, config map[string]*string) error {
	return s.kafkaApi.UpdateInstanceConfig(instanceId, s.client.RegionId, config)
}

// UpdateAlikafkaInstance updates a Kafka instance
func (s *KafkaService) UpdateAlikafkaInstance(instance *kafka.KafkaInstance) error {
	return s.kafkaApi.UpdateInstance(instance)
}

func (s *KafkaService) AliKafkaInstanceStateRefreshFunc(id string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeAlikafkaInstance(id)
		if err != nil {
			if NotFoundError(err) {
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		status := fmt.Sprint(object.ViewInstanceStatusCode)

		// Check if the current status is a fail state
		for _, failState := range failStates {
			if status == failState {
				return object, status, fmt.Errorf("resource in failed state: %s", status)
			}
		}

		return object, status, nil
	}
}

func (s *KafkaService) AliKafkaInstancePropertyRefreshFunc(id string, property string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeAlikafkaInstance(id)
		if err != nil {
			if NotFoundError(err) {
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		var val interface{}
		switch property {
		case "disk_size":
			val = tea.IntValue(object.DiskSize)
		case "eip_max":
			val = tea.IntValue(object.EipMax)
		case "spec_type":
			val = tea.StringValue(object.SpecType)
		case "view_instance_status_code":
			val = object.ViewInstanceStatusCode
		}

		return object, fmt.Sprint(val), nil
	}
}

// DescribeAlikafkaInstance retrieves a Kafka instance using CWS-Lib-Go
func (s *KafkaService) DescribeAlikafkaInstance(instanceId string) (*kafka.KafkaInstance, error) {
	instance, err := s.kafkaApi.GetInstance(instanceId)
	if err != nil {
		if IsExpectedErrors(err, []string{"instance not found", "instance not found in list"}) {
			return nil, WrapErrorf(NotFoundErr("AlikafkaInstance", instanceId), NotFoundMsg, ProviderERROR)
		}
		return nil, WrapError(err)
	}

	if instance.ViewInstanceStatusCode == kafka.KafkaViewInstanceStatusReleased {
		return nil, WrapErrorf(NotFoundErr("AlikafkaInstance", instanceId), NotFoundMsg, ProviderERROR)
	}

	return instance, nil
}

func (s *KafkaService) CreateAlikafkaInstance(config kafka.InstanceCreationConfig) (*kafka.InstanceCreationResult, error) {
	result, err := s.kafkaApi.CreateInstance(config)
	if err != nil {
		return nil, WrapError(err)
	}
	return result, nil
}

// UpgradeAlikafkaInstance upgrades a Kafka instance using CWS-Lib-Go
func (s *KafkaService) UpgradeAlikafkaInstance(instance *kafka.KafkaInstance) error {
	return s.kafkaApi.UpgradeInstance(instance)
}

// ListAlikafkaInstances lists Kafka instances using CWS-Lib-Go
func (s *KafkaService) ListAlikafkaInstances(regionId string) ([]*kafka.KafkaInstance, error) {
	instances, err := s.kafkaApi.ListInstances(regionId)
	if err != nil {
		return nil, WrapError(err)
	}
	return instances, nil
}

// DeleteAlikafkaInstance deletes a Kafka instance using CWS-Lib-Go
func (s *KafkaService) DeleteAlikafkaInstance(instanceId string) error {
	if err := s.kafkaApi.DeleteInstance(instanceId); err != nil {
		return WrapError(err)
	}
	return nil
}
