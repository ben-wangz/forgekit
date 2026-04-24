# forgekit 测试策略说明

本文档说明本项目当前的测试原则、范围与执行方式。

## 测试原则

1. **集成测试优先**：优先保障用户真实调用链路（CLI 黑盒）。
2. **用例精简**：只保留核心路径，不追求大而全覆盖。
3. **稳定优先**：避免脆弱、重依赖、难维护的测试。
4. **问题驱动补充**：unit test 在出现缺陷或定位困难时再增加。

## 当前测试结构

- `tests/integration/`：CLI 黑盒集成测试（主保障）。
- `tests/version/`：保留少量核心回归（版本读取/更新与关键解析逻辑）。
- `tests/publish/`：保留少量核心规则回归（tag 解析核心约束）。
- `tests/lint/`：保留关键行为回归（默认配置发现）。

说明：`internal/*` 下目前基本不放大量单测，测试重心放在 `tests/` 下的行为验证。

## 集成测试最小集合

当前维护以下核心集成用例（见 `tests/integration/forgekit_cli_test.go`）：

1. `version get forgekit` 成功。
2. `lint --config lint.yaml` 成功。
3. `lint` 行数超限失败。
4. `publish chart build --multi-tag` 缺少 `--semver` 时失败。

这 4 条作为接入链路最小回归基线。

## 何时新增测试

仅在以下情况新增：

1. 线上或 CI 出现回归问题。
2. 新增高频命令参数或关键行为变更。
3. 现有最小用例无法覆盖新增风险。

## 运行方式

本地与 CI 默认统一使用：

```bash
go test ./...
```

仅执行集成测试：

```bash
go test ./tests/integration/...
```
