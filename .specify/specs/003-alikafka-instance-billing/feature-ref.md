# Feature Reference: AliKafka Resource Layering

- **Feature ID**: 009
- **Feature Detail**: [.specify/memory/features/009.md](.specify/memory/features/009.md)
- **Spec**: [.specify/specs/003-alikafka-instance-billing/requirements.md](.specify/specs/003-alikafka-instance-billing/requirements.md)
- **Plan**: [.specify/specs/003-alikafka-instance-billing/plan.md](.specify/specs/003-alikafka-instance-billing/plan.md)

## Scope Alignment

本计划覆盖 AliKafka 实例创建的计费组合一致性，属于 Feature 009 的实例生命周期一致性范围。

## Resource / Service / Contract Mapping

| Resource | Resource File | Service File(s) | Contract Path |
|---|---|---|---|
| alicloud_alikafka_instance | alicloud/resource_alicloud_alikafka_instance.go | alicloud/service_alicloud_alikafka_instance.go | /alikafka/instances |
