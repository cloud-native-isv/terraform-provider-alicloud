# Requirements Specification: SLS Logtail Pipeline Config Migration

**Requirement Branch**: `004-logtail-pipeline-config`  
**Created**: 2026-02-12  
**Status**: Draft  
**Input**: User description: "\"alicloud_logtail_config\": resourceAliCloudLogtailConfig() 的 service 层从旧版 LogtailConfig API 迁移到新版 LogtailPipelineConfig API；由于新版数据结构不同，需要重新设计 schema；不考虑向前兼容，尽可能清晰合理。"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - 管理 Logtail Pipeline Config（Priority: P1）

作为 Terraform 使用者，我希望通过 `alicloud_logtail_config` 资源在指定的 SLS Project/Logstore 下创建、读取、更新、删除一个 Logtail Pipeline Config，从而以声明式方式管理采集与处理流水线配置。

**Why this priority**: 这是资源的核心价值；没有完整的 CRUD + Import，用户无法可靠地用 Terraform 管理该配置。

**Independent Test**: 在一个全新 Project/Logstore 中，仅通过创建资源即可完成配置生效，并能通过读取与 Import 验证状态一致。

**Acceptance Scenarios**:

1. **Given** 用户提供最小必填字段与一个有效的流水线配置，**When** 执行 `apply`，**Then** 资源创建成功且 `read` 能返回远端真实配置。
2. **Given** 资源已存在，**When** 执行 `import` 并再次 `plan`，**Then** 计划应无差异（无永久 diff）。
3. **Given** 用户修改一个远端允许更新的配置字段，**When** 执行 `apply`，**Then** 配置被更新且状态同步到最新值。
4. **Given** 用户删除资源，**When** 执行 `apply`，**Then** 远端配置被删除且资源从 state 移除。

---

### User Story 2 - 清晰的 Schema 与可预期的 Diff（Priority: P2）

作为 Terraform 使用者，我希望该资源的 schema 能清晰表达新版流水线配置概念（输入/处理器/输出等），并且对等价配置的表达进行规范化，避免由于字段排序、默认值或空值导致的无意义 diff。

**Why this priority**: 新版 API 数据结构变化大；若 schema 仍然依赖不透明 JSON 或不稳定规范化，用户将持续遭遇 drift 与难以定位的问题。

**Independent Test**: 对同一份逻辑等价配置，连续执行两次 `plan` 均应输出“无变更”。

**Acceptance Scenarios**:

1. **Given** 用户未显式设置某些可选字段，**When** 创建完成后再次 `plan`，**Then** 不应因为默认值回填而产生 diff。
2. **Given** 用户提供的配置包含列表/对象字段，**When** 系统保存并再次读取，**Then** 规范化后的表现应稳定且可预期（不会在每次读取时改变）。

---

### User Story 3 - 仅使用新版能力并可验证（Priority: P3）

作为 Provider 维护者，我希望该资源与其 service 层仅使用新版 Logtail Pipeline Config 能力，不再依赖旧版能力，从而降低未来兼容风险并保持与官方能力的演进一致。

**Why this priority**: 迁移的主要动机之一是淘汰旧版 API；若旧版调用残留，将导致维护成本与潜在故障。

**Independent Test**: 对本资源相关代码进行检索与自动化测试，可证明不再引用旧版端点/操作。

**Acceptance Scenarios**:

1. **Given** 代码库中存在旧版 LogtailConfig 相关调用，**When** 完成此需求交付，**Then** 对本资源/服务相关实现不应再引用旧版操作。
2. **Given** 有覆盖 CRUD + Import 的自动化测试，**When** 运行测试，**Then** 全部通过。

### Edge Cases

- 远端配置不存在（被手动删除或从未创建）时，读取应将资源标记为不存在并从 state 清理。
- 用户提交不合法或不完整的流水线配置时，应在执行期失败并给出可操作的错误信息（指出字段与原因）。
- 远端存在短暂一致性延迟时，创建/更新后短时间内读取不到最新状态，不应导致永久失败或错误 diff。
- 用户修改“远端不允许原地更新”的关键字段时，应表现为需要重建（替换）或明确报错（以行为一致、可预期为准）。

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: 系统 MUST 支持在指定的 SLS Project/Logstore 范围内，对 Logtail Pipeline Config 执行创建、读取、更新、删除，并支持导入既有配置到 Terraform state。
- **FR-002**: 系统 MUST 提供与新版流水线概念一致、可理解的资源 schema，以结构化方式表达关键配置（例如：输入、处理器、输出、样例日志等）。
- **FR-003**: 系统 MUST 对用户输入进行校验：缺失必填信息、互斥字段同时出现、明显无效值等情况必须被拒绝，并返回可操作的错误信息。
- **FR-004**: 系统 MUST 在 `read` 时将远端真实配置映射回 state，并对等价配置进行规范化，确保在无变更时 `plan` 不产生永久 diff。
- **FR-005**: 系统 MUST 明确区分“可原地更新”与“需替换”的变更类型，并在 Terraform 行为上保持一致与可预测。
- **FR-006**: 系统 MUST 在 schema 层提供合理默认值与可选项，使用户能够以最小配置快速完成创建，同时允许表达更复杂的流水线能力。
- **FR-007**: 系统 MUST 明确声明本次变更不保证向前兼容，并提供更新后的资源使用说明与示例（面向使用者）。
- **FR-008**: 系统 MUST 确保本资源对应的 service 层使用“新版 Logtail Pipeline Config 能力”完成所有核心操作（获取/创建/更新/删除），不再依赖旧版能力。
- **FR-009**: 系统 MUST 在失败时提供足够的诊断信息（例如：错误原因、相关资源标识），以便用户与维护者快速定位问题。
- **FR-010**: 系统 MUST 提供覆盖最小可用场景的自动化测试用例，至少覆盖：创建、读取、更新、导入与销毁。

### Assumptions & Dependencies

- 用户已具备可用的 SLS Project 与 Logstore，且具备创建/修改相关配置的权限。
- 目标环境已支持新版 Logtail Pipeline Config 能力（否则应失败并给出明确错误）。
- 资源名称保持为 `alicloud_logtail_config`，但其语义与可配置项以新版流水线概念为准（不承诺与旧版行为一致）。

### Key Entities *(include if requirement involves data)*

- **SLS Project**: Log Service 的项目边界；Logtail Pipeline Config 以 Project 作为命名与权限范围之一。
- **Logstore**: 日志存储目标；流水线配置通常与某个 Logstore 关联。
- **Logtail Pipeline Config**: 一份流水线配置实体；至少包含：唯一名称/标识、输入定义、处理器链、输出定义、创建/更新时间。
- **Pipeline Component**: 输入/处理器/输出等组件定义；包含类型与其配置参数。

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 新用户在阅读更新后的资源文档后，可在 $\le 5$ 分钟内完成一次成功的创建与验证（创建 + 读取）。
- **SC-002**: 对于无变更的资源，连续执行两次 `plan` 均不产生差异（无永久 diff）。
- **SC-003**: 本资源相关实现中，不再存在对旧版 LogtailConfig 能力的依赖（可通过代码检索与评审验证）。
- **SC-004**: CRUD + Import 相关的自动化测试全部通过，且覆盖 P1 场景。
- **SC-005**: 对于至少 $95\%$ 的常见配置错误，错误信息能够明确指出“哪个字段/哪类约束”导致失败（通过测试用例或人工抽样验证）。

## Clarifications

<!-- 
This section will be populated by /speckit.clarify command with questions and answers.
Format: - Q: <question> → A: <answer>
-->
