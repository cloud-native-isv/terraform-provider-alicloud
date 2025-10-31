# Feature Specification: ConcurrentUpdateBucketFailed error support

**Feature Branch**: `004-concurrentupdatebucketfailed-error-support`  
**Created**: 2025-10-31  
**Status**: Draft  
**Input**: User description: "ConcurrentUpdateBucketFailed error support: 在调用alicloud/resource_alicloud_oss_public_access_block.go:79: Resource alicloud_oss_bucket_public_access_block /?publicAccessBlock Failed!!! [SDK alibaba-cloud-sdk-go ERROR]:的时候出现的错误，<error>SDKError: StatusCode: 409; Code: ConcurrentUpdateBucketFailed; Message: Concurrent update bucket detected, please cooldown and retry later.; Data: { ... }</error>，需要在resource_alicloud_oss_public_access_block.go中添加针对这个错误的重试逻辑。"

## User Scenarios & Testing (mandatory)

### User Story 1 - 自动重试并成功应用 (Priority: P1)

作为使用 Terraform 管理 OSS 桶“公共访问阻断”设置的用户，当后端因并发更新导致 409 ConcurrentUpdateBucketFailed 时，提供者应自动进行合理的冷却与重试，尽量在资源超时窗口内完成设置，用户无需手动重试。

**Why this priority**: 该问题直接导致计划执行失败，影响主路径（apply）的成功率与用户体验。

**Independent Test**: 通过并发触发对同一桶公共访问阻断设置的变更，模拟 409 错误，验证在无人工干预下，计划最终成功完成。

**Acceptance Scenarios**:

1. Given 对同一 OSS 桶的公共访问阻断设置存在并发变更，When 发生 409 ConcurrentUpdateBucketFailed，Then 提供者应按回退算法自动重试，且在可用超时内成功完成。
2. Given 自动重试期间未再出现 409，When 下一次调用返回成功，Then 本次操作应立即结束并报告成功。

---

### User Story 2 - 重试后仍失败的清晰反馈 (Priority: P2)

作为用户，如果并发冲突持续存在超出可接受的重试窗口，系统应尽快失败并提供清晰、可操作的错误信息，提示用户稍后重试或排查并发来源。

**Why this priority**: 降低排障时间成本，避免无止境等待，保障可预期性。

**Independent Test**: 持续制造并发冲突直至超时，验证错误信息包含冲突原因与建议动作。

**Acceptance Scenarios**:

1. Given 并发冲突持续，When 达到资源操作的超时阈值，Then 终止重试并返回包含“并发更新检测、请冷却后重试”的清晰提示。

---

### User Story 3 - 向后兼容且无额外配置 (Priority: P3)

作为现有用户，我不需要新增配置项即可获得该稳定性增强；在无并发冲突情况下，既有行为与时延不受显著影响。

**Why this priority**: 降低升级成本，避免引入不必要的配置复杂度。

**Independent Test**: 在无并发干扰下执行 apply，验证时延与结果与变更前一致或等效。

**Acceptance Scenarios**:

1. Given 无并发冲突，When 执行设置，Then 操作一次成功且时延不显著增加。

### Edge Cases

- 同一资源被外部系统（控制台/自动化任务/策略）同时变更，触发连续 409：应采用指数退避+抖动，避免放大冲突。
- 计划中其他错误（鉴权、配额、参数非法）不应被误判为可重试并发错误。
- 创建与更新路径都可能触发该错误；删除和读取流程不受影响（不加入该专项重试）。
- 在用户自定义超时较短时，应尊重超时边界，避免过度延迟整体计划。

## Requirements (mandatory)

### Functional Requirements

- FR-001: 系统必须识别 OSS 返回的 409 并发更新错误（业务码 ConcurrentUpdateBucketFailed），将其视为可重试的临时冲突。
- FR-002: 系统必须在“创建/更新”公共访问阻断设置的操作中对该错误执行自动重试，使用合理的指数退避与随机抖动策略，初始等待不超过数秒，最大等待遵循资源操作超时边界。
- FR-003: 重试应在以下任一条件满足时终止：
  - 成功完成目标操作（立即结束并返回成功）；
  - 达到创建/更新操作的超时阈值（结束并返回明确、可读的错误信息）。
- FR-004: 系统必须记录每次重试的次数与最后一次错误摘要于日志中，便于排障（不暴露敏感信息）。
- FR-005: 对于非并发类错误（鉴权失败、参数错误、资源不存在等），不得纳入该专项重试路径，应快速失败并反馈真实原因。
- FR-006: 读取（Read）流程保持幂等，不针对该并发错误额外重试；删除（Delete）流程不在本次变更范围内。
- FR-007: 该能力默认开启，不引入新用户配置项；遵循现有资源超时设置（Create/Update Timeouts）。

### Key Entities (if data involved)

- OSS 公共访问阻断设置：与某桶关联的布尔型策略开关，目标是将用户声明的期望状态可靠一致地应用至后端。

## Assumptions

- 将并发冲突视为短暂性问题，常可在冷却后成功；采用指数退避+抖动是行业通用稳定性策略。
- 与项目既有重试/超时规范保持一致，不引入特例；失败消息遵循现有错误信息风格。
- 不增加用户可见配置，降低向后兼容与升级成本。

## Success Criteria (mandatory)

### Measurable Outcomes

- SC-001: 在历史可复现的并发场景下，≥95% 的计划执行无需人工重试即可一次成功完成。
- SC-002: 在无并发冲突的常规场景中，90 分位执行时延较变更前不增加超过 5%。
- SC-003: 与该资源相关的“并发导致失败”的用户反馈/工单在两个月内下降 ≥70%。
- SC-004: 在开启重试后，错误信息的可读性与可操作性获得 ≥90% 内测用户正向反馈（问卷/可用性评测）。
