# TD MVP-A Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 交付一个本地优先的终端 Todo MVP（A 版），包含 CLI、SQLite、左 1/4 导航 + 右 3/4 列表 TUI、剪贴板规则解析与 AI 解析（失败可回退）。

**Architecture:** 采用分层结构 `domain -> app(usecase) -> repo(sqlite)`，CLI 与 TUI 作为两个入口复用同一应用层。AI 只通过 `internal/ai` 接口接入，所有 AI 写入前先走 schema 校验，失败时回退到规则解析，保证核心流程稳定可用。

**Tech Stack:** Go 1.23+、Cobra、Bubble Tea/Bubbles/Lipgloss、modernc.org/sqlite、atotto/clipboard（或 golang.design/x/clipboard）、OpenAI-compatible REST。

## Scope Lock (MVP-A)

- In: `add/ls/show/done/reopen/edit/rm/restore/purge/ui`、`--clip`、`--clip --ai`、左导航（今日/Inbox/日志/项目/回收站）。
- Out: `archived` 状态、triage 模式、expand 模式、复杂详情抽屉。
- 状态机固定为：`inbox | todo | doing | done | deleted`。

## Execution Rules

- 严格按 `@superpowers:test-driven-development`：先写失败测试，再最小实现，再回归。
- 每个 Task 完成后小步提交，提交前执行对应测试命令。
- 严格 YAGNI：仅实现 MVP-A 范围，不提前实现 triage/expand。

### Task 1: 项目脚手架与基础入口

**Files:**
- Create: `go.mod`
- Create: `cmd/td/main.go`
- Create: `internal/app/app.go`
- Create: `internal/domain/task.go`
- Create: `internal/config/config.go`
- Create: `internal/timeutil/clock.go`
- Create: `Makefile`
- Create: `.gitignore`
- Test: `cmd/td/main_test.go`

**Step 1: 写失败测试（CLI 可启动）**

```go
func TestMainHelp(t *testing.T) {
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"--help"})
	err := cmd.Execute()
	require.NoError(t, err)
}
```

**Step 2: 运行测试确认失败**

Run: `go test ./cmd/td -run TestMainHelp -v`  
Expected: FAIL（`NewRootCmd` 未定义）

**Step 3: 最小实现 root command**

```go
func NewRootCmd() *cobra.Command {
	return &cobra.Command{Use: "td"}
}
```

**Step 4: 运行测试确认通过**

Run: `go test ./cmd/td -run TestMainHelp -v`  
Expected: PASS

**Step 5: Commit**

```bash
git add go.mod cmd/td/main.go cmd/td/main_test.go internal/app/app.go internal/domain/task.go internal/config/config.go internal/timeutil/clock.go Makefile .gitignore
git commit -m "chore: bootstrap td project skeleton"
```

### Task 2: SQLite schema + migration + 状态约束

**Files:**
- Create: `internal/repo/sqlite/migrations/0001_init.sql`
- Create: `internal/repo/sqlite/migrate.go`
- Create: `internal/repo/sqlite/db.go`
- Create: `internal/repo/sqlite/schema_test.go`

**Step 1: 写失败测试（建表、索引、状态约束）**

```go
func TestSchemaInit(t *testing.T) {
	db := openTestDB(t)
	require.NoError(t, Migrate(db))
	require.True(t, tableExists(t, db, "tasks"))
	require.True(t, indexExists(t, db, "idx_tasks_status"))
}
```

**Step 2: 运行测试确认失败**

Run: `go test ./internal/repo/sqlite -run TestSchemaInit -v`  
Expected: FAIL（迁移尚未实现）

**Step 3: 最小实现 migration**

```sql
CREATE TABLE IF NOT EXISTS schema_migrations(version INTEGER PRIMARY KEY);
CREATE TABLE IF NOT EXISTS tasks (...);
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_tasks_project ON tasks(project);
CREATE INDEX IF NOT EXISTS idx_tasks_due_at ON tasks(due_at);
CREATE INDEX IF NOT EXISTS idx_tasks_updated_at ON tasks(updated_at);
CREATE INDEX IF NOT EXISTS idx_tasks_done_at ON tasks(done_at);
```

**Step 4: 运行测试确认通过**

