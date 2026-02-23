# td

本项目是 **TD MVP-A**：本地优先的终端 Todo 工具，提供 CLI 与双栏 TUI。

## 功能

- 任务命令：`add / ls / show / done / reopen / edit / rm / restore / purge`
- TUI：左侧导航（今日 / Inbox / 日志 / 项目 / 回收站）+ 右侧列表
- 剪贴板：`--clip` 规则解析
- AI 解析：`--clip --ai`（schema 校验失败自动回退规则解析）
- 数据存储：SQLite（本地）

## 状态机

- `inbox`
- `todo`
- `doing`
- `done`
- `deleted`

## 安装与运行

### 方式 1：下载 Release 制品（推荐）

在 [Releases](https://github.com/roobtyan/td/releases) 下载对应平台二进制：

- macOS Intel: `td-darwin-amd64`
- macOS Apple Silicon: `td-darwin-arm64`
- Linux x86_64: `td-linux-amd64`
- Linux ARM64: `td-linux-arm64`

下载后加执行权限并放入 PATH：

```bash
chmod +x td-*
mv td-darwin-arm64 /usr/local/bin/td
```

### 方式 2：源码运行

```bash
go test ./... -v
go run ./cmd/td --help
go run ./cmd/td add "buy milk"
go run ./cmd/td ls
go run ./cmd/td ui
```

## 常用命令

```bash
td add "buy milk"
td add --clip
td add --clip --ai
td ls
td show 1
td done 1
td reopen 1
td rm 1
td restore 1
td purge 1
td ui
```

## 配置与数据目录

- 默认目录：`$HOME/.td`
- 可通过 `TD_HOME` 覆盖
- 数据库文件：`$TD_HOME/data/td.db`

## MVP-A 边界

- 包含：CLI 核心命令、双栏 TUI、剪贴板规则解析、AI 解析回退
- 不包含：`archived`、triage、expand

## 隐私说明

- 数据默认仅保存在本地 SQLite。
- 仅在显式使用 `--clip --ai` 时才触发 AI 解析流程。
