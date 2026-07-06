---
name: draw-plantuml
description: |
  Draw system architecture diagrams with PlantUML, render to SVG/PNG via PlantUML server, and output as HTML with rendered images.
  Use standard UML semantics (Component, Deployment, Sequence, Class/Package) to describe system architecture.
  Also supports five non-UML specialty diagrams: WBS (工作分解结构), Gantt (甘特图), MindMap (思维导图), JSON 数据可视化, YAML 显示效果图.
  Use when the user mentions "架构图", "architecture diagram", "UML图", "plantuml", "系统架构图", "画架构", "设计图", "组件图", "部署图", "时序图", "类图", "包图", "系统设计",
  "流程图", "状态图", "活动图", "用例图", "状态机图", "模块图", "交互图",
  "sequence diagram", "class diagram", "component diagram", "deployment diagram",
  "activity diagram", "state diagram", "use case diagram", "package diagram",
  "工作分解结构", "WBS", "甘特图", "gantt", "项目计划图", "进度图", "思维导图", "mindmap", "脑图",
  "JSON可视化", "JSON数据图", "json diagram", "YAML可视化", "YAML显示", "yaml diagram", "配置可视化", "数据结构图",
  "复刻图", "图片重绘", "图片转UML", "replicate diagram", "redraw", "image to UML"
skill_id: "<SKILL:.specify/skills/draw-plantuml/SKILL.md>"
---

# 架构图绘制技能

使用 PlantUML 语法和标准 UML 语义绘制系统架构图，通过 PlantUML 服务器渲染为 SVG/PNG，并输出为包含渲染图表和说明文字的完整 HTML 文档。

## 核心原则

- **UML 语义，而非随意方框**：UML 类图表必须遵循标准 UML 图表类型，使用正确的 UML 元素和关系
- **架构优先的叙事**：图和文字互补——文字解释*为什么*，图展示*什么*
- **统一样式**：使用 `skinparam` / `<style>` 保持统一样式，UML 图每张核心元素 ≤7 个（硬上限 ≤15）
- **专项图表遵循其原生语义**：WBS/甘特图/思维导图/JSON/YAML 五类非 UML 图表使用各自的原生语法（`@startwbs`/`@startgantt`/`@startmindmap`/`@startjson`/`@startyaml`）与原生配色，不套用 UML 的 skinparam 单色规则

## 工作流

按以下 8 个步骤顺序执行。每一步的核心说明如下，详细操作请阅读对应的参考文档。

### Step 1: 语义解析

分析用户输入以理解绘制意图。根据输入的完整性，通过补充推断或交互式提问（`AskUserQuestion`，最多一轮 ≤4 个问题）向用户确认，确保绘制意图完全明确后再进入下一步。

→ 详细方法参见 [00-semantic-analysis.md](references/howto/00-semantic-analysis.md)

### Step 2: 选择正确的图表类型

根据用户描述的系统特征和要表达的架构视角，从 8 种标准 UML 图表类型中选择最合适的一种或多种。每张图聚焦单一视角。

→ 选择方法参见 [01-choose-diagram-type.md](references/howto/01-choose-diagram-type.md)

### Step 3: 选择合适的图表元素

确定图表类型后，阅读对应的操作指南，选择正确的 UML 元素（组件、节点、生命线、类、状态等）和关系类型（依赖、关联、实现等）。

→ 各图表类型操作指南见 [references/howto/](references/howto/) 目录（02–09）

### Step 4: 规划图表的整体布局

在编写代码之前，分析组件间的语义关系以确定自然位置。识别组件角色（Hub/Edge/Peer/Entry/Sink/External），根据关系模式规划布局，先画位置草图再编写代码。

→ 布局规划方法参见 [10-layout-planning.md](references/howto/10-layout-planning.md)，基础语义布局规则参见 [layout.md §一](references/guide/layout.md)

### Step 5: 阅读最佳实践

在生成代码之前，阅读最佳实践文档，了解布局优化、内容组织、标签精简（≤10 字符 + 富文本注释）、视觉高亮和按图表类型的布局指南等需要注意的事项。

→ 参见 [layout.md](references/guide/layout.md) 和 [content.md](references/guide/content.md)

### Step 6: 生成 PlantUML 代码

根据所选图表类型的操作指南和最佳实践，编写具体的 PlantUML 代码。用 `@startuml`/`@enduml` 包裹，先声明元素再声明关系，应用方向关键字和分组控制布局。

→ 代码生成指南参见 [11-code-generation.md](references/howto/11-code-generation.md)，语法参考参见 [syntax-reference.md](references/guide/syntax-reference.md)

### Step 7: 应用标准样式

代码生成后，根据样式文档应用统一的标准样式配置（skinparam、布局方向、色彩模式等），确保视觉一致性。

→ 样式配置参见 [style.md](references/guide/style.md)

### Step 8: 渲染、匹配与微调

使用渲染脚本将 PlantUML 代码渲染为 SVG/PNG 图片。读取生成的图片，与最初用户输入的要求进行匹配比对，发现差异时微调代码并重新渲染，最终组装为 HTML 文档输出。

→ 渲染、验证和输出指南参见 [12-rendering-and-output.md](references/howto/12-rendering-and-output.md)

## 专项图表（非 UML）

除 8 种标准 UML 图表外，本技能还支持 5 种专项图表。它们不遵循 UML 语义，各自有独立语法与原生配色。当用户意图属于以下场景时，在 Step 2 直接选用对应专项图表，并阅读其操作指南：

| 专项图表 | 适用场景 | 起止标记 | 操作指南 |
|---------|---------|---------|---------|
| **WBS 工作分解结构** | 项目/交付物层级分解 | `@startwbs`/`@endwbs` | [13-wbs-diagram.md](references/howto/13-wbs-diagram.md) |
| **甘特图 Gantt** | 项目进度、任务依赖、里程碑 | `@startgantt`/`@endgantt` | [14-gantt-diagram.md](references/howto/14-gantt-diagram.md) |
| **思维导图 MindMap** | 知识梳理、发散规划 | `@startmindmap`/`@endmindmap` | [15-mindmap-diagram.md](references/howto/15-mindmap-diagram.md) |
| **JSON 数据可视化** | 展示 JSON 数据结构 | `@startjson`/`@endjson` | [16-json-diagram.md](references/howto/16-json-diagram.md) |
| **YAML 显示效果图** | 展示 YAML 配置结构 | `@startyaml`/`@endyaml` | [17-yaml-diagram.md](references/howto/17-yaml-diagram.md) |

> 专项图表的渲染同样走 Step 8 的渲染脚本；无需 Graphviz（`dot`）即可渲染。样式与美观要点见各操作指南的「布局与美观技巧」小节。

## 输出要求

- 输出为单个 HTML 文档，包含渲染的 SVG/PNG 图表（不嵌入原始 PlantUML 文本）
- 图表通过 [render-plantuml.sh](scripts/render-plantuml.sh) 渲染
- SVG/PNG 与 HTML 保存在同一目录，HTML 通过相对路径引用图片
- PlantUML 源文件（`.puml`）保存以供未来编辑
- 每张图至少包含标题、渲染图片和简要说明

## 参考文档

所有参考文档（操作指南、最佳实践、官方文档）的完整索引和说明，参见 [references/README.md](references/README.md)。
