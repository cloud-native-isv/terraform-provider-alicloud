# Feature Specification: Add Terraform resource alicloud_sls_consumer_group

**Feature Branch**: `003-add-alicloud-sls`  
**Created**: 2025-10-28  
**Status**: Draft  
**Input**: User description: "add alicloud_sls_consumer_group resource: 添加自定义sls alicloud_sls_consumer_group resource，名称为alicloud_sls_consumer_group，使用func resourceAliCloudSlsConsumerGroup() *schema.Resource等函数实现。新代码需要遵循现有代码的规范，现有的sls相关代码使用如下命令获取：cd /cws_data/terraform-provider-alicloud/alicloud && ls -l resource_alicloud_sls_* data_source_alicloud_sls_* service_alicloud_sls_*"

## Clarifications

### Session 2025-10-28

- Q: 资源字段的可变更策略如何？哪些字段变更需要重建，哪些可原位更新？ → A: Option A（project、logstore、consumer_group 变更需重建；仅 timeout、order 可原位更新）
- Q: Import/内部 ID 的编码格式为何？ → A: Option A（使用冒号分隔的三段式：project:logstore:consumer_group）
- Q: Create 行为：若同名消费组已存在如何处理？ → A: Option A（存在即接管并对齐 HCL 期望：更新 timeout/order）

## User Scenarios & Testing *(mandatory)*

### User Story 1 - 以基础参数创建并管理 SLS 消费组 (Priority: P1)

作为 Terraform 使用者（DevOps/平台工程师），我可以在 HCL 中声明一个 SLS 消费组（指定项目、日志库、消费组名称、超时窗口与是否顺序消费），执行 `apply` 后资源被成功创建；后续我修改参数并 `apply`，状态被正确同步；当我 `destroy` 时资源被安全删除。

**Why this priority**: 这是该资源最核心的价值——通过 IaC 管理 SLS 消费组的生命周期（创建/更新/删除）。

**Independent Test**: 仅通过一个最小化配置（project、logstore、consumer_group、timeout、order）即可验证全生命周期 CRUD，独立可演示。

**Acceptance Scenarios**:

1. Given 存在指定的 Project 与 Logstore，When 以唯一的 consumer_group 名称执行 `apply`，Then 消费组在目标 Logstore 下被创建并在 `terraform state` 中可见。
2. Given 已创建的消费组，When 修改 `timeout` 或 `order` 并 `apply`，Then 资源状态与远端一致，计划与实际无漂移。
3. Given 已创建的消费组，When 执行 `destroy`，Then 资源被删除且再次读取返回不存在，状态清空。
4. Given 同名消费组已存在但配置不同，When 执行 `apply`，Then 资源被接管并对齐到 HCL 期望（更新 timeout/order），后续 `plan` 无漂移。

---

### User Story 2 - 导入现有消费组纳入托管 (Priority: P2)

作为已有资源迁移用户，我可以将既有的 SLS 消费组通过 `terraform import` 纳入状态管理，后续对参数调整与删除操作均受 Terraform 控制。

**Why this priority**: 导入能力是 Terraform 资源的常规需求，便于存量资源迁移与统一治理。

**Independent Test**: 构造仅包含 `id` 的占位配置，执行 `terraform import`，其中 `id` 使用格式 `project:logstore:consumer_group`；随后 `terraform plan` 应无变更或仅提示不可导入字段。

**Acceptance Scenarios**:

1. Given 远端已存在消费组，When 提供正确的导入 ID 并执行 `terraform import`，Then `terraform state` 中出现该资源且 `plan` 与远端一致。

---

### User Story 3 - 输入校验与可用性反馈 (Priority: P3)

作为资源使用者，当我提供不合法的名称、缺失必填字段或超出阈值的配置时，能获得清晰的错误提示，避免不必要的 API 调用；遇到暂时性错误时会自动重试并在超时后给出可理解的失败信息。

**Why this priority**: 提升使用体验，减少无效调用，改进问题定位效率。