Run: `go test ./internal/repo/sqlite -run TestSchemaInit -v`  
Expected: PASS

**Step 5: Commit**

```bash
git add internal/repo/sqlite/migrations/0001_init.sql internal/repo/sqlite/migrate.go internal/repo/sqlite/db.go internal/repo/sqlite/schema_test.go
git commit -m "feat: add sqlite schema and migration baseline"
```

### Task 3: Domain + Repository CRUD（含状态流转）

**Files:**
- Create: `internal/domain/status.go`
- Create: `internal/domain/errors.go`
- Create: `internal/repo/repository.go`
- Create: `internal/repo/sqlite/task_repo.go`
- Test: `internal/repo/sqlite/task_repo_test.go`

**Step 1: 写失败测试（新增、查询、状态流转）**

```go
func TestTaskLifecycle(t *testing.T) {
	id := createTask(t, repo, "write plan")
	require.NoError(t, repo.MarkDone(ctx, []int64{id}))
	task := mustGetTask(t, repo, id)
	require.Equal(t, domain.StatusDone, task.Status)
}
```

**Step 2: 运行测试确认失败**

Run: `go test ./internal/repo/sqlite -run TestTaskLifecycle -v`  
Expected: FAIL（repo 方法缺失）

**Step 3: 最小实现 CRUD + 状态转移检查**

```go
func CanTransit(from, to Status) bool { ... } // 仅允许 MVP-A 定义流转
```

**Step 4: 运行测试确认通过**

Run: `go test ./internal/repo/sqlite -run TestTaskLifecycle -v`  
Expected: PASS

**Step 5: Commit**

```bash
git add internal/domain/status.go internal/domain/errors.go internal/repo/repository.go internal/repo/sqlite/task_repo.go internal/repo/sqlite/task_repo_test.go
git commit -m "feat: implement task repository and state transitions"
```

### Task 4: CLI 命令第一批（add/ls/show）

**Files:**
- Create: `internal/cli/root.go`
- Create: `internal/cli/add.go`
- Create: `internal/cli/ls.go`
- Create: `internal/cli/show.go`
- Create: `internal/app/usecase/add_task.go`
- Create: `internal/app/usecase/list_task.go`
- Test: `internal/cli/add_test.go`
- Test: `internal/cli/ls_test.go`

**Step 1: 写失败测试（add + ls + show）**

```go
func TestAddAndList(t *testing.T) {
	out := runCLI(t, "add", "buy milk")
	require.Contains(t, out, "created")
	out = runCLI(t, "ls")
	require.Contains(t, out, "buy milk")
}
```

**Step 2: 运行测试确认失败**

Run: `go test ./internal/cli -run TestAddAndList -v`  
Expected: FAIL（命令未注册）

**Step 3: 最小实现命令与 usecase**

```go
cmd.Flags().StringP("project", "p", "", "project")
cmd.Flags().StringP("priority", "P", "P2", "priority")
```

**Step 4: 运行测试确认通过**

Run: `go test ./internal/cli -run TestAddAndList -v`  
Expected: PASS

**Step 5: Commit**

```bash
git add internal/cli/root.go internal/cli/add.go internal/cli/ls.go internal/cli/show.go internal/app/usecase/add_task.go internal/app/usecase/list_task.go internal/cli/add_test.go internal/cli/ls_test.go
git commit -m "feat: add basic CLI commands add ls show"
```

### Task 5: CLI 命令第二批（done/reopen/edit/rm/restore/purge）

**Files:**
- Create: `internal/cli/done.go`
- Create: `internal/cli/reopen.go`
- Create: `internal/cli/edit.go`
- Create: `internal/cli/rm.go`
- Create: `internal/cli/restore.go`
- Create: `internal/cli/purge.go`
- Create: `internal/app/usecase/update_task.go`
- Test: `internal/cli/lifecycle_test.go`

**Step 1: 写失败测试（完整生命周期）**

```go
func TestLifecycleCommands(t *testing.T) {
	id := createViaCLI(t, "add", "task A")
	runCLI(t, "done", id)
	runCLI(t, "reopen", id)
	runCLI(t, "rm", id)
	runCLI(t, "restore", id)
}
```

