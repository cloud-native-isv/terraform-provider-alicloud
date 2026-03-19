# Requirements Specification: 统一日志服务资源前缀

**Requirement Branch**: `006-unify-log-prefix`  
**Created**: 2026-03-19  
**Status**: Draft  
**Input**: User description: "当前项目中存在大量的alicloud_log_*前缀的resource或datasource如alicloud_log_alert、也存在大量的alicloud_sls_*前缀的resource或datasource如alicloud_sls_scheduled_sql，这两种resource实际上都是alicloud中的日志服务，需要将他们的命名规则进行统一。鉴于外部系统大量使用了alicloud_log_*前缀的resource，需要将将这些命名都统一到alicloud_log_*前缀。"

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

### User Story 1 - 使用统一前缀完成新建与维护 (Priority: P1)

作为 Terraform 使用者，我希望日志服务相关能力都以 `alicloud_log_*` 前缀提供与检索，这样我在编写与维护配置时不需要区分 `log` 与 `sls` 两套命名规则。

**Why this priority**: 这是命名统一的核心用户价值，直接决定新用户可发现性与老用户维护成本。

**Independent Test**: 仅实现该故事时，用户应可在日志服务能力范围内只使用 `alicloud_log_*` 前缀完成配置编排与查询，并获得完整功能覆盖。

**Acceptance Scenarios**:

1. **Given** 用户准备新增日志服务资源或数据源配置，**When** 用户按能力检索官方文档与示例，**Then** 所有可用能力均可通过 `alicloud_log_*` 前缀找到对应入口。
2. **Given** 用户使用同一日志服务场景，**When** 用户对比命名结果，**Then** 不会出现同一能力同时要求使用 `alicloud_sls_*` 作为唯一入口的情况。
3. **Given** 用户维护历史 `alicloud_sls_*` 前缀配置，**When** 用户查看统一范围定义，**Then** 所有 `alicloud_sls_*` 均在本次统一治理范围内。

---

### User Story 2 - 老配置可平滑过渡到统一前缀 (Priority: P2)

作为已有 `alicloud_sls_*` 配置的使用者，我希望有明确的重建迁移路径，确保在统一切换窗口内完成替换与验证。

**Why this priority**: 强制统一会直接影响现网配置，必须通过清晰迁移步骤控制升级不确定性。

**Independent Test**: 仅实现该故事时，用户应可依据重建步骤完成从 `alicloud_sls_*` 到 `alicloud_log_*` 的替换，并在目标版本通过校验。

**Acceptance Scenarios**:

1. **Given** 用户当前配置包含 `alicloud_sls_*`，**When** 用户按迁移说明销毁旧入口并以 `alicloud_log_*` 重建后升级到目标版本，**Then** 配置校验通过且核心业务流程结果一致。

---

### User Story 3 - 团队可治理命名一致性 (Priority: P3)

作为平台维护者，我希望能够识别并治理日志服务命名不一致项，以降低后续文档、培训与支持成本。

**Why this priority**: 长期治理有助于降低认知负担，但优先级低于用户直接使用与迁移。

**Independent Test**: 仅实现该故事时，维护者应可拿到完整的命名映射与状态标记（统一中、已统一、待下线），并据此开展版本沟通。

**Acceptance Scenarios**:

1. **Given** 维护者需要评估统一进度，**When** 查看命名治理清单，**Then** 可明确每一项日志服务能力的目标命名与当前阶段。

---

### Edge Cases

- 同一日志服务能力已同时存在 `alicloud_log_*` 与 `alicloud_sls_*` 两个入口时，必须定义唯一规范入口与迁移顺序，避免用户误判。
- 历史配置只使用 `alicloud_sls_*` 且被自动化系统批量复用时，必须提供可执行迁移指引与风险提示。
- 升级到统一版本后若仍存在 `alicloud_sls_*` 配置，必须返回明确错误并提示对应 `alicloud_log_*` 替代名称。
- 对依赖稳定资源标识的场景，销毁重建可能导致状态变化，必须在迁移说明中提供风险提示与回滚建议。
- 命名变更过程中，文档、示例与发布说明可能出现不同步，必须提供单一权威映射来源。
- 所有 `alicloud_sls_*` 前缀项默认纳入统一范围，不再按服务归属做二次排除。

## Assumptions

