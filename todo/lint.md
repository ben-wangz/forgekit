# lint 当前实现状态（仅普通模式）

## 目标（已收敛）

仅保留 `forgekit lint` 的普通模式，方便其他项目直接复用 `forgekit` 二进制（对齐 `version` / `publish` 的使用方式）。

不实现 pre-commit 子模式，不在 `forgekit` 内维护 git hook 安装逻辑。

## 已实现

### 1) 命令入口已接入统一 CLI

- `cmd/forgekit/main.go`
  - 已注册 `lint` 子命令：`forgekit lint`。
  - 与 `version` / `publish` 共用全局参数风格（`--project-root`、`--help`、`--version`）。

### 2) lint 普通模式执行链路可用

- `internal/lint/command.go`
  - 支持参数：
    - `--config <path>`（可选，默认 `<project-root>/lint.yaml`）
    - `--project-root <path>`（覆盖全局推导）
  - 配置文件相对路径按 `project-root` 解析。
  - 读取配置后，按顺序执行：
    1. `commands`
    2. `max_lines_by_ext`

### 3) 配置解析已实现

- `internal/lint/config.go`
  - 已支持字段：
    - `commands`
    - `max_lines_by_ext`
    - `include`
    - `exclude`

### 4) 命令检查已实现

- `internal/lint/runner.go`
  - 使用 `exec.Command` 串行执行 `commands`。
  - 工作目录固定为配置文件所在目录。
  - stdout/stderr 直出。
  - 失败即返回错误，成功打印 `✓ <name> passed`。

### 5) 行数检查已实现

- `internal/lint/checker.go`
  - 遍历配置目录下文件并执行 include/exclude 过滤。
  - 根据扩展名读取 `max_lines_by_ext` 阈值。
  - 超限文件输出错误并最终返回失败。

### 6) 基础测试已补充

- `tests/lint/run_test.go`
  - 覆盖默认配置发现（从工作目录向上找到 git root 后读取 `lint.yaml`）。
  - 覆盖行数超限失败场景。

## 当前限制与已知差异

1. `include` 规则当前基于文件名（`filepath.Base`）匹配，不是基于相对路径匹配。
2. glob 匹配为自实现（正则转换），并非 `doublestar` 语义，复杂模式兼容性需谨慎。
3. 暂无 pre-commit 模式（符合当前目标）。
4. 目前测试覆盖较基础，参数边界和复杂模式尚未系统覆盖。

## 与目标一致性结论

结论：当前实现已经满足“仅普通模式 + 可被其他项目直接引用 forgekit 二进制”的主目标。

仍建议后续补齐少量鲁棒性工作（见下）。

## 后续建议（小步完善）

- 增加测试：
  - `--config` 绝对/相对路径行为。
  - `exclude` 与多级目录匹配。
  - `commands` 失败时的错误信息与中断行为。
- 明确文档：
  - 在 README 中强调 `lint` 仅普通模式，不提供 hook 管理。
  - 补充 `include` 当前按文件名匹配的说明，避免误解。

## forgekit 自身落地

已新增自用落地方案文档：`todo/lint-forgekit-self.md`，用于讨论如何在本仓库启用 `forgekit lint` 并逐步收敛规则。
