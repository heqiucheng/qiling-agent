# 企灵 Agent UI 设计系统规范

**状态**：Draft for implementation  
**版本**：v0.1  
**日期**：2026-05-28  
**适用范围**：Web 管理后台、销售工作台、后续企业微信侧边栏页面

## 1. 设计目标

企灵 Agent 的 UI 必须让用户一眼感到：

```text
可靠
专业
清晰
智能
可控
```

它不是营销页，不是炫技型 AI 页面，也不是杂乱 CRM。它是销售每天要使用的工作台，因此视觉必须服务效率、信任和长期使用。

核心目标：

- 一线销售能快速知道今天该跟谁、为什么跟、发什么。
- 销售主管能快速看到机会、风险和改进建议。
- AI 推荐必须显得可信，但不能显得神秘或不可控。
- 所有页面风格统一，禁止在页面里随意写局部样式。

## 2. 设计关键词

```text
企业级
克制
高密度
可信 AI
行动导向
可解释
可审计
```

不使用以下方向：

```text
花哨渐变
大面积发光背景
营销页式大 Hero
拟物装饰
过度圆角
卡片套卡片
大面积单色主题
```

## 3. AI 可信感设计

企灵的 AI 感不是靠炫酷效果，而是靠“结构化、可解释、可确认”。

### 3.1 AI 输出必须有三层信息

每个 AI 推荐都应显示：

```text
结论：建议做什么
理由：为什么这么判断
风险：哪里需要人工确认
```

示例：

```text
推荐动作：解释方案价值并引导预约
推荐理由：客户连续两次询问价格和售后，说明有购买兴趣但存在决策顾虑。
风险提示：涉及价格承诺，建议人工确认后发送。
```

### 3.2 AI 视觉组件

统一使用以下组件表达 AI 能力：

- `AgentInsight`：AI 洞察卡片。
- `AgentReasoning`：推荐理由区域。
- `AgentRiskFlag`：风险提示。
- `AgentConfidence`：可信度或置信提示。
- `AgentRunStatus`：分析中、已完成、失败、需确认。

AI 区域的视觉应使用低饱和科技感，不使用大面积霓虹或强渐变。

### 3.3 推荐可信度表达

推荐结果不要只显示“高/中/低”，还要说明依据。

```text
高意向：最近 2 次主动询问价格和交付，且 24 小时内有回复。
流失风险：超过 5 天无回复，且上次沟通停留在价格异议。
```

## 4. 统一样式硬性规则

以下规则是硬性要求，后续前端实现必须遵守。

### 4.1 禁止页面内写样式

禁止在页面文件中直接写：

```text
style="..."
内联 style 对象
随意 scoped 样式
页面级硬编码颜色
页面级硬编码字号
页面级硬编码间距
```

页面只能组合组件和传入语义化 props。

### 4.2 样式必须统一调用

所有样式必须来自统一设计系统：

```text
设计 Token
全局 CSS 变量
基础组件样式
布局组件
语义化状态组件
```

推荐结构：

```text
src/styles/tokens.css
src/styles/base.css
src/styles/layout.css
src/styles/components.css
src/styles/states.css
src/styles/themes.css
```

如果使用 Tailwind，也必须通过统一配置和组件封装使用，不能在页面中堆长 class 形成隐性样式分叉。

### 4.3 颜色只能使用语义 Token

禁止直接使用十六进制颜色：

```text
#1677ff
#ff4d4f
#111827
```

必须使用语义变量：

```css
var(--color-brand-primary)
var(--color-text-primary)
var(--color-risk-danger)
var(--color-intent-high)
```

### 4.4 组件优先

页面不得重复实现相同 UI。以下内容必须组件化：

```text
客户阶段标签
意向等级标签
风险提示
AI 推荐理由
话术卡片
跟进任务卡片
指标卡
空状态
上传状态
复制按钮
反馈按钮组
```

## 5. 设计 Token

### 5.1 品牌颜色

企灵 Agent 的主色不采用大面积蓝紫渐变。推荐使用“深青蓝 + 可信绿色 + 中性灰”的组合，既有 AI 感，也更稳重。

