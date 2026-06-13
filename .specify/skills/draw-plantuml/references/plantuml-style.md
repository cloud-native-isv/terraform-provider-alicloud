# PlantUML 标准样式配置

本文件定义 draw-plantuml 技能生成的所有 PlantUML 图表必须遵循的统一样式规范。在 Step 3 编写完 PlantUML 代码后，**必须**按照本配置对样式进行校验和调整。

## 基础样式模板（所有图表类型通用）

以下配置项必须插入在 `@startuml` 之后、图表内容之前：

```plantuml
@startuml
' === 布局方向 ===
top to bottom direction

' === 通用样式 ===
skinparam monochrome true
skinparam shadowing false
skinparam roundCorner 20

' === SVG 输出优化 ===
skinparam svgDimensionStyle false
skinparam svgLinkTarget _blank

' ... 图表内容 ...
@enduml
```

## 条件样式（按图表类型启用）

当图表包含 `actor` 或属于用例图（Use Case Diagram）时，在通用样式之后额外追加：

```plantuml
' === Actor 样式（仅含 actor/usecase 时启用） ===
skinparam actorStyle awesome
```

## 配置项说明

| 配置项 | 作用 | 适用范围 |
|--------|------|----------|
| `top to bottom direction` | 图的方向从上到下，保持阅读顺序一致 | 所有图表 |
| `skinparam monochrome true` | 黑白单色输出，适合文档和打印 | 所有图表 |
| `skinparam shadowing false` | 去除阴影效果，保持视觉简洁 | 所有图表 |
| `skinparam roundCorner 20` | 统一圆角半径为 20px | 所有图表 |
| `skinparam svgDimensionStyle false` | SVG 不内联 width/height，便于 CSS 控制尺寸 | 所有图表 |
| `skinparam svgLinkTarget _blank` | SVG 中的超链接在新窗口打开 | 所有图表 |
| `skinparam actorStyle awesome` | Actor 使用 FontAwesome 风格图标 | 仅用例图/含 actor 的图 |

## 样式校验要点

在完成 PlantUML 代码后，逐项检查：

1. **布局方向**：确认 `top to bottom direction` 在 `@startuml` 之后的第一行（注释除外）
2. **通用 skinparam**：确认 5 项通用 skinparam 全部存在且值正确
3. **条件 skinparam**：如图表含 actor 或为用例图，确认 `actorStyle awesome` 已添加
4. **位置**：所有样式配置必须在 `@startuml` 之后、图表元素定义之前
5. **无冲突**：确认图表内容中没有覆盖上述 skinparam 的重复声明