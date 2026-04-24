# forgekit 集成测试精简方案

参考 `docs/integration.md` 的实际使用方式，当前只保留一组“最小但有效”的黑盒集成测试。

## 目标

- 用少量用例保障核心接入链路可用。
- 覆盖外部项目最常用命令：`version get`、`lint`、`publish` 参数校验。
- 避免大而全测试集，优先稳定与可维护。

## 用例范围（精简版）

仅保留 4 条：

1. `forgekit --project-root <root> version get forgekit` 成功。
2. `forgekit --project-root <root> lint --config lint.yaml` 成功。
3. `forgekit --project-root <root> lint --config lint.yaml` 在行数超限时失败。
4. `forgekit --project-root <root> publish chart build --multi-tag`（缺少 `--semver`）失败并给出约束提示。

说明：第 4 条只做参数约束回归，不依赖真实 registry/helm push。

## 组织方式

- 目录：`tests/integration/`
- 文件：先维持单文件 `tests/integration/forgekit_cli_test.go`
- 调用方式：统一 `go run ./cmd/forgekit ...`，全部走 CLI 黑盒路径

## 明确不做

- 不做真实外部依赖场景（registry、k3s、真实推送）。
- 不覆盖全部子命令组合。
- 不在当前阶段扩展大量失败分支用例。

## 运行与门禁

- 保持当前入口：`go test ./...`
- 集成测试默认并入同一入口，不单独拆 Job。

## 后续扩展触发条件

仅在以下情况新增集成测试：

1. 线上/CI 出现回归缺陷。
2. 新增高频命令参数或行为变更。
3. 现有 4 条用例不能覆盖关键风险。

## 当前落实状态

- `tests/integration/forgekit_cli_test.go` 已按精简版覆盖以上 4 条核心用例。
