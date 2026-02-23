# TD MVP-A Interaction Spec

## Overview

MVP-A 使用单页全屏 TUI（htop 风格）：

- Header：左上角 `TD` ASCII 标识 + 今日进度条
- Body：左栏导航（1/4）+ 右栏任务列表（3/4）
- Footer：状态消息 + 快捷键提示

CLI 与 TUI 共享同一应用层与 sqlite 数据源。

## Left Nav Views

- 今日（Today）
- Inbox
- 日志（Log）
- 项目（Project）
- 回收站（Trash）

### Query Rules

- Inbox：`status = inbox`
- 今日：`status in (todo, doing)` 且满足 `doing` 或 `overdue` 或 `due today`
- 日志：`status = done` 且 `done_at >= now - 14d`
- 项目：`project = X` 且 `status in (inbox, todo, doing)`
- 回收站：`status = deleted`

## Right List

- 显示当前视图任务
- 支持上下移动游标
- 随左栏视图切换即时刷新

## Header Progress

- 进度标题：`Today Progress`
- 口径：沿用 Today 规则集合
- 分母：满足 Today 条件的任务总数（`doing` + `overdue/due today`）
- 分子：上述集合中 `status=done` 的数量
- 条形：`[#...-...]` + `done/total + percent`

## Clipboard Flow

- `--clip` / `p`：读取剪贴板，规则解析后创建任务
- `--clip --ai` / `Ctrl+A`：先 AI 结构化解析，schema 校验通过则使用 AI 结果；否则回退规则解析