**Independent Test**: 针对非法名称、缺字段、越界超时等情形分别执行 `plan/apply` 并断言错误信息与行为符合预期。

**Acceptance Scenarios**:

1. Given 缺少 `project` 或 `logstore`，When 执行 `plan/apply`，Then 立即给出必填项缺失错误并阻止操作。
2. Given 名称不符合命名约束，When 执行 `plan`，Then 本地校验失败并提示命名规范。

---

### Edge Cases

- 创建时 consumer_group 已存在于同一 Project/Logstore 下：执行接管（adopt）并对齐 HCL 期望值，仅调整可更新参数（timeout、order），保证后续 `plan` 无漂移；若标识字段不同则失败并指引更正配置或导入正确对象。
- 删除时资源已不存在：操作应安全完成且状态清空。
- 长时间后端处理（例如平台繁忙）：在合理超时内进行带退避的重试；超时后失败并给出可理解信息。
- 区域或项目/日志库不一致：应返回清晰的定位信息，提示用户检查 provider 区域或资源定位。
- 参数变更导致需要 Recreation 的情形（例如名称/绑定目标变更）：`plan` 应明确标注替换；其中 project、logstore、consumer_group 任一变更需替换（ForceNew），timeout 与 order 可原位更新。

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: 支持通过声明（project、logstore、consumer_group）创建消费组，并在 `apply` 后可见。
- **FR-002**: 支持更新可变参数（如顺序消费开关、超时窗口等），并确保状态与远端一致且无漂移。
- **FR-003**: 支持删除消费组，删除后再次读取应判定为不存在并清空状态。
- **FR-004**: 支持 `terraform import` 将现有消费组纳入状态管理；导入后 `plan` 不应出现非必要变更。
- **FR-005**: 对必填字段进行本地校验（例如 project、logstore、consumer_group 为必填；名称长度/字符集限制）。
- **FR-006**: 对关键参数提供合理默认值（例如 超时窗口、顺序消费开关），并在文档中清晰说明。
- **FR-007**: 对常见的后端暂时性错误进行自动重试，并提供超时控制；对不可重试错误直接失败且包含上下文信息。
- **FR-008**: 在 `plan`/`apply` 各阶段输出清晰的用户可读描述（如将进行创建/替换/删除）。
- **FR-009**: 在读取阶段完整回填（Computed）属性，保证幂等与可见性。
- **FR-010**: 支持超时配置（Create/Update/Delete）并在到达超时后给出失败结果。

- **FR-011**: 字段可变更策略：project、logstore、consumer_group 为标识字段，任意变更均触发资源替换（ForceNew）；timeout 与 order 为行为参数，支持原位更新。

- **FR-012**: Import 行为：接受并解析 ID 为 `project:logstore:consumer_group` 的三段式格式；解析失败需给出清晰的用户提示。

- **FR-013**: Create 幂等策略：若同名消费组已存在，则自动接管并将可更新参数（timeout、order）收敛至 HCL 期望；标识字段不一致时应失败并提供修复建议。

### Key Entities *(include if feature involves data)*

- **ConsumerGroup（消费组）**: 隶属于某 Project 与 Logstore；具有名称、超时窗口、是否顺序消费等属性；在生命周期内可被创建、更新、删除与导入。
- **Project（项目）**: SLS 的资源命名空间，承载一个或多个 Logstore。
- **Logstore（日志库）**: 日志数据容器；消费组依附于 Logstore 存在。

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 使用者可在 5 分钟内完成从首次声明到成功创建并验证读取的全流程（含一次 plan 与一次 apply）。
- **SC-002**: 95% 的创建与更新操作在设定的超时阈值内完成并返回可理解反馈。
- **SC-003**: 导入现有消费组后执行 `plan`，无额外非预期变更（≥ 95% 情况）。
- **SC-004**: 针对必填项缺失或命名不合法，`plan` 阶段即可拦截并提供明确提示（误报率 < 1%）。
