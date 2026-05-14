# Requirements Specification: CMS Provider 封装能力改造

**Requirement Branch**: `007-refactor-cms-api`  
**Created**: 2026-05-14  
**Status**: Draft  
**Input**: User description: "` for f in  {service,resource,data_source}_alicloud_cms_{addon,agg_task,alert,cloud_resource,context,context_store,dataset,delivery_task,entity_store,integration_policy,memory,memory_store,pipeline,prometheus,service,umodel,workspace}.go;do echo "$f";done`这个shell脚本获取所有和cms相关的Provider代码，我想将这个代码都改造成使用github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api中提供的封装API，其中ALIYUN CMS服务对应的本地代码在/cws_data/terraform-provider-alicloud/pkg/cws-lib-go/lib/cloud/aliyun/api/cms目录。"

## Related Feature *(mandatory)*

<!--
  ACTION REQUIRED: Keep the default values as "Need clarification" in the initial draft.
  /speckit.clarify must resolve this section to the final Feature binding before planning.
-->

**Feature ID**: Need clarification  
**Feature Name**: Need clarification

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

### User Story 1 - CMS 代码统一使用封装能力 (Priority: P1)

作为 Provider 维护者，我希望指定范围内的 CMS 资源、数据源与服务代码统一通过本地封装能力完成云服务交互，以便消除重复调用路径并符合项目分层原则。

**Why this priority**: 这是本次改造的核心目标，直接决定 CMS 代码是否具备一致的维护边界与可复用能力。

**Independent Test**: 可独立核对脚本枚举出的 CMS 文件，确认每个需要云服务交互的入口均通过服务层与封装能力完成核心操作，并保持原有 Terraform 用户体验。

**Acceptance Scenarios**:

1. **Given** 脚本枚举出的 CMS Provider 文件，**When** 完成改造审查，**Then** 所有需要访问 CMS 的资源与数据源均不再绕过本地封装能力。
2. **Given** 用户使用既有 CMS 资源或数据源配置，**When** 执行创建、读取、更新、删除或查询流程，**Then** 用户无需修改配置即可获得与改造前一致或更清晰的结果。

---

### User Story 2 - 服务层补齐 CMS 生命周期能力 (Priority: P2)

作为贡献者，我希望 CMS 服务层具备覆盖资源与数据源所需的完整操作能力，以便资源层和数据源层只表达 Terraform 行为而不承载云服务交互细节。

**Why this priority**: 服务层能力完整性是统一调用路径的前提，可降低后续新增 CMS 能力时的重复成本。

**Independent Test**: 可独立检查每类 CMS 对象的服务层操作是否覆盖其核心生命周期或查询需求，并通过资源层或数据源层调用完成验证。

**Acceptance Scenarios**:

1. **Given** 某个 CMS 对象缺少服务层操作，**When** 补齐后资源层或数据源层调用该能力，**Then** 调用方无需处理底层交互细节即可完成目标操作。
2. **Given** CMS 查询结果存在分页或空结果，**When** 服务层返回给资源层或数据源层，**Then** 调用方获得完整且可判定的业务结果。

---

### User Story 3 - CMS 用户体验保持兼容 (Priority: P3)

作为 Terraform 使用者，我希望 CMS 相关资源和数据源的可用性、字段语义和错误反馈不因内部改造而回退，以便现有配置可以安全升级。

**Why this priority**: 内部改造不应破坏已有使用方式，兼容性决定改造能否安全发布。

**Independent Test**: 可独立使用已有 CMS 配置样例验证核心流程，确认配置无需变更且状态、查询结果与错误信息可被用户理解。

**Acceptance Scenarios**:

1. **Given** 已有 CMS 资源配置，**When** 升级到改造后的 Provider 并执行计划与应用，**Then** 不出现由内部调用路径变化引起的非预期差异。
2. **Given** CMS 云服务返回业务错误或对象不存在，**When** Provider 将结果呈现给用户，**Then** 用户能够获得明确、可诊断且与当前行为兼容的反馈。

---

### Edge Cases

- 脚本枚举出的文件中存在当前仅有占位内容或暂未注册的资源/数据源时，必须区分“需要改造的已启用能力”和“待启用能力”，避免把占位文件误判为可交付功能。
- 某些 CMS 对象只支持查询、只支持异步变更或缺少更新能力时，必须按其实际生命周期定义可验证的完成标准。
- 本地封装能力与 Provider 既有字段命名、默认值或错误语义不完全一致时，必须优先保持用户可见行为兼容。
- CMS 查询结果为空、对象已被外部删除或云服务返回临时性错误时，必须保证状态同步和诊断信息清晰。
- 多个 CMS 对象之间存在依赖关系时，必须避免单个对象改造破坏其他对象的创建、读取或删除顺序。

