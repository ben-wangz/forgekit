# forgekit 自身接入 `forgekit lint` 方案

## 目标

- 让 `forgekit` 项目自身可直接执行 `forgekit lint`，验证工具在真实仓库中的可用性。
- 规则先“可落地、低阻塞”，再逐步收紧，避免一次性引入过多改造成本。
- 保持与对外使用方式一致：只依赖普通模式，不依赖 pre-commit。

## 已确认决策

1. 阶段 1 包含 `go test ./...`。
2. `.go` 初始最大行数阈值为 `250`。
3. `README` 在实现与规则稳定后再修改（最后处理）。

## 约束与现状

- 当前 `lint` 已支持：`commands` + `max_lines_by_ext` + `include/exclude`。
- `include` 目前按文件名匹配（非相对路径）。
- glob 为当前实现语义（非 `doublestar` 完整兼容）。
- 项目已有 Go 代码与测试，适合先覆盖 `.go` 文件。

## 推荐落地路径（分三阶段）

### 阶段 1：最小可用（立即可做）

在仓库根目录新增 `lint.yaml`，仅覆盖 Go 代码的基础规则：

1. `gofmt` 格式检查。
2. `go test ./...` 基础验证。
3. `.go` 文件最大行数限制设为 `250`。

建议配置草案：

```yaml
include:
  - "*.go"

exclude:
  - "**/vendor/**"
  - "**/dist/**"
  - "**/build/**"

max_lines_by_ext:
  .go: 250

commands:
  - name: "Go formatting check"
    cmd: "sh"
    args:
      - "-c"
      - "test -z \"$(gofmt -l .)\""

  - name: "Go unit tests"
    cmd: "go"
    args: ["test", "./..."]
```

执行方式：

```bash
forgekit lint
```

### 阶段 2：规则收敛（稳定后）

- 将 `.go` 行数阈值从 `300` 逐步降到 `250` 或 `200`。
- 观察 `include`/`exclude` 对多级目录命中是否符合预期，必要时先通过模式规避。
- 对失败输出做可读性优化（例如增加规则名、文件计数）。

### 阶段 3：工程化接入（CI）

- 在 CI 增加一步：下载指定版本 `forgekit` 后执行 `forgekit lint`。
- 通过 `--project-root` 显式绑定仓库根目录，减少环境差异。
- 与本地开发保持同一命令，确保“本地过即 CI 过”。

## 实施顺序

1. 落地根目录 `lint.yaml`（含 `gofmt` + `go test ./...` + `.go: 250`）。
2. 在本地与 CI 验证 `forgekit lint` 行为。
3. 最后更新 `README` 的 lint 章节与仓库自用说明。

## 存量超限治理

行数超限文件的重构方案已单独整理到 `todo/lint-refactor-plan.md`，后续按该方案分批实施并回归。

## 验收标准

- 根目录存在可用 `lint.yaml`。
- 在本仓库执行 `forgekit lint` 可稳定运行并产生可解释结果。
- CI 中可复用同一命令，无需额外脚本包装。
