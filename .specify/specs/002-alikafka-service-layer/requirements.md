# Requirements Specification: AliKafka 资源分层改造

**Requirement Branch**: `002-alikafka-service-layer`  
**Created**: 2026-02-09  
**Status**: Draft  
**Input**: User description: "更新如下resource的实现,使用service_alicloud_alikafka*.go文件中封装的service方法，如果service层没有对应的方法需要在service层通过调用cws-lib-go提供的API方法实现对应的方法,先使用`ls alicloud/service_alicloud_alikafka*.go`命令获取所有的service层实现，重点在于补全而**不是替换**各个service中的实现，service实现之后需要将各个resource中的实现替换为调用service层的代码，通过这个实现完成“分层架构设计”：      "alicloud_alikafka_consumer_group":                               resourceAliCloudAlikafkaConsumerGroup(),
  "alicloud_alikafka_deployment":                                   resourceAliCloudAlikafkaDeployment(),
  "alicloud_alikafka_instance":                                     resourceAliCloudAlikafkaInstance(),
  "alicloud_alikafka_instance_allowed_ip_attachment":               resourceAliCloudAliKafkaInstanceAllowedIpAttachment(),
  "alicloud_alikafka_sasl_acl":                                     resourceAliCloudAlikafkaSaslAcl(),
  "alicloud_alikafka_sasl_user":                                    resourceAliCloudAlikafkaSaslUser(),
  "alicloud_alikafka_topic":                                        resourceAliCloudAlikafkaTopic(),"

## User Scenarios & Testing *(mandatory)*

<!--
  IMPORTANT: User stories should be PRIORITIZED as user journeys ordered by importance.
  Each user story/journey must be INDEPENDENTLY TESTABLE - meaning if you implement just ONE of them,
  you should still have a viable MVP (Minimum Viable Product) that delivers value.
  
  Assign priorities (P1, P2, P3, etc.) to each story, where P1 is the most critical.
  Think of each story as a standalone slice of functionality that can be:
  - Developed independently
  - Tested independently
  - Deployed independently
  - Demonstrated to users independently
-->

### User Story 1 - 统一分层架构 (Priority: P1)

作为维护者，我希望 AliKafka 相关资源的业务操作统一通过服务层完成，以确保分层架构一致并减少重复实现。

**Why this priority**: 这是确保架构一致性与后续维护成本可控的关键基础。

**Independent Test**: 可独立验证资源层是否仅调用服务层，并确保核心操作路径可用。

**Acceptance Scenarios**:

1. **Given** 已存在 7 个 AliKafka 资源实现，**When** 分层改造完成，**Then** 资源层仅通过服务层完成业务操作且无绕过路径。
2. **Given** 资源层调用路径已统一，**When** 执行资源的核心操作，**Then** 行为与改造前保持一致且状态可正确同步。

---

### User Story 2 - 服务层能力补齐 (Priority: P2)

作为贡献者，我希望服务层为 AliKafka 资源提供完整且可复用的操作能力，以便在不触及资源层细节的情况下扩展功能。

**Why this priority**: 服务层能力完备性决定了资源层改造的可执行性与复用性。

**Independent Test**: 可独立检查服务层是否覆盖资源所需的操作集合并可被调用。

**Acceptance Scenarios**:

1. **Given** 服务层已有部分能力，**When** 补齐缺失操作，**Then** 资源层无需额外逻辑即可完成对应操作。

---

### User Story 3 - 用户体验不回退 (Priority: P3)

作为使用者，我希望 AliKafka 资源的使用体验不因内部改造而回退，既有配置可以继续正常使用。

**Why this priority**: 避免对现有用户造成破坏性影响，降低升级成本。

**Independent Test**: 可独立验证既有配置在改造后可继续完成核心操作。

**Acceptance Scenarios**:

1. **Given** 既有的资源配置与状态，**When** 进行资源的创建/读取/删除流程，**Then** 不需要修改配置即可成功完成。

---

### Edge Cases

- 服务层缺失某个资源操作时，系统应能识别并阻止资源层继续绕过执行。
- 资源字段映射不一致时，系统应保证关键状态字段仍能正确同步。
- 某些操作仅在特定资源状态下可执行时，系统应避免触发无效操作。

## Assumptions

- 不新增或删除任何 AliKafka 资源类型，仅改造其内部调用路径。
- 外部接口与现有资源 Schema 保持兼容，不要求用户修改既有配置。
- 改造仅影响 AliKafka 相关资源，不扩展到其他服务。

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: 系统必须覆盖并改造以下 7 类 AliKafka 资源：消费者组、部署、实例、实例白名单绑定、SASL ACL、SASL 用户、主题。
- **FR-002**: 系统必须为上述资源提供可复用的服务层操作能力，覆盖其核心生命周期操作。
- **FR-003**: 资源层必须仅通过服务层完成业务操作，不得绕过服务层直接访问底层接口。
- **FR-004**: 服务层必须对资源所需的分页与重试行为进行统一封装，资源层不需要处理分页细节。
- **FR-005**: 改造后必须保持现有资源 Schema 与行为兼容，避免破坏既有配置与状态。
- **FR-006**: 系统必须为关键失败场景提供明确可诊断的错误反馈，便于定位问题。

### Key Entities

- **AliKafka 实例**: Kafka 服务实例，包含实例标识、规格、状态等核心属性。
- **主题**: 实例内的消息主题，包含名称、分区、副本等属性。
- **SASL 用户**: 访问实例的认证主体，包含用户名、状态等属性。
- **SASL ACL**: SASL 用户的访问控制规则，包含授权范围与权限类型。
- **消费者组**: 消费者分组与其运行状态信息。
- **部署**: 实例相关的部署配置与状态信息。
- **实例白名单绑定**: 实例访问控制的 IP 白名单关联关系。

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 7 个 AliKafka 资源的核心操作路径全部通过服务层完成，且不存在绕过服务层的操作路径。
- **SC-002**: 对每个资源的创建/读取/删除流程均可一次完成并正确同步状态，验收通过率达到 100%。
- **SC-003**: 既有资源配置无需修改即可完成核心操作，向后兼容覆盖率达到 100%。
- **SC-004**: 服务层为每类资源至少提供一组可复用的核心操作能力，覆盖范围达到 7/7。

## Clarifications

<!-- 
This section will be populated by /speckit.clarify command with questions and answers.
Format: - Q: <question> → A: <answer>
-->
