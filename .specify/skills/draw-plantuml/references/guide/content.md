# PlantUML 内容组织指南

通过规范代码结构、合理控制内容、统一标签和注释规则，使生成的 UML 图逻辑清晰、层次分明且认知友好。

> **定位**：本文件是绘图**内容质量指南**，与 [style.md](./style.md)（统一样式配置）和 [layout.md](./layout.md)（布局优化技巧）互补。在 Step 6 生成 PlantUML 代码时**必须参照本指南**。

---

## 一、内容组织与认知控制

### 1.1 单一职责原则

**单图聚焦一个主题**，避免信息过载：
- 架构图：仅展示模块边界与依赖，不包含方法细节
- 时序图：仅描述一次典型交互，异常流程单独成图
- 类图：按职责域拆分，不要把所有类画进一张图

### 1.2 元素数量控制

| 层级 | 核心元素上限 | 处理方式 |
|------|------------|---------|
| **认知友好** | ≤7 个核心元素 | 一眼可理解，首选 |
| **可接受** | 8-12 个元素 | 需要辅以分组和标注 |
| **需拆分** | >12 个元素 | 拆分为概览图 + 详细子图 |
| **硬上限** | 15 个元素 | 绝对不超过，否则渲染质量和可读性都不可控 |

### 1.3 C4 层级化拆分

复杂系统按 C4 模型分层绘制，每层一张图：

1. **系统上下文图 (Context)** — 系统与外部用户/系统的关系
2. **容器图 (Container)** — 系统内部的主要技术容器（服务、数据库等）
3. **组件图 (Component)** — 单个容器内部的逻辑组件
4. **类图 (Class)** — 单个组件内部的代码结构

跨图引用时在子图标题中注明：
```plantuml
title 支付服务内部组件（容器图见 01-system-containers）
```

### 1.4 代码结构映射布局

**按逻辑顺序编写元素**，先定义核心参与者/类，再描述关系：

```plantuml
' ✓ 好：先声明核心元素，再描述关系
participant OrderService
participant PaymentService
participant InventoryService

OrderService -> PaymentService : 发起支付
PaymentService -> InventoryService : 扣减库存
```

```plantuml
' ✗ 差：边定义边画关系，阅读混乱
participant OrderService
OrderService -> PaymentService : 发起支付
participant InventoryService
PaymentService -> InventoryService : 扣减库存
```

### 1.5 标签精简与富文本注释补充

> **组件文本描述能力有限时，用富文本注释补充。** 这是一条适用于所有图类型的通用规则，在 Step 6 生成代码时**必须遵守**。

#### 核心规则：标签 ≤10 字符 + note 补充

元素名称/标签不超过 10 个字符。当短标签无法充分描述元素职责时，使用 `note` 元素补充详细说明。

**为什么要限制标签长度**：过长的元素标签会导致框体过宽、自动换行不可预测、认知负荷增加。

```plantuml
' ✗ 差：元素标签过长，导致框体过宽
component [用户订单管理服务主模块] as OrderMain
participant "用户认证与授权服务中心" as Auth

' ✓ 好：简短标签 + note 补充说明
component [订单服务] as Order
participant "认证中心" as Auth

note right of Order
  包含订单创建、取消、
  退款、状态查询等子模块
end note
```

#### 关系线标签同样适用

```plantuml
' ✗ 差
A -> B : 发送异步HTTP请求并携带JWT令牌

' ✓ 好
A -> B : 异步调用
note on link
  HTTP POST + JWT Bearer
end note
```

#### note 语法速查

| 场景 | 语法 | 说明 |
|------|------|------|
| 元素旁注释 | `note right of X` / `note left of` / `note top of` / `note bottom of` | 定位在元素指定方向 |
| 多行注释 | `note right of X` ... `end note` | 多段说明文字 |
| 浮动注释 | `note "text" as N` + `N .. X` | 独立放置，用虚线连接 |
| 关系线注释 | `note on link` ... `end note` | 附加在箭头连线上 |
| 简短箭头标签 | `A -> B : 简短文字` | ≤10 字符，直接在箭头上 |

