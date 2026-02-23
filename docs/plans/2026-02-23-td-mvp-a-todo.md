# TD MVP-A TODO Checklist

> 范围已锁定为 A 版（极简首发）：不实现 `archived` / triage / expand。

## 0. 启动前门禁

- [ ] 明确状态机仅包含：`inbox|todo|doing|done|deleted`
- [ ] 明确时间策略：DB 全部 UTC 存储，展示按 `core.timezone`
- [ ] 确认默认创建策略：Inbox= `inbox`，今日/项目= `todo`
- [ ] 确认日志视图窗口默认 14 天（可配置）

## 1. 基础工程（CLI + Config + App 启动）

- [ ] 初始化 Go 模块与目录结构（`cmd/td`, `internal/*`）
- [ ] 建立 root command 与 `td --help`
- [ ] 增加配置加载（`TD_HOME`、默认数据路径、`config.toml`）
- [ ] 增加统一时钟与时间工具
- [ ] 添加最小 Makefile（`test`, `build`）

验收标准：
- [ ] `go test ./cmd/td -v` 通过
- [ ] `td --help` 可执行

## 2. SQLite 与数据层

- [ ] 建立 migration 机制与 `schema_migrations`
- [ ] 建表：`tasks`, `tags`, `task_tags`, `ai_cache`
- [ ] 建索引：`status/project/due_at/updated_at/done_at`
- [ ] 实现 repository 接口与 sqlite 实现
- [ ] 实现状态流转校验（非法流转返回错误）

验收标准：
- [ ] `go test ./internal/repo/sqlite -v` 通过
- [ ] 空库首次启动可自动迁移

## 3. CLI 核心命令

- [ ] `td add [text]`（默认 `status=inbox`, `priority=P2`）
- [ ] `td ls [query]`（支持基础过滤）
- [ ] `td show <id>`
- [ ] `td done <id...>` / `td reopen <id...>`
- [ ] `td edit <id>`（调用 `$EDITOR`）
- [ ] `td rm <id...>`（软删除）
- [ ] `td restore <id...>`
- [ ] `td purge <id...>`（仅删除回收站中的任务）

验收标准：
- [ ] 生命周期命令集成测试通过
- [ ] 非法 `purge` 能正确报错

## 4. 视图查询（TUI 左栏语义）

- [ ] 实现 `Inbox`：`status=inbox`
- [ ] 实现 `今日`：`status in (todo,doing)` 且 overdue/due-today/doing
- [ ] 实现 `日志`：`status=done` 且 `done_at >= now-14d`
- [ ] 实现 `项目`：`project=X` 且 `status in (inbox,todo,doing)`
- [ ] 实现 `回收站`：`status=deleted`

验收标准：
- [ ] 五个集合的 usecase 单测覆盖通过
- [ ] 边界时刻（跨天）测试通过

## 5. TUI 主界面（左 1/4 + 右 3/4）

- [ ] 建立 Bubble Tea model/update/view 主循环
- [ ] 左栏导航渲染与选择切换（`j/k`, `Enter`, `g/G`）
- [ ] 右栏列表渲染与选择（`j/k`）
- [ ] 焦点切换（`Tab`）
- [ ] 基础操作键位：`a/x/e/Backspace/R/P/s/f/1/2/3/4`
- [ ] 在日志/回收站阻止新增并提示

验收标准：
- [ ] TUI 启动后可在五个视图切换
- [ ] 切换视图后右栏列表即时刷新

## 6. 剪贴板规则解析（CLI + TUI）

- [ ] 实现剪贴板读取适配层
- [ ] 实现文本清洗（噪声头、多空行）
- [ ] 实现 URL 提取并写入 `meta_json.links`
- [ ] CLI: `td add --clip`（预览/创建）
- [ ] TUI: `p` 触发剪贴板预览并确认创建

验收标准：
- [ ] 纯文本、含链接、超长文本三类样例通过
- [ ] 超长截断时有提示

## 7. AI Parse Clipboard（可回退）

- [ ] 接入 provider 配置（默认关闭）
- [ ] 实现输入脱敏与长度限制
- [ ] 实现 ParseTask schema 校验
- [ ] CLI: `td add --clip --ai` + `--dry-run` + `--explain`
- [ ] TUI: 剪贴板预览里 `Ctrl+A` 触发 AI 解析建议，确认后写入
- [ ] AI 失败/超时/无 key 时回退规则解析
- [ ] 实现 `ai_cache` 命中逻辑

验收标准：
- [ ] AI 开关关闭时流程不报错
- [ ] AI 异常可自动回退并成功创建任务

## 8. 收尾与发布准备

- [ ] README（安装、命令、键位、隐私说明）
- [ ] 交互规格文档（左栏语义、键位表、剪贴板流程）
- [ ] 导出基础 smoke 脚本
- [ ] 全量测试与手动回归

验收标准：
- [ ] `go test ./... -v` 通过
- [ ] `td ui` 基础交互可运行
- [ ] 文档明确声明 MVP-A 不含 triage/expand

## 建议提交节奏（可选）

- [ ] `chore: bootstrap project skeleton`
- [ ] `feat: add sqlite schema and repository`
- [ ] `feat: add core cli lifecycle commands`
- [ ] `feat: add two-pane tui navigation`
- [ ] `feat: add clipboard rule parser`
- [ ] `feat: add ai parse with fallback`
- [ ] `docs: add mvp-a interaction specs`
