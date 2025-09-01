# ARMS Alert Event Service Layer Implementation

## 任务概述
实现 terraform-provider-alicloud 中针对 ARMS 告警事件系统的 Service 层封装，包括 AlertItem、AlertEvent、AlertAlarm、AlertActivity 等 API 的服务层函数。

## 实现计划

### 阶段 1: Service 层实现 ✓
- [ ] 实现 AlertItem 相关服务函数
  - [ ] DescribeArmsAlertItems - 获取告警项列表
  - [ ] DescribeArmsAlertItemById - 根据ID获取单个告警项
- [ ] 实现 AlertEvent 相关服务函数  
  - [ ] DescribeArmsAlertEvents - 获取告警事件列表（带过滤条件）
  - [ ] DescribeArmsAlertEventById - 根据ID获取单个告警事件
  - [ ] DescribeArmsActiveAlertEvents - 获取活跃告警事件
  - [ ] DescribeArmsAlertEventsByTimeRange - 根据时间范围获取告警事件
- [ ] 实现 AlertActivity 相关服务函数
  - [ ] DescribeArmsAlertActivities - 从AlertEvent中提取Activities
  - [ ] DescribeArmsAlertActivitiesByAlertId - 根据AlertId获取Activities
- [ ] 实现 AlertAlarm 相关服务函数
  - [ ] DescribeArmsAlertAlarms - 从AlertEvent中提取Alarms
  - [ ] DescribeArmsAlertAlarmsByAlertId - 根据AlertId获取Alarms
- [ ] 实现 ID 编码/解码函数
  - [ ] EncodeArmsAlertEventId / DecodeArmsAlertEventId
  - [ ] EncodeArmsAlertItemId / DecodeArmsAlertItemId
  - [ ] EncodeArmsAlertActivityId / DecodeArmsAlertActivityId
  - [ ] EncodeArmsAlertAlarmId / DecodeArmsAlertAlarmId

### 阶段 2: DataSource 更新 ✓
- [ ] 更新 data_source_alicloud_arms_alert_activities.go
  - [ ] 替换直接API调用为Service层调用
  - [ ] 使用 DescribeArmsAlertActivitiesByAlertId 方法
  - [ ] 保持现有的过滤逻辑和返回格式

### 阶段 3: 测试验证 ✓
- [ ] 编译验证
- [ ] 语法检查
- [ ] 确保API调用正确

## 技术要点

### API 调用层级关系
- AlertItem: 通过 armsAPI.ListAllAlertItems() 获取
- AlertEvent: 通过 armsAPI.ListAllAlertEvents() 获取，包含 Activities 和 Alarms
- AlertActivity: 从 AlertEvent.Activities 中提取
- AlertAlarm: 从 AlertEvent.Alarms 中提取

### ID 编码策略
- AlertItem: 直接使用 AlertId
- AlertEvent: 使用 EventId 
- AlertActivity: 使用 "alertId_eventId_activityId" 格式
- AlertAlarm: 使用 "alertId_eventId_alarmId" 格式

### 过滤逻辑
- 在Service层实现过滤逻辑
- 支持按alertId、时间范围、严重级别等过滤
- 保持与原有DataSource兼容的接口
