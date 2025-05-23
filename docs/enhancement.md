# Alibaba Cloud Flink Terraform Resources and Data Sources

## Resources

### 1. alicloud_flink_workspace

**描述**: 创建和管理阿里云Flink工作空间，作为运行Flink应用程序的基础环境。

**参数**:
| 参数名 | 类型 | 必选 | ForceNew | 描述 |
|-------|------|------|----------|------|
| name | string | 是 | 是 | Flink工作空间名称 |
| resource_group_id | string | 是 | 是 | 资源组ID |
| zone_id | string | 是 | 是 | 可用区ID |
| vpc_id | string | 是 | 是 | VPC ID |
| vswitch_ids | list(string) | 是 | 是 | 交换机ID列表 |
| storage | block | 是 | 是 | 存储配置块 |
| storage.oss_bucket | string | 是 | - | OSS存储桶名称 |
| resource | block | 是 | 是 | 资源配置块 |
| resource.cpu | number | 是 | - | CPU单位(毫核，如4000表示4核) |
| resource.memory | number | 是 | - | 内存大小(GB) |
| ha | block | 否 | 是 | 高可用配置块 |
| ha.zone_id | string | 是 | - | 高可用可用区ID |
| ha.vswitch_ids | list(string) | 是 | - | 高可用交换机ID列表 |
| ha.resource | block | 是 | - | 高可用资源配置块 |
| ha.resource.cpu | number | 是 | - | 高可用CPU单位(毫核) |
| ha.resource.memory | number | 是 | - | 高可用内存大小(GB) |
| tags | map(string) | 否 | 是 | 标签映射 |
| description | string | 否 | 是 | 工作空间描述 |
| security_group_id | string | 否 | 是 | 安全组ID |
| architecture_type | string | 否 | 是 | 架构类型，可选值: "X86", "ARM"，默认: "X86" |
| auto_renew | bool | 否 | 是 | 是否自动续费，默认: true |
| charge_type | string | 否 | 是 | 计费方式，可选值: "POST"(按量付费), "PRE"(包年包月)，默认: "POST" |
| duration | number | 否 | 是 | 订阅时长，默认: 1 |
| pricing_cycle | string | 否 | 是 | 计费周期，默认: "Month" |
| monitor_type | string | 否 | 是 | 监控类型，可选值: "ARMS", "TAIHAO"，默认: "ARMS" |
| promotion_code | string | 否 | 是 | 促销码 |
| use_promotion_code | bool | 否 | 是 | 是否使用促销码 |
| extra | string | 否 | 是 | 额外配置信息 |

**返回值**:
| 属性名 | 类型 | 描述 |
|-------|------|------|
| id | string | 工作空间ID |
| status | string | 工作空间状态 |
| create_time | string | 创建时间 |
| vpc_endpoint | string | VPC访问端点 |
| public_endpoint | string | 公网访问端点 |

### 2. alicloud_flink_namespace

**描述**: 在Flink工作空间内创建和管理命名空间，用于隔离不同的Flink应用程序和资源。

**参数**:
| 参数名 | 类型 | 必选 | ForceNew | 描述 |
|-------|------|------|----------|------|
| workspace_id | string | 是 | 是 | Flink工作空间ID |
| name | string | 是 | 是 | 命名空间名称 |
| description | string | 否 | 是 | 命名空间描述 |

**返回值**:
| 属性名 | 类型 | 描述 |
|-------|------|------|
| id | string | 命名空间ID，格式: "${workspace_id}:${name}" |
| create_time | string | 创建时间 |
| status | string | 命名空间状态 |

### 3. alicloud_flink_deployment

**描述**: 创建和管理Flink作业部署，定义Flink作业的运行时配置。

