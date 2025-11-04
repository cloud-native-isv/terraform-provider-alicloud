# Implementation Plan: 更新VPC资源以使用新的CWS-Lib-Go API

**Branch**: `001-cws-data-cws` | **Date**: 2025-11-04 | **Spec**: /cws_data/terraform-provider-alicloud/.specify/specs/001-cws-data-cws/spec.md
**Input**: Feature specification from `/.specify/specs/001-cws-data-cws/spec.md`

**Note**: 本计划由 speckit 规划流程生成与完善，执行至 Phase 2 规划即止。

## Summary

基于新引入的 CWS-Lib-Go VPC API，对 Provider 中以下资源实现进行升级改造：EIP 地址（alicloud_eip_address）、EIP 关联（alicloud_eip_association）、NAT 网关（alicloud_nat_gateway）、SNAT 条目（alicloud_snat_entry），以实现：
- 通过 Service 层调用 CWS-Lib-Go API（替代直接 RPC/SDK 调用），符合分层架构治理；
- 完整的状态刷新与 WaitFor 模式，避免在 Create 中直接 Read；
- 统一的错误处理与重试机制，保持向后兼容；
- 对 CWS-Lib-Go 中暂缺 API，输出需求文档并作为后续任务。

## Technical Context

**Language/Version**: Go 1.24  
**Primary Dependencies**: HashiCorp Terraform Plugin SDK v1.17.2; github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api; Alibaba Cloud official SDKs (wrapped by CWS-Lib-Go)  
**Storage**: N/A  
**Testing**: go test; terraform-plugin-sdk acceptance tests (terraform-plugin-test v2)  
**Target Platform**: Linux server  
**Project Type**: single (Terraform Provider in Go)  
**Performance Goals**: 满足规格中的 SC-001（资源操作响应时间提升≥20%）且无稳定性回退  
**Constraints**: 完全遵循宪章分层与错误/状态管理规范；保持向后兼容；不得直接发起 HTTP/RPC  
**Scale/Scope**: 影响 4 个 VPC 相关资源；在 `alicloud/` 目录内进行受限范围的重构与新增 Service 层文件

## Constitution Check

GATE（设计前检查）
- 架构分层：资源层通过 Service 层调用 CWS-Lib-Go API，避免直接 SDK/RPC → 符合
- 状态管理：Create 使用 WaitFor；Read 设置 computed 字段；Delete 使用 StateChangeConf → 符合
- 错误处理：优先使用 IsNotFoundError/IsAlreadyExistError/NeedRetry，统一 WrapError → 符合
- 代码质量：命名约定（Id 字段、资源命名）、Schema 描述齐全、分页封装在 API 层 → 符合
- 验证要求：变更后可编译通过并具备基本/验收测试；本阶段仅产出设计与文档，不改变编译结果 → 无阻塞

Re-check（设计后复核）
- 设计产物（research.md、data-model.md、contracts、quickstart.md）均已围绕上述原则展开，未引入例外 → 通过

## Project Structure

### Documentation (this feature)

```
/cws_data/terraform-provider-alicloud/.specify/specs/001-cws-data-cws/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
└── contracts/
```

### Source Code (repository root)

```
/cws_data/terraform-provider-alicloud/
├── alicloud/
│   ├── resource_alicloud_eip_address.go            # 将改造为调用 Service 层
│   ├── resource_alicloud_eip_association.go        # 将改造为调用 Service 层
│   ├── resource_alicloud_nat_gateway.go            # 将改造为调用 Service 层
│   ├── resource_alicloud_snat_entry.go             # 将改造为调用 Service 层
│   ├── service_alicloud_vpc_eip.go                 # 新增：EIP 相关 Service + 状态/等待
│   ├── service_alicloud_vpc_nat_gateway.go         # 新增：NAT Gateway 相关 Service
│   └── service_alicloud_vpc_snat_entry.go          # 新增：SNAT Entry 相关 Service
├── go.mod
└── ...
```

**Structure Decision**: 维持单仓单 Provider 的 Go 项目结构；在 `alicloud/` 下新增 Service 层文件（service_alicloud_vpc_*.go），资源文件仅调用 Service 层，完全符合治理与分层原则。

## Complexity Tracking

无当前需豁免的宪章违规项。
