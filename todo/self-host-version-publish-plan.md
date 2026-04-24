# forgekit 自举发布最终方案（VERSION 主导 + tag 锚点）

## 结论

- `VERSION` 是唯一版本真相（source of truth），由 `forgekit version` 维护。
- `tag` 仅作为 GitHub Release 锚点（触发发布、承载 artifacts），不再作为版本来源。
- 发布必须满足一致性：`VERSION == tag 去掉前缀 v 后的值`。

## 目标

1. 让 `forgekit` 自身可用 `forgekit version` 管理自身版本。
2. 保留 GitHub Release 分发能力（必须依赖 tag）。
3. 强化可追溯性：版本、tag、commit、二进制元信息四者一致。

## 核心原则

### 1) VERSION 主导

- 新增根目录 `VERSION`，例如 `0.2.0`。
- 新增根目录 `version-control.yaml`：

```yaml
binaries:
  - name: forgekit
    path: .
    versionFile: VERSION
```

- 所有版本变更通过以下命令完成：
  - `forgekit version get forgekit`
  - `forgekit version bump forgekit patch|minor|major`

### 2) tag 锚点

- release 依然由 tag 触发（`v*`），但 tag 值来自 `VERSION`。
- 创建 tag 的标准命令：
  - `git tag "v$(forgekit version get forgekit)"`

### 3) 一致性强校验

在 `.github/workflows/release.yml` 中新增校验：

1. 读取 `VERSION`。
2. 读取 `${GITHUB_REF_NAME#v}`。
3. 不一致则立即失败（禁止发布）。

## 发布流程（最终）

1. `forgekit lint`
2. `forgekit version bump forgekit <patch|minor|major>`
3. 提交版本变更（至少包含 `VERSION`）
4. 创建并推送 tag：`v$(forgekit version get forgekit)`
5. GitHub Actions `release.yml` 触发并执行一致性校验后发布

## CI 与工作流改造

### CI（`.github/workflows/ci.yml`）

- 增加：`go run ./cmd/forgekit version get forgekit`
- 增加：`go run ./cmd/forgekit lint`

目的：确保自举工具链持续可用。

### Release（`.github/workflows/release.yml`）

- 保持 tag 触发。
- 增加 VERSION/tag 一致性校验步骤。
- 继续在构建 ldflags 注入 `main.version` 与 `main.commit`。

## 追溯策略

- 发布物料至少包含：
  - 版本号（来自 `VERSION`）
  - tag（`vX.Y.Z`）
  - commit SHA
  - checksums
- 二进制仍注入 commit 信息，支持运行时追溯：
  - `forgekit --version` 可显示版本与 commit。

## 非目标

- 不做“无 tag 发布”（GitHub Release 与 tag 绑定，当前不迁移存储渠道）。
- 不引入额外发布系统（如对象存储/OCI 作为主分发）作为当前阶段任务。

## 实施阶段

### Phase 1（版本源落地）

- 新增 `VERSION` 与 `version-control.yaml`。
- 本地验证：
  - `forgekit version get forgekit`
  - `forgekit version bump forgekit patch`

### Phase 2（工作流收口）

- 更新 `ci.yml`（`forgekit version get forgekit` + `forgekit lint`）。
- 更新 `release.yml`（VERSION/tag 一致性校验）。

### Phase 3（文档定稿）

- README 增加“forgekit 自身发布”章节。
- 给出标准发布 playbook（可放 `docs/release.md`）。

## 验收标准

- `forgekit version get forgekit` 返回根目录 `VERSION` 值。
- 版本 bump 后可通过 tag 成功触发发布。
- tag 与 VERSION 不一致时，release workflow 必须失败。
- README 中有可执行的自举发布步骤。

## 进度

- Phase 1：已完成。
- Phase 2：已完成。
- Phase 3：已完成（`README` 与 `docs/release.md` 已补充）。