**Step 2: 运行测试确认失败**

Run: `go test ./internal/cli -run TestLifecycleCommands -v`  
Expected: FAIL

**Step 3: 最小实现生命周期命令**

```go
Use: "rm <id...>" // 仅软删除到 deleted
Use: "purge <id...>" // 仅允许 deleted 状态
```

**Step 4: 运行测试确认通过**

Run: `go test ./internal/cli -run TestLifecycleCommands -v`  
Expected: PASS

**Step 5: Commit**

```bash
git add internal/cli/done.go internal/cli/reopen.go internal/cli/edit.go internal/cli/rm.go internal/cli/restore.go internal/cli/purge.go internal/app/usecase/update_task.go internal/cli/lifecycle_test.go
git commit -m "feat: implement task lifecycle CLI commands"
```

### Task 6: 查询集合（今日/Inbox/日志/项目/回收站）

**Files:**
- Create: `internal/domain/query.go`
- Create: `internal/app/usecase/nav_query.go`
- Modify: `internal/repo/sqlite/task_repo.go`
- Test: `internal/app/usecase/nav_query_test.go`

**Step 1: 写失败测试（五个集合）**

```go
func TestTodayView(t *testing.T) {
	seedTodayData(t, repo)
	tasks, err := uc.ListByView(ctx, app.ViewToday, now)
	require.NoError(t, err)
	require.NotEmpty(t, tasks)
}
```

**Step 2: 运行测试确认失败**

Run: `go test ./internal/app/usecase -run TestTodayView -v`  
Expected: FAIL

**Step 3: 最小实现集合查询规则**

```go
// Today = doing OR overdue OR due today, status in todo/doing
```

**Step 4: 运行测试确认通过**

Run: `go test ./internal/app/usecase -run TestTodayView -v`  
Expected: PASS

**Step 5: Commit**

```bash
git add internal/domain/query.go internal/app/usecase/nav_query.go internal/repo/sqlite/task_repo.go internal/app/usecase/nav_query_test.go
git commit -m "feat: add nav view query rules"
```

### Task 7: TUI 主界面（左 1/4 导航 + 右 3/4 列表）

**Files:**
- Create: `internal/tui/model.go`
- Create: `internal/tui/layout.go`
- Create: `internal/tui/nav.go`
- Create: `internal/tui/list.go`
- Create: `internal/tui/keymap.go`
- Create: `internal/cli/ui.go`
- Test: `internal/tui/model_test.go`

**Step 1: 写失败测试（布局比例与焦点切换）**

```go
func TestLayoutRatio(t *testing.T) {
	m := NewModel(...)
	view := m.View()
	require.Contains(t, view, "Inbox")
}
```

**Step 2: 运行测试确认失败**

Run: `go test ./internal/tui -run TestLayoutRatio -v`  
Expected: FAIL

**Step 3: 最小实现 TUI 框架**

```go
// Tab 切换焦点，Enter 切换左栏集合，右栏显示任务列表
```

**Step 4: 运行测试确认通过**

Run: `go test ./internal/tui -run TestLayoutRatio -v`  
Expected: PASS

**Step 5: Commit**

```bash
git add internal/tui/model.go internal/tui/layout.go internal/tui/nav.go internal/tui/list.go internal/tui/keymap.go internal/cli/ui.go internal/tui/model_test.go
git commit -m "feat: add two-pane TUI navigation and list"
```

### Task 8: 剪贴板规则解析（CLI + TUI）

**Files:**
- Create: `internal/clipboard/read.go`
- Create: `internal/clipboard/normalize.go`
- Create: `internal/clipboard/extract.go`
- Create: `internal/app/usecase/add_from_clipboard.go`
- Modify: `internal/cli/add.go`
- Modify: `internal/tui/model.go`
- Test: `internal/clipboard/normalize_test.go`
- Test: `internal/app/usecase/add_from_clipboard_test.go`

**Step 1: 写失败测试（clip 创建）**

```go
func TestAddFromClipboardRule(t *testing.T) {
	task, err := uc.AddFromClipboard(ctx, text, false)
	require.NoError(t, err)
	require.NotEmpty(t, task.Title)
}
```

