# Implementation Plan: ConcurrentUpdateBucketFailed error support

**Branch**: `004-concurrentupdatebucketfailed-error-support` | **Date**: 2025-10-31 | **Spec**: /.specify/specs/004-concurrentupdatebucketfailed-error-support/spec.md
**Input**: Feature specification from `/.specify/specs/004-concurrentupdatebucketfailed-error-support/spec.md`

**Note**: This plan is produced per /.github/prompts/speckit.plan.prompt.md workflow.

## Summary

目标：当 OSS 返回 409 并发更新错误（Code: ConcurrentUpdateBucketFailed）时，为资源 `alicloud_oss_bucket_public_access_block` 的创建/更新路径增加自动重试，采用指数退避+随机抖动，并在资源超时窗口内尽力成功，失败时给出清晰可操作的错误信息。技术方案基于已有的 `resource.Retry` 模式与错误判断工具函数，在 `client.Do("Oss", ...)` 调用外层添加重试判定（匹配 `ConcurrentUpdateBucketFailed`），并尊重 Terraform 超时。

## Technical Context

**Language/Version**: Go 1.24  
**Primary Dependencies**: hashicorp/terraform-plugin-sdk v1.17.2, github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api, aliyun/alibaba-cloud-sdk-go (legacy via `client.Do`)  
**Storage**: N/A  
**Testing**: `make` 编译校验；`go test ./...` 单测；必要时 Terraform acceptance（本计划阶段不强制）  
**Target Platform**: Linux server  
**Project Type**: single (Terraform Provider)  
**Performance Goals**: 在无并发冲突路径下，不显著增加时延（P90 增量 ≤ 5%）  
**Constraints**: 遵循资源超时（Create/Update Timeouts），仅对并发冲突错误重试，错误包装使用统一规范  
**Scale/Scope**: Provider 级别资源，影响范围限于 OSS 桶公共访问阻断设置

## Constitution Check

Gates from constitution (pass unless另有说明)：
- 架构分层：现有实现中 `resource_alicloud_oss_bucket_public_access_block.go` 直接调用 `client.Do`（非 Service 层封装）；本次仅加重试，不新增直连调用，属历史模式内的小改动。标记为“已知偏差”，短期内接受并在后续考虑引入 Service 层封装的 Put 接口。  
- 状态管理：Create/Update 后已使用 `BuildStateConf` + Service 层 Describe/StateRefresh；保持不变，符合规范。  
- 错误处理：将使用封装判断（`IsExpectedErrors`/`NeedRetry`）与 `WrapErrorf`；新增并发错误码进入可重试分支，符合规范。  
- 超时与重试：沿用 `resource.Retry` 并尊重 `d.Timeout(schema.TimeoutCreate/Update)`；符合规范。  

结论：除“架构分层（历史原因）”外均满足，偏差已登记于 Complexity Tracking。

### Post-Design Re-check

Phase 1 产出不改变前述评估：
- 重试仅在 Create/Update 路径引入，状态管理保持规范；
- 错误处理与封装一致；
- 架构分层偏差保持原状（历史原因），不扩大影响面。

## Project Structure

### Documentation (this feature)

```
.specify/specs/004-concurrentupdatebucketfailed-error-support/
├── plan.md              # 本文件
├── research.md          # Phase 0 输出
├── data-model.md        # Phase 1 输出
├── quickstart.md        # Phase 1 输出
└── contracts/
    └── openapi.yaml     # Phase 1 输出（用于描述重试相关的接口契约）
```

### Source Code (repository root)

```
alicloud/
├── resource_alicloud_oss_public_access_block.go   # 将在 Create/Update 路径增加并发错误重试
├── service_alicloud_oss_public_access_block.go    # 仅用于 Describe/State 刷新（已存在）
└── errors.go                                      # 复用错误判断与包装
```

**Structure Decision**: 沿用单仓 Provider 结构；仅在既有资源实现内添加重试，不新增模块。

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| Resource 直接调用 SDK (`client.Do`) 而非 Service 写操作 | 历史代码路径小改动，当前仅需加重试，风险最小 | 立即改造到 Service 层需新增/验证 Put 接口封装，超出本修复范围，影响面大、回归风险高 |
