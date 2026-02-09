# Feature Reference: AliKafka Resource Layering

- **Feature ID**: 009
- **Feature Detail**: [.specify/memory/features/009.md](.specify/memory/features/009.md)
- **Spec**: [.specify/specs/002-alikafka-service-layer/requirements.md](.specify/specs/002-alikafka-service-layer/requirements.md)
- **Plan**: [.specify/specs/002-alikafka-service-layer/plan.md](.specify/specs/002-alikafka-service-layer/plan.md)

## Scope Alignment

本计划覆盖 AliKafka 相关资源的服务层对齐与核心生命周期操作补齐，保持与 Feature 009 的范围一致。

## Resource / Service / Contract Mapping

| Resource | Resource File | Service File(s) | Contract Path |
|---|---|---|---|
| alicloud_alikafka_instance | alicloud/resource_alicloud_alikafka_instance.go | alicloud/service_alicloud_alikafka_instance.go | /alikafka/instances |
| alicloud_alikafka_topic | alicloud/resource_alicloud_alikafka_topic.go | alicloud/service_alicloud_alikafka_topic.go | /alikafka/instances/{instanceId}/topics |
| alicloud_alikafka_sasl_user | alicloud/resource_alicloud_alikafka_sasl_user.go | alicloud/service_alicloud_alikafka_sasl_user.go | /alikafka/instances/{instanceId}/sasl-users |
| alicloud_alikafka_sasl_acl | alicloud/resource_alicloud_alikafka_sasl_acl.go | alicloud/service_alicloud_alikafka_sasl_acl.go | /alikafka/instances/{instanceId}/sasl-acls |
| alicloud_alikafka_consumer_group | alicloud/resource_alicloud_alikafka_consumer_group.go | alicloud/service_alicloud_alikafka_consumer_group.go | /alikafka/instances/{instanceId}/consumer-groups |
| alicloud_alikafka_deployment | alicloud/resource_alicloud_alikafka_deployment.go | alicloud/service_alicloud_alikafka.go | /alikafka/instances/{instanceId}/deployments |
| alicloud_alikafka_instance_allowed_ip_attachment | alicloud/resource_alicloud_alikafka_instance_allowed_ip_attachment.go | alicloud/service_alicloud_alikafka.go | /alikafka/instances/{instanceId}/allowed-ips |