**参数**:
| 参数名 | 类型 | 必选 | ForceNew | 描述 |
|-------|------|------|----------|------|
| workspace_id | string | 是 | 是 | Flink工作空间ID |
| namespace | string | 是 | 是 | 命名空间名称 |
| name | string | 是 | 否 | 部署名称 |
| artifact_uri | string | 是 | 否 | 部署构件URI，如OSS中JAR文件路径 |
| flink_configuration | map(string) | 否 | 否 | Flink配置属性映射 |
| job_manager_resource_spec | block | 是 | 否 | JobManager资源规格 |
| job_manager_resource_spec.cpu | number | 是 | - | JobManager的CPU核数 |
| job_manager_resource_spec.memory | string | 是 | - | JobManager的内存，如"1g" |
| task_manager_resource_spec | block | 是 | 否 | TaskManager资源规格 |
| task_manager_resource_spec.cpu | number | 是 | - | TaskManager的CPU核数 |
| task_manager_resource_spec.memory | string | 是 | - | TaskManager的内存，如"2g" |
| environment_variables | map(string) | 否 | 否 | 环境变量映射 |
| parallelism | number | 否 | 否 | 作业并行度，默认: 1 |
| logging_profile | string | 否 | 否 | 日志配置 |
| state_backend | string | 否 | 否 | 状态后端类型，如"rocksdb"或"filesystem" |
| checkpoint_config | block | 否 | 否 | 检查点配置 |
| checkpoint_config.interval | number | 否 | - | 检查点间隔(毫秒) |
| checkpoint_config.mode | string | 否 | - | 检查点模式，如"EXACTLY_ONCE" |
| checkpoint_config.timeout | number | 否 | - | 检查点超时(毫秒) |
| checkpoint_config.min_pause_between_checkpoints | number | 否 | - | 检查点最小间隔(毫秒) |

**返回值**:
| 属性名 | 类型 | 描述 |
|-------|------|------|
| id | string | 部署ID，格式: "${namespace}:${deployment_id}" |
| deployment_id | string | 部署唯一标识符 |
| create_time | string | 创建时间 |
| update_time | string | 最后更新时间 |
| status | string | 部署状态 |
| effective_config | map(string) | 有效的配置参数 |

### 4. alicloud_flink_variable

**描述**: 在Flink命名空间中管理环境变量，用于在Flink作业中使用。

**参数**:
| 参数名 | 类型 | 必选 | ForceNew | 描述 |
|-------|------|------|----------|------|
| workspace_id | string | 是 | 是 | Flink工作空间ID |
| namespace | string | 是 | 是 | 命名空间名称 |
| name | string | 是 | 是 | 变量名称 |
| value | string | 是 | 否 | 变量值 |
| description | string | 否 | 否 | 变量描述 |
| is_sensitive | bool | 否 | 否 | 是否敏感变量，默认: false |

**返回值**:
| 属性名 | 类型 | 描述 |
|-------|------|------|
| id | string | 变量ID，格式: "${workspace_id}:${namespace}:${name}" |
| create_time | string | 创建时间 |
| update_time | string | 最后更新时间 |

### 5. alicloud_flink_member

**描述**: 管理Flink命名空间的成员访问权限。

**参数**:
| 参数名 | 类型 | 必选 | ForceNew | 描述 |
|-------|------|------|----------|------|
| workspace_id | string | 是 | 是 | Flink工作空间ID |
| namespace | string | 是 | 是 | 命名空间名称 |
| member_id | string | 是 | 是 | 成员ID(RAM用户ID或邮箱) |
| role | string | 是 | 否 | 角色，可选值: "ADMIN", "DEVELOPER", "VIEWER" |

**返回值**:
| 属性名 | 类型 | 描述 |
|-------|------|------|
| id | string | 成员ID，格式: "${namespace}:${member_id}" |
| create_time | string | 创建时间 |

### 6. alicloud_flink_deployment_draft (暂未实现)

**描述**: 管理Flink部署草稿，用于保存部署配置而不立即应用。

**参数**:
| 参数名 | 类型 | 必选 | ForceNew | 描述 |
|-------|------|------|----------|------|
| workspace_id | string | 是 | 是 | Flink工作空间ID |
| namespace | string | 是 | 是 | 命名空间名称 |
| name | string | 是 | 否 | 草稿名称 |
| deployment_id | string | 否 | 否 | 关联的部署ID |
| artifact_uri | string | 是 | 否 | 部署构件URI |
| flink_configuration | map(string) | 否 | 否 | Flink配置属性映射 |
| job_manager_resource_spec | block | 是 | 否 | JobManager资源规格 |
| task_manager_resource_spec | block | 是 | 否 | TaskManager资源规格 |
| environment_variables | map(string) | 否 | 否 | 环境变量映射 |

