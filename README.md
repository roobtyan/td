# td

本项目是 TD MVP-A：本地优先的终端 Todo 工具，包含 CLI 与双栏 TUI。

## MVP-A Scope

- 包含：`add/ls/show/done/reopen/edit/rm/restore/purge/ui`
- 包含：`--clip`、`--clip --ai`（AI 失败自动回退规则解析）
- 包含：左栏固定导航（今日 / Inbox / 日志 / 项目 / 回收站）
- 不包含：`archived`、triage、expand

状态机固定：

- `inbox`
- `todo`
- `doing`
- `done`
- `deleted`

## Quick Start

```bash
go test ./... -v
go run ./cmd/td --help
go run ./cmd/td add "buy milk"
go run ./cmd/td ls
go run ./cmd/td ui
```

## Commands

- `td add <text>`
- `td add --clip`
- `td add --clip --ai`
- `td ls`
- `td show <id>`
- `td done <id...>`
- `td reopen <id...>`
- `td edit <id> <title>`
- `td rm <id...>`
- `td restore <id...>`
- `td purge <id...>`
- `td ui`

## Privacy

- 数据默认落地到本地 `TD_HOME` 目录下 sqlite 文件。
- `--clip --ai` 仅在开启时调用 AI 解析；失败会自动回退到规则解析。