**Step 2: 运行测试确认失败**

Run: `go test ./internal/clipboard ./internal/app/usecase -run Clipboard -v`  
Expected: FAIL

**Step 3: 最小实现规则解析与预览**

```go
// 读取文本 -> 清洗 -> 提取 URL -> 生成 title/notes/links
```

**Step 4: 运行测试确认通过**

Run: `go test ./internal/clipboard ./internal/app/usecase -run Clipboard -v`  
Expected: PASS

**Step 5: Commit**

```bash
git add internal/clipboard/read.go internal/clipboard/normalize.go internal/clipboard/extract.go internal/app/usecase/add_from_clipboard.go internal/cli/add.go internal/tui/model.go internal/clipboard/normalize_test.go internal/app/usecase/add_from_clipboard_test.go
git commit -m "feat: support clipboard rule-based task creation"
```

### Task 9: AI Parse（结构化输出 + 校验 + 回退）

**Files:**
- Create: `internal/ai/provider.go`
- Create: `internal/ai/openai/client.go`
- Create: `internal/ai/schema/parse_task_schema.go`
- Create: `internal/ai/cache.go`
- Create: `internal/ai/redact.go`
- Create: `internal/app/usecase/ai_parse_task.go`
- Modify: `internal/cli/add.go`
- Modify: `internal/tui/model.go`
- Test: `internal/ai/schema/parse_task_schema_test.go`
- Test: `internal/app/usecase/ai_parse_task_test.go`

**Step 1: 写失败测试（AI 返回 JSON 通过 schema）**

```go
func TestParseTaskSchema(t *testing.T) {
	err := ValidateParseTaskJSON(sampleResponse)
	require.NoError(t, err)
}
```

**Step 2: 运行测试确认失败**

Run: `go test ./internal/ai/... ./internal/app/usecase -run ParseTask -v`  
Expected: FAIL

**Step 3: 最小实现 AI 调用、校验、失败回退**

```go
if err != nil || !valid {
	return ruleParser.Parse(input), nil
}
```

**Step 4: 运行测试确认通过**

Run: `go test ./internal/ai/... ./internal/app/usecase -run ParseTask -v`  
Expected: PASS

**Step 5: Commit**

```bash
git add internal/ai/provider.go internal/ai/openai/client.go internal/ai/schema/parse_task_schema.go internal/ai/cache.go internal/ai/redact.go internal/app/usecase/ai_parse_task.go internal/cli/add.go internal/tui/model.go internal/ai/schema/parse_task_schema_test.go internal/app/usecase/ai_parse_task_test.go
git commit -m "feat: add ai parse with schema validation and fallback"
```

### Task 10: 端到端验证与文档

**Files:**
- Create: `README.md`
- Create: `docs/specs/mvp-a-interaction.md`
- Create: `docs/specs/mvp-a-keybindings.md`
- Create: `scripts/e2e_smoke.sh`

**Step 1: 写失败脚本（关键路径）**

```bash
#!/usr/bin/env bash
set -euo pipefail
td add "a"
td ls | grep "a"
```

**Step 2: 运行脚本确认失败**

Run: `bash scripts/e2e_smoke.sh`  
Expected: FAIL（路径或命令未完成）

**Step 3: 补齐 README 与交互文档**

```md
- Left Nav: 今日 / Inbox / 日志 / 项目 / 回收站
- Right List: 查询、筛选、排序、编辑
- Clip + AI: 预览后确认写入
```

**Step 4: 再次运行全量验证**

Run: `go test ./... -v && bash scripts/e2e_smoke.sh`  
Expected: PASS

**Step 5: Commit**

```bash
git add README.md docs/specs/mvp-a-interaction.md docs/specs/mvp-a-keybindings.md scripts/e2e_smoke.sh
git commit -m "docs: add mvp-a usage and interaction specs"
```

## Definition of Done

- `go test ./... -v` 全部通过。
- CLI 关键命令可用且状态流转满足约束。
- TUI 达到固定布局：左 1/4 导航、右 3/4 列表。
- `--clip` 与 `--clip --ai` 都可创建任务，AI 异常时自动回退规则解析。
- 文档明确声明 MVP-A 范围，不包含 triage/expand/archived。