**返回值**:
| 属性名 | 类型 | 描述 |
|-------|------|------|
| id | string | 草稿ID |
| create_time | string | 创建时间 |
| update_time | string | 最后更新时间 |
| status | string | 草稿状态 |

### 7. alicloud_flink_connector (暂未实现)

**描述**: 管理Flink连接器，用于连接外部数据源和目标。

**参数**:
| 参数名 | 类型 | 必选 | ForceNew | 描述 |
|-------|------|------|----------|------|
| workspace_id | string | 是 | 是 | Flink工作空间ID |
| namespace | string | 是 | 是 | 命名空间名称 |
| name | string | 是 | 是 | 连接器名称 |
| type | string | 是 | 是 | 连接器类型，如"kafka", "jdbc", "oss" |
| configuration | map(string) | 是 | 否 | 连接器配置 |
| description | string | 否 | 否 | 连接器描述 |
| version | string | 否 | 是 | 连接器版本 |

**返回值**:
| 属性名 | 类型 | 描述 |
|-------|------|------|
| id | string | 连接器ID |
| create_time | string | 创建时间 |
| update_time | string | 最后更新时间 |
| status | string | 连接器状态 |

### 8. alicloud_flink_job (暂未实现)

**描述**: 管理Flink作业的运行时实例，控制作业的启动、停止等操作。

**参数**:
| 参数名 | 类型 | 必选 | ForceNew | 描述 |
|-------|------|------|----------|------|
| workspace_id | string | 是 | 是 | Flink工作空间ID |
| namespace | string | 是 | 是 | 命名空间名称 |
| deployment_id | string | 是 | 是 | 关联的部署ID |
| job_name | string | 否 | 否 | 作业名称，默认使用部署名称 |
| allow_non_restored_state | bool | 否 | 否 | 允许不恢复状态，默认: false |
| savepoint_path | string | 否 | 否 | 从指定保存点恢复 |
| operation_params | map(string) | 否 | 否 | 操作参数 |
| auto_restart | bool | 否 | 否 | 作业失败时自动重启，默认: true |

**返回值**:
| 属性名 | 类型 | 描述 |
|-------|------|------|
| id | string | 作业ID |
| job_id | string | Flink作业ID |
| status | string | 作业状态，如"RUNNING", "STOPPED", "FAILED" |
| start_time | string | 作业开始时间 |
| end_time | string | 作业结束时间 |
| last_savepoint_path | string | 最近保存点路径 |
| last_checkpoint_path | string | 最近检查点路径 |
| flink_web_ui_url | string | Flink Web UI URL |

## 数据源

### 1. alicloud_flink_workspaces

**描述**: 获取阿里云账户中的Flink工作空间列表。

**参数**:
| 参数名 | 类型 | 必选 | 描述 |
|-------|------|------|------|
| ids | list(string) | 否 | 按ID过滤工作空间 |
| name_regex | string | 否 | 按名称正则表达式过滤工作空间 |
| output_file | string | 否 | 保存结果的文件路径 |
| status | string | 否 | 按状态过滤，如"RUNNING", "CREATING" |
| resource_group_id | string | 否 | 按资源组ID过滤 |
| tags | map(string) | 否 | 按标签过滤 |

**返回值**:
| 属性名 | 类型 | 描述 |
|-------|------|------|
| ids | list(string) | 工作空间ID列表 |
| workspaces | list(object) | 工作空间对象列表 |
| workspaces[].id | string | 工作空间ID |
| workspaces[].name | string | 工作空间名称 |
| workspaces[].resource_group_id | string | 资源组ID |
| workspaces[].vpc_id | string | VPC ID |
| workspaces[].zone_id | string | 可用区ID |
| workspaces[].vswitch_ids | list(string) | 交换机ID列表 |
| workspaces[].status | string | 工作空间状态 |
| workspaces[].charge_type | string | 计费方式 |
| workspaces[].ha | bool | 是否启用高可用 |
| workspaces[].tags | map(string) | 标签映射 |
| workspaces[].create_time | string | 创建时间 |
| workspaces[].vpc_endpoint | string | VPC访问端点 |
| workspaces[].public_endpoint | string | 公网访问端点 |