> **详细注释模式和示例见 [§二–三](#二注释策略)。**

---

## 二、注释策略

> 核心规则见 [§1.5 标签精简与富文本注释补充](#一内容组织与认知控制)。本节提供注释写作的详细模式和示例。

注释紧贴关联元素，精简高效：

```plantuml
' ✓ 好：精简且位置明确
note right of AuthService
  Token TTL < 2h
  刷新间隔 30min
end note

' ✓ 好：关系线上的注释
note on link
  HTTP/2 + TLS 1.3
end note

' ✗ 差：长段落注释
note right of AuthService
  这个服务负责处理所有的认证逻辑，
  包括但不限于：Token 生成、Token 验证、
  Token 刷新、密码重置等一系列操作...
end note
```

---

## 三、元素标签长度控制（≤ 10 字符）

> 核心规则和示例见 [§1.5 标签精简与富文本注释补充](#一内容组织与认知控制)。本节补充常见反模式。

**核心规则**：元素名称/标签不超过 10 个字符。超过时替换为更简短的描述，必要说明通过 `note` 元素补充。

过长的元素标签会导致：
- 元素框体过宽，破坏布局平衡
- 自动换行产生不可预测的渲染结果
- 认知负荷增加，读者难以快速扫描

```plantuml
' ✗ 差：元素标签过长，导致框体过宽
component [用户订单管理服务主模块] as OrderMain
participant "用户认证与授权服务中心" as Auth
A -> B : 发送订单创建请求并等待确认

' ✓ 好：简短标签 + note 补充说明
component [订单服务] as Order
participant "认证中心" as Auth
A -> B : 创建订单

note right of Order
  包含订单创建、取消、
  退款、状态查询等子模块
end note
```

**特别注意关系线标签**：关系线上的文本同样遵守 ≤10 字符规则，超过时用 `note on link` 补充：

```plantuml
' ✗ 差
A -> B : 发送异步HTTP请求并携带JWT令牌

' ✓ 好
A -> B : 异步调用
note on link
  HTTP POST + JWT Bearer
end note
```

---

## 四、别名与标签规范

- **别名必须可读**：`[OrderService] as OS` 优于 `[C1]`
- **关系标签必须描述交互方式**：`A --> B : HTTP调用` 优于 `A --> B : uses`
- **基数标记显式化**：`"1" --> "0..*"` 明确关联语义

---

## 五、协作与维护规范

### 5.1 文件命名

使用 `{nn}-{short-title}.puml` 格式，例如：
- `01-system-overview.puml`
- `02-order-flow-sequence.puml`
- `03-auth-component.puml`

### 5.2 元信息注释

在图表开头（`@startuml` 之后、样式之前）标注上下文：

```plantuml
@startuml
' 描述: 订单创建主流程（不含异常分支）
' 关联: 容器图见 01-system-containers.puml

' === 样式 ===
skinparam monochrome true
...
```

### 5.3 避免硬编码路径

样式复用时使用相对路径：
```plantuml
!include ../styles/base-style.puml
```

### 5.4 版本控制友好

- 每行一个元素或关系声明，便于 diff 对比
- 空行分隔逻辑区块（元素定义 / 关系 / 注释）
- 别名声明与关系声明分开，不要混写

---

## 六、质量自检清单

在交付前逐项确认：

- [ ] 单图核心元素 ≤7（可接受 ≤12，硬上限 15）
- [ ] **所有元素名称和关系标签 ≤10 字符；超过时用 note 补充说明**
- [ ] 每个元素至少有一条关系（无孤立元素）
- [ ] 关系标签描述了交互方式（不是泛泛的 "uses"）
- [ ] 元素声明在前，关系声明在后（代码结构清晰）
- [ ] 逻辑相关元素通过分组或方向指令保持邻近
- [ ] 图表聚焦单一主题（不混合架构/流程/数据模型）
- [ ] 复杂系统已按 C4 层级拆分为多图
- [ ] 文件命名遵循 `{nn}-{short-title}.puml` 规范
- [ ] 跨图引用在 title 中注明

---

## 扩展阅读

- **布局优化技巧**：参见 [layout.md](./layout.md)
- **统一样式配置**：参见 [style.md](./style.md)
- **PlantUML 语法参考**：参见 [syntax-reference.md](./syntax-reference.md)