```css
:root {
  --color-brand-primary: #0f766e;
  --color-brand-primary-hover: #0d9488;
  --color-brand-primary-soft: #ccfbf1;

  --color-ai-accent: #2563eb;
  --color-ai-accent-soft: #dbeafe;

  --color-bg-page: #f6f8fb;
  --color-bg-surface: #ffffff;
  --color-bg-subtle: #f1f5f9;

  --color-text-primary: #0f172a;
  --color-text-secondary: #475569;
  --color-text-tertiary: #64748b;
  --color-text-inverse: #ffffff;

  --color-border-subtle: #e2e8f0;
  --color-border-strong: #cbd5e1;

  --color-success: #16a34a;
  --color-warning: #d97706;
  --color-danger: #dc2626;
  --color-info: #2563eb;
}
```

### 5.2 业务语义颜色

```css
:root {
  --color-intent-high: #dc2626;
  --color-intent-medium: #d97706;
  --color-intent-low: #64748b;

  --color-stage-new: #2563eb;
  --color-stage-discovery: #7c3aed;
  --color-stage-objection: #d97706;
  --color-stage-commit: #16a34a;
  --color-stage-silent: #64748b;
  --color-stage-risk: #dc2626;

  --color-task-pending: #2563eb;
  --color-task-done: #16a34a;
  --color-task-skipped: #64748b;
  --color-task-risk: #dc2626;
}
```

### 5.3 字体

```css
:root {
  --font-family-base: "Inter", "PingFang SC", "Microsoft YaHei", system-ui, sans-serif;
  --font-family-mono: "JetBrains Mono", "SFMono-Regular", Consolas, monospace;

  --font-size-xs: 12px;
  --font-size-sm: 13px;
  --font-size-md: 14px;
  --font-size-lg: 16px;
  --font-size-xl: 18px;
  --font-size-2xl: 22px;
  --font-size-3xl: 28px;

  --line-height-tight: 1.25;
  --line-height-normal: 1.5;
  --line-height-relaxed: 1.7;
}
```

页面内不允许使用超大字号。企灵是工作台，不是营销页。

### 5.4 间距

统一使用 4px 基准：

```css
:root {
  --space-1: 4px;
  --space-2: 8px;
  --space-3: 12px;
  --space-4: 16px;
  --space-5: 20px;
  --space-6: 24px;
  --space-8: 32px;
  --space-10: 40px;
  --space-12: 48px;
}
```

### 5.5 圆角和阴影

```css
:root {
  --radius-sm: 4px;
  --radius-md: 6px;
  --radius-lg: 8px;
  --radius-xl: 12px;

  --shadow-surface: 0 1px 2px rgba(15, 23, 42, 0.06);
  --shadow-popover: 0 8px 24px rgba(15, 23, 42, 0.12);
}
```

规则：

- 普通卡片最大 `8px` 圆角。
- 弹窗和浮层可使用 `12px` 圆角。
- 禁止大圆角糖果风。
- 阴影要轻，不能像营销落地页。

## 6. 布局系统

### 6.1 应用框架

默认后台结构：

```text
左侧导航
顶部上下文栏
主内容区
右侧可选详情/AI 理由抽屉
```

### 6.2 页面宽度

- 工作台和复盘中心：适合宽屏数据浏览。
- 客户详情：左侧聊天记录，右侧客户画像和 AI 推荐。
- 上传页面：居中主流程，但不做大面积营销式空白。

### 6.3 信息密度

企灵面向销售日常使用，信息密度应高于普通官网，但低于传统 ERP。

建议：

- 列表行高：48px - 56px。
- 卡片内间距：16px。
- 指标卡高度：96px - 120px。
- 详情页分栏比例：主内容 2/3，侧栏 1/3。

## 7. 核心组件规范

### 7.1 Button

按钮类型：

```text
Primary：主要动作，如复制话术、确认导入
Secondary：次要动作，如换一种、查看详情
Ghost：轻操作，如跳过、关闭
Danger：高风险动作，如删除、清空
```

话术任务中的主按钮必须是“复制话术”，不是“发送”。第一版不自动发送客户消息。

### 7.2 Badge

必须有统一组件：