### 2. alicloud_flink_namespaces

**描述**: 获取Flink工作空间中的命名空间列表。

**参数**:
| 参数名 | 类型 | 必选 | 描述 |
|-------|------|------|------|
| workspace_id | string | 是 | Flink工作空间ID |
| name_regex | string | 否 | 按名称正则表达式过滤命名空间 |
| output_file | string | 否 | 保存结果的文件路径 |
| status | string | 否 | 按状态过滤，如"ACTIVE" |

**返回值**:
| 属性名 | 类型 | 描述 |
|-------|------|------|
| ids | list(string) | 命名空间ID列表 |
| namespaces | list(object) | 命名空间对象列表 |
| namespaces[].id | string | 命名空间ID |
| namespaces[].name | string | 命名空间名称 |
| namespaces[].description | string | 命名空间描述 |
| namespaces[].workspace_id | string | 所属工作空间ID |
| namespaces[].create_time | string | 创建时间 |
| namespaces[].status | string | 命名空间状态 |
| namespaces[].members_count | number | 成员数量 |
| namespaces[].deployments_count | number | 部署数量 |

### 3. alicloud_flink_deployments

**描述**: 获取Flink命名空间中的部署列表。

**参数**:
| 参数名 | 类型 | 必选 | 描述 |
|-------|------|------|------|
| workspace_id | string | 是 | Flink工作空间ID |
| namespace | string | 是 | 命名空间名称 |
| name_regex | string | 否 | 按名称正则表达式过滤部署 |
| output_file | string | 否 | 保存结果的文件路径 |
| status | string | 否 | 按状态过滤 |

**返回值**:
| 属性名 | 类型 | 描述 |
|-------|------|------|
| ids | list(string) | 部署ID列表 |
| deployments | list(object) | 部署对象列表 |
| deployments[].id | string | 部署ID |
| deployments[].name | string | 部署名称 |
| deployments[].artifact_uri | string | 部署构件URI |
| deployments[].flink_configuration | map(string) | Flink配置映射 |
| deployments[].job_manager_resource_spec | object | JobManager资源规格 |
| deployments[].job_manager_resource_spec.cpu | number | JobManager CPU核数 |
| deployments[].job_manager_resource_spec.memory | string | JobManager内存 |
| deployments[].task_manager_resource_spec | object | TaskManager资源规格 |
| deployments[].task_manager_resource_spec.cpu | number | TaskManager CPU核数 |
| deployments[].task_manager_resource_spec.memory | string | TaskManager内存 |
| deployments[].environment_variables | map(string) | 环境变量映射 |
| deployments[].create_time | string | 创建时间 |
| deployments[].update_time | string | 最后更新时间 |
| deployments[].status | string | 部署状态 |
| deployments[].job_status | string | 关联作业状态 |
| deployments[].parallelism | number | 并行度 |
| deployments[].state_backend | string | 状态后端类型 |

### 4. alicloud_flink_zones

**描述**: 获取支持Flink服务的可用区信息。

**参数**:
| 参数名 | 类型 | 必选 | 描述 |
|-------|------|------|------|
| output_file | string | 否 | 保存结果的文件路径 |
| region | string | 否 | 区域ID，默认使用provider配置 |

**返回值**:
| 属性名 | 类型 | 描述 |
|-------|------|------|
| ids | list(string) | 可用区ID列表 |
| zones | list(object) | 可用区对象列表 |
| zones[].id | string | 可用区ID |
| zones[].name | string | 可用区名称 |
| zones[].available_resource_creation | set(string) | 可在该可用区创建的资源类型 |
| zones[].supported_instance_types | list(string) | 支持的实例类型 |
| zones[].supported_disk_categories | list(string) | 支持的磁盘类型 |
| zones[].architecture_types | list(string) | 支持的架构类型，如"X86", "ARM" |
| zones[].region | string | 所属区域ID |