- 本需求范围限定为阿里云日志服务相关的 Terraform resource 与 data source 命名治理。
- 本次统一范围按前缀判定：所有 `alicloud_sls_*` 资源与数据源均纳入治理。
- 统一目标前缀为 `alicloud_log_*`，`alicloud_sls_*` 视为历史命名并进入迁移治理。
- 迁移策略采用强制统一：目标版本停止 `alicloud_sls_*` 入口，仅保留 `alicloud_log_*`。
- 停用生效策略不区分 major/minor/patch，按最近发布版本直接生效。
- 不在本需求内引入新的日志服务业务能力，仅处理命名一致性与可用性。

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: 日志服务相关的 Terraform resource 与 data source 必须提供 `alicloud_log_*` 作为统一命名入口。
- **FR-002**: 对历史 `alicloud_sls_*` 命名，必须建立一一对应或明确聚合关系的命名映射，并对不可直接映射项给出说明。
- **FR-003**: 用户必须能够通过统一命名文档，在不依赖隐式知识的情况下定位目标能力与替代关系。
- **FR-004**: 命名统一过程中必须提供可执行的一次性迁移路径，并明确目标版本停止 `alicloud_sls_*` 入口。
- **FR-005**: 每次命名状态调整必须同步更新发布说明与治理清单，确保外部系统可感知变更。
- **FR-006**: 统一范围边界必须被明确记录（纳入项、排除项、待评估项），并可供评审验证。
- **FR-007**: 在目标版本中，系统必须对 `alicloud_sls_*` 的使用给出明确错误与替代项提示，避免静默失败。
- **FR-008**: 官方迁移策略必须明确为“销毁重建”，不提供 `terraform state mv` 或自动状态迁移承诺。
- **FR-009**: 统一范围判定规则必须固定为“`alicloud_sls_*` 全量纳入”，禁止个别前缀项例外豁免。
- **FR-010**: `alicloud_sls_*` 停用生效策略必须定义为“按最近发布版本生效”，不得绑定特定语义版本类型。

### Key Entities *(include if requirement involves data)*

- **命名映射项（Naming Mapping Item）**: 表示一个日志服务能力在历史命名与统一命名之间的关系，关键属性包括能力标识、历史名称、统一名称、映射状态、说明。
- **治理清单（Governance Catalog）**: 表示本次命名统一覆盖的能力集合，关键属性包括覆盖范围、边界分类（纳入/排除/待评估）、更新时间。
- **迁移指引项（Migration Guidance Item）**: 表示一个历史命名迁移到统一命名的可执行说明，关键属性包括适用场景、前置条件、预期结果、风险提示。

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% 纳入范围内的日志服务能力可通过 `alicloud_log_*` 统一前缀被检索和使用。
- **SC-002**: 100% 的历史 `alicloud_sls_*` 能力具备明确迁移映射，并在目标版本中不再作为可用入口。
- **SC-003**: 在一次版本升级周期内，至少 90% 受影响用户可在 60 分钟内完成单个能力的重建迁移与验证。
- **SC-004**: 命名差异导致的使用咨询在两个发布周期内下降至少 50%。
- **SC-005**: 统一范围定义与实际清单的一致性达到 100%，即所有 `alicloud_sls_*` 均存在对应 `alicloud_log_*` 映射或下线说明。
- **SC-006**: 每次停用发布均在发布说明中明确“本次直接生效”的规则，发布信息覆盖率达到 100%。

### Measurement Sources & Collection Methods

- **SC-001 Source**: 通过发布前能力清点与文档审计统计纳入范围能力总数与统一命名覆盖数；每次发布候选版本测量一次。
- **SC-002 Source**: 通过命名映射清单审计统计历史命名项与可迁移说明覆盖率；每次需求变更后更新并复核。
- **SC-003 Source**: 通过内部试运行与迁移演练记录迁移耗时及成功率；按版本发布周期汇总。
- **SC-004 Source**: 通过 issue/工单中“命名混淆”标签进行月度统计，对比统一前基线与统一后两个发布周期数据。

## Clarifications

### Session 2026-03-19

- Q: 对历史 `alicloud_sls_*` 命名，目标兼容策略应是哪一种？ → A: 强制统一：下一版本立即停止 `alicloud_sls_*`，仅保留 `alicloud_log_*`。
- Q: 对已有 Terraform 状态中 `alicloud_sls_*` 资源，官方迁移方式应定义为哪一种？ → A: 不支持状态迁移，要求用户销毁重建。
- Q: 对“纳入统一范围”的判定规则，应采用哪一种？ → A: 只要名字含 `alicloud_sls_*` 就全部纳入。
- Q: 强制停用 `alicloud_sls_*` 的生效版本策略应是什么？ → A: 不区分版本语义，随最近发布生效。

<!-- 
This section will be populated by /speckit.clarify command with questions and answers.
Format: - Q: <question> → A: <answer>
-->