```text
IntentBadge
StageBadge
RiskBadge
TaskStatusBadge
```

Badge 不能在页面里临时配颜色。

### 7.3 AgentInsight

用于展示 AI 判断：

```text
标题：AI 建议
结论：建议动作
理由：判断依据
风险：人工确认点
操作：复制 / 换一种 / 标记不准
```

### 7.4 ScriptCard

用于推荐话术：

```text
客户名
话术类型
推荐话术
推荐理由
风险提示
复制按钮
换一种按钮
标记不准
```

### 7.5 EmptyState

空状态必须引导下一步动作：

```text
暂无客户 -> 上传聊天记录 / 配置企业微信
暂无待跟进 -> 查看客户列表 / 生成复盘
暂无复盘数据 -> 导入更多聊天记录
```

## 8. 页面设计要求

### 8.1 工作台

第一屏优先级：

```text
今日待确认话术
高意向客户
风险客户
复盘摘要
```

禁止把工作台设计成纯数据大屏。它必须回答：今天先做什么。

### 8.2 客户列表

列表必须支持：

```text
搜索
阶段筛选
意向筛选
风险筛选
负责人筛选
```

每一行必须展示足够决策信息，不要求用户进入详情才知道客户状态。

### 8.3 客户详情

推荐布局：

```text
左侧：聊天记录时间线
中间：客户资料和跟进历史
右侧：客户画像、AI 推荐、风险提示
```

如果屏幕较窄，右侧 AI 推荐可以折叠为抽屉。

### 8.4 数据接入

必须让用户清楚看到两条路径：

```text
企业微信接入
上传/粘贴聊天记录
```

企业微信接入未完成时，页面不能成为死路，必须引导用户先上传聊天记录体验 Agent。

### 8.5 复盘中心

复盘中心优先展示：

```text
最重要机会
最大风险
需要主管介入
话术效果
客户阶段分布
改进建议
```

禁止只展示漂亮指标。每个洞察必须能落到行动。

## 9. 状态设计

必须统一设计以下状态：

```text
加载中
AI 分析中
上传中
解析中
解析失败
空数据
复制成功
已跳过
已标记不准
同步失败
权限不足
```

AI 分析中状态必须说明当前步骤，例如：

```text
正在解析聊天记录
正在识别客户阶段
正在生成推荐话术
正在检查风险提示
```

## 10. 可访问性和可读性

- 正文字号不低于 14px。
- 说明文字不低于 12px。
- 颜色对比至少满足常规后台可读性要求。
- 所有按钮有明确文本或可访问名称。
- 风险提示不能只靠颜色表达，必须有文字。
- 表格和列表需要清晰 hover、selected、disabled 状态。

## 11. 前端实现约束

实现前必须创建统一样式入口。

推荐：

```text
src/styles/tokens.css
src/styles/base.css
src/styles/layout.css
src/styles/components.css
src/styles/states.css
src/styles/themes.css
```

推荐组件目录：

```text
src/components/ui/
src/components/agent/
src/components/customer/
src/components/review/
src/components/upload/
```

禁止：

- 页面内直接写复杂样式。
- 每个页面自己定义按钮颜色。
- 每个页面自己定义 Badge。
- 页面里散落 AI 状态文案。
- 页面里硬编码业务颜色。

## 12. 后续后端和 AI 设计提醒

以下内容不是本 UI 文档的详细设计范围，但后续进入后端和 AI 架构时必须单独讲清楚：

- 向量库选型和数据分层。
- 长期记忆：客户画像、历史沟通、成交偏好、销售反馈。
- 短期记忆：当前会话上下文、最近 N 条消息、当前任务状态。
- RAG 召回策略：按客户、产品、销售阶段、话术类型召回。
- 召回率和准确率评估。
- Agent 执行记录和可追溯性。
- Prompt 版本管理。
- AI 输出结构化校验。
- 反馈闭环：采纳、修改、跳过、标记不准如何反哺规则。
- 敏感信息和权限隔离。

在实现到这些阶段时，必须向用户说明：

```text
用了什么技术点
为什么这么选
难点是什么
风险在哪里
如何测试和验证
```