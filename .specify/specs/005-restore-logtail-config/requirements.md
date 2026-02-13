# Requirements Specification: 恢复 Logtail Config 兼容能力

**Requirement Branch**: `005-restore-logtail-config`  
**Created**: 2026-02-13  
**Status**: Draft  
**Input**: User description: "在git commit 5ee181f26776da03d621a876f1a603e0444666cd中我们将alicloud_logtail_config这个resource完全替换成了新版实现而且实现方式没有向前兼容，这对上层应用的影响太大。将这个修改中的内容进行重命名然后找回并保留旧版的实现。新版实现的重命名规则：1）resource：alicloud_logtail_config -> alicloud_logtail_pipeline_config;2）file name：resource_alicloud_logtail_config.go -> resource_alicloud_logtail_pipeline_config.go"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - 保持既有配置可持续使用 (Priority: P1)

作为正在使用既有 `alicloud_logtail_config` 的 Terraform 用户，我希望升级 provider 后无需立即重写现有配置，仍可执行计划与变更，以避免生产环境中断和大规模改造。

**Why this priority**: 直接影响存量用户的可用性与升级风险，是本需求的核心业务价值。

**Independent Test**: 使用仅包含既有 `alicloud_logtail_config` 语法的配置执行初始化、计划与应用，流程可完整通过且结果符合预期。

**Acceptance Scenarios**:

1. **Given** 用户已有仅使用 `alicloud_logtail_config` 的配置，**When** 升级到包含本需求的 provider 版本并执行计划，**Then** 不因资源命名变化产生“未知资源类型”或强制迁移报错。
2. **Given** 用户已有 `alicloud_logtail_config` 资源状态，**When** 执行读取与更新操作，**Then** 资源行为与变更前保持一致。

---

### User Story 2 - 新版能力以新资源名独立提供 (Priority: P1)

作为希望使用新版 Pipeline Config 能力的用户，我希望通过独立的新资源名访问新版能力，避免与旧资源语义混淆。

**Why this priority**: 新旧能力分离是控制兼容风险与降低认知负担的关键。

**Independent Test**: 新建仅使用 `alicloud_logtail_pipeline_config` 的配置并完成创建、更新、删除全流程。

**Acceptance Scenarios**:

1. **Given** 用户使用新版资源名 `alicloud_logtail_pipeline_config`，**When** 执行计划与应用，**Then** 可以正常完成资源生命周期操作。
2. **Given** 用户同时维护旧资源与新资源并指向同一远端对象，**When** 分别执行计划与应用，**Then** 两者可被独立识别，且写入结果按“最后一次成功写入生效”呈现可预测行为。

---

### User Story 3 - 升级路径清晰可控 (Priority: P2)

作为平台维护者，我希望新旧资源并存期间有清晰边界，便于分批迁移并控制业务风险。

**Why this priority**: 减少一次性重构压力，支持灰度迁移与变更治理。

**Independent Test**: 针对旧资源维持稳定、针对新资源提供新能力；在同一工作空间内分别验证均可独立演进。

**Acceptance Scenarios**:

1. **Given** 项目存在多套环境，**When** 仅对部分环境切换到新资源名，**Then** 其余环境仍可继续使用旧资源名且不受影响。

### Edge Cases

- 当用户配置中同时出现旧资源与新资源并指向同一远端对象时，系统按“最后一次成功写入生效”处理，并确保结果可被后续读取验证。
- 当用户升级后仅执行只读操作（如计划/刷新）时，不应被迫立即修改旧资源名。
- 当用户误将旧资源配置直接复制为新资源而未调整语义时，应提供可理解的错误反馈，避免静默错误。

## Assumptions

- 本需求聚焦于资源命名与兼容行为，不要求在本阶段完成自动迁移工具。
- 既有 `alicloud_logtail_config` 的用户预期是“最小变更升级”，即优先保持现有行为稳定。
- 新版能力继续存在，但以 `alicloud_logtail_pipeline_config` 作为明确入口。
- 旧资源的兼容支持窗口至少覆盖 2 个小版本，之后再评估后续策略。

## Out of Scope

- 本次不包含将旧资源状态自动迁移到 `alicloud_logtail_pipeline_config`。

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: 系统 MUST 保留 `alicloud_logtail_config` 资源类型，确保存量配置在升级后仍可被识别并执行。
- **FR-002**: 系统 MUST 将当前新版能力以 `alicloud_logtail_pipeline_config` 资源类型对外提供，避免占用旧资源名。
- **FR-003**: 系统 MUST 确保旧资源与新资源可在同一工作空间并存；当两者指向同一远端对象时，允许同时管理并采用“最后一次成功写入生效”规则。
- **FR-004**: 系统 MUST 维持旧资源在读取、计划、应用等关键流程中的可用性与行为连续性。
- **FR-005**: 系统 MUST 在用户错误混用新旧语义时返回明确、可操作的错误信息。
- **FR-006**: 系统 MUST 为本次命名调整提供清晰说明，包含旧资源继续可用与新资源使用入口。
- **FR-007**: 系统 MUST 对 `alicloud_logtail_config` 提供至少 2 个小版本的兼容支持窗口，并在窗口结束前给出后续策略说明。
- **FR-008**: 系统 MUST 在文档中标注 `alicloud_logtail_config` 的弃用状态与兼容窗口；运行时不新增弃用提示。
- **FR-009**: 系统 MUST 仅提供新旧资源并存能力与迁移指引；不提供自动或半自动状态迁移机制。

### Key Entities *(include if requirement involves data)*

- **Legacy Logtail Config Resource**: 面向存量用户的旧资源实体，以稳定兼容为目标。
- **Pipeline Logtail Config Resource**: 面向新版能力的新资源实体，以清晰语义和独立演进为目标。
- **Terraform Resource State**: 记录资源实例身份与期望状态，用于保证升级前后行为一致与可追踪。

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% 的既有 `alicloud_logtail_config` 示例配置在升级后可完成计划阶段且无“资源类型不存在”错误。
- **SC-002**: 覆盖创建、更新、删除三类关键流程的验证中，`alicloud_logtail_pipeline_config` 的通过率达到 100%。
- **SC-003**: 在并存场景验证中，新旧资源的独立识别成功率达到 100%，且同一对象冲突写入结果与“最后一次成功写入生效”规则一致率达到 100%。
- **SC-004**: 与该变更相关的升级阻断类问题（资源名不兼容导致的阻断）相较变更前版本下降至少 90%。
- **SC-005**: 在兼容窗口内，涉及 `alicloud_logtail_config` 的回归验证在每个小版本发布前通过率保持 100%。
- **SC-006**: 发布说明与资源文档在变更发布时 100% 包含旧资源弃用状态、兼容窗口和新资源替代路径。

## Clarifications

### Session 2026-02-13

- Q: 当 `alicloud_logtail_config` 与 `alicloud_logtail_pipeline_config` 指向同一个远端 Logtail 配置对象时，应采用哪种规则？ → A: 允许并存并同时管理同一远端对象（最后写入生效）。
- Q: 旧资源 `alicloud_logtail_config` 的兼容支持周期应如何定义？ → A: 至少兼容 2 个小版本后再评估。
- Q: 对旧资源 `alicloud_logtail_config` 的弃用提示策略应如何设置？ → A: 仅在文档中提示，不在运行时提示。
- Q: 本次范围是否包含将旧资源状态自动迁移到新资源（`alicloud_logtail_pipeline_config`）？ → A: 不包含，仅提供并存与文档迁移指引。