## Assumptions

- 本需求范围以用户提供的脚本枚举结果为准，覆盖 CMS 的服务层、资源层与数据源层文件。
- 本次改造不新增用户可见的 Terraform 资源或数据源能力，除非该能力已在现有范围内明确需要启用。
- 现有 CMS 资源和数据源的 Schema、导入标识、状态语义和文档承诺默认保持兼容。
- 本地封装能力已覆盖主要 CMS 对象；如个别对象缺口存在，应在计划阶段作为补齐任务处理。
- 功能归属初步匹配到 CWS-Lib-Go Integration，但 `Related Feature` 字段按流程保留待澄清。

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: 系统必须识别并纳入用户脚本枚举出的全部 CMS Provider 文件，形成可审查的改造范围。
- **FR-002**: 系统必须确保 CMS 资源层和数据源层仅通过服务层完成业务操作，不得直接承担云服务交互职责。
- **FR-003**: 系统必须为脚本范围内的每类 CMS 对象提供可复用的服务层能力，覆盖其创建、读取、更新、删除或查询等适用操作。
- **FR-004**: 系统必须通过本地封装能力完成 CMS 云服务交互，并保持统一的请求、响应、错误和分页处理边界。
- **FR-005**: 系统必须保持现有 CMS 用户可见行为兼容，包括资源/数据源名称、字段含义、状态处理、导入方式和错误反馈。
- **FR-006**: 系统必须为每类已启用 CMS 资源和数据源提供可验证的改造证据，说明其调用路径已符合分层架构。
- **FR-007**: 系统必须在发现封装能力缺口、对象生命周期差异或兼容性风险时记录明确的处理结果，避免静默跳过。
- **FR-008**: 系统必须确保 CMS 查询类能力返回完整结果，并由服务层统一处理分页、空结果和对象不存在等场景。
- **FR-009**: 系统必须保持改造范围可追踪，使后续评审能够从脚本枚举结果映射到对应对象、服务能力和验证结果。

### Key Entities

- **CMS Provider 文件**: 用户脚本枚举出的服务层、资源层和数据源层文件，是本次范围控制和验收的基础对象。
- **CMS 资源对象**: 表示可由用户声明并由 Provider 管理生命周期的 CMS 能力，例如工作空间、数据集、告警、投递任务等。
- **CMS 数据源对象**: 表示可由用户查询并用于配置编排的 CMS 信息集合，例如工作空间、Prometheus、集成策略等查询结果。
- **服务层能力**: 表示供资源层和数据源层复用的业务操作集合，负责生命周期、查询、状态同步与错误归一。
- **封装能力覆盖项**: 表示某个 CMS 对象与本地封装能力之间的对应关系，包含对象名称、支持操作、兼容性状态和验证结果。

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% 脚本枚举出的 CMS Provider 文件完成范围判定，并具备“已改造、无需改造或待启用”的明确结论。
- **SC-002**: 100% 已启用 CMS 资源和数据源的用户可见核心流程在改造后保持兼容，无需用户修改既有配置。
- **SC-003**: 100% 已启用 CMS 资源和数据源的云服务交互路径符合资源/数据源到服务层再到封装能力的分层边界。
- **SC-004**: 每类已启用 CMS 对象至少具备 1 组可复用服务层能力与对应验证记录，覆盖率达到 100%。
- **SC-005**: CMS 改造导致的新增使用问题在发布后两个反馈周期内不高于改造前基线，且所有新增问题均可定位到具体对象和调用路径。

### Measurement Sources & Collection Methods

- **SC-001 Source**: 通过脚本枚举清单与改造范围清单逐项比对；计划阶段建立基线，实施完成后复核一次。
- **SC-002 Source**: 通过既有 CMS 配置样例、验收用例和用户可见变更审计统计兼容结果；每个发布候选版本测量一次。
- **SC-003 Source**: 通过代码审查清单和调用路径审计统计符合分层边界的对象数量；每轮评审后更新。
- **SC-004 Source**: 通过服务层能力矩阵与验证记录统计覆盖率；每完成一类 CMS 对象后更新。
- **SC-005 Source**: 通过 issue、工单和发布反馈中 CMS 标签统计新增问题数量与定位结果；发布后按反馈周期汇总。

## Clarifications

<!-- 
This section will be populated by /speckit.clarify command with questions and answers.
Format: - Q: <question> → A: <answer>
-->
