# AI 助手指令

## 概述

你是一个专家级的 AI 编程助手（GitHub Copilot），正在协助用户开发 Terraform Provider Alicloud 项目。

请务必遵循以下核心原则和指令。详细的开发指南、架构设计原则和代码示例，请参考 **[docs/development_guide.md](../docs/development_guide.md)** 文档。

## 1. 核心指令

### 1.1 参考文档
- **所有代码生成和修改任务，必须首先参考 `docs/development_guide.md` 中的规范和最佳实践。**
- 该文档包含了详细的架构分层、API 调用规范、错误处理模式和状态管理规范。

### 1.2 代码生成规范
- **文件操作**:
    - 复杂文件操作时，先生成 Python 或 Shell 脚本，然后执行脚本。
    - 批量操作前务必备份。
    - 使用版本控制跟踪所有变更。
- **语言使用**:
    - 生成文档时使用中文。
    - 代码注释和日志使用英文。
    - API 文档和错误消息使用英文以保持国际化兼容性。
- **代码拆分**:
    - 编程语言代码文件（*.go 等）超过 1000 行时需要拆分。
    - 按功能模块拆分，确保每个文件职责单一且清晰。
- **代码验证**:
    - **每次代码生成后必须验证语法**：执行 `cd /cws_data/terraform-provider-alicloud && make` 命令验证代码语法正确性。
    - 确保代码能够正常编译通过，无语法错误和导入错误。
    - 如果编译失败，必须修复所有错误后才能继续下一步开发。
- **构建产物**:
    - **不要在根目录生成二进制文件**：所有二进制文件只能输出到 `bin` 目录中。
    - 必须在 `.gitignore` 中忽略 `bin` 目录中的二进制文件。

### 1.3 架构原则摘要
- **分层架构**: Resource/DataSource -> Service -> API (CWS-Lib-Go) -> SDK。
- **强类型**: 必须优先使用 CWS-Lib-Go 提供的强类型，严禁在新代码中使用 `map[string]interface{}`。
- **状态管理**: 必须在 Service 层实现 `StateRefreshFunc` 和 `WaitFor` 函数，禁止在 Resource 层的 Create 方法中直接进行轮询或调用 Read。

## 2. 特定环境信息

### 2.1 遗留库
- **CWS-Lib-Go**: CWS-Lib-Go 库已通过 git submodule 引入到项目根目录下的 `pkg/cws-lib-go` 目录中。相关的 API 定义位于 `pkg/cws-lib-go/lib/cloud/aliyun/api/` 目录。可以使用 `find`, `cat`, `awk`, `grep` 等命令查看这些库文件以获取类型定义和 API 接口信息。

## Active Technologies
- Go 1.22+ (002-fix-oss-force-destroy)
- N/A (Cloud Resource) (002-fix-oss-force-destroy)
- Go 1.22 + `github.com/aliyun/aliyun-oss-go-sdk/oss` (002-fix-oss-force-destroy)
- Alibaba Cloud OSS (002-fix-oss-force-destroy)
- Go 1.20+ + terraform-plugin-sdk v1.17.x, cws-lib-go OSS API, aliyun-oss-go-sdk (indirect) (003-oss-prune-bucket)
- N/A (remote OSS service) (003-oss-prune-bucket)

## Recent Changes
- 002-fix-oss-force-destroy: Added Go 1.22+
