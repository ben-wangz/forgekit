# 超长文件重构方案（对齐 `lint.yaml` 的 `.go: 250`）

## 背景

当前 `forgekit lint` 已在仓库生效，`commands`（`gofmt`、`go test ./...`）通过，但行数检查失败。超限文件如下：

- `internal/publish/config.go`（339）
- `internal/publish/oci_registry.go`（298）
- `internal/version/command.go`（290）
- `internal/version/manager.go`（274）
- `internal/version/yaml_patch.go`（353）

目标是通过重构将上述文件都压到 250 行以内，且不改变 CLI 对外行为。

## 重构原则

1. 只做“拆分与职责收敛”，不改变命令参数和输出语义。
2. 优先按“领域职责”拆分，而不是机械按行数切块。
3. 每次重构只覆盖 1~2 个文件，保证可回归、可定位。
4. 每步都跑 `go test ./...` 与 `go run ./cmd/forgekit lint`。

## 文件级拆分设计

### 1) `internal/publish/config.go`

问题：容器配置、Chart 配置、registry 推导、label 解析都集中在一个文件。

拆分建议：

- 保留 `config.go`：仅放 `ContainerConfig`/`ChartConfig` 结构体和 `loadContainerConfig`/`loadChartConfig` 主流程。
- 新增 `config_container_flags.go`：容器 CLI 参数解析与校验（`--container-dir`、`--module`、`--push` 等）。
- 新增 `config_chart_registry.go`：`resolveChartRegistry`、`resolveChartCredentials`、`registryHost`、`appendChartsSuffix`、`normalizeRegistry`。
- 新增 `config_container_env.go`：`getContainerRegistry`、`defaultImageNameFromModule`、`BUILD_ARG_*` 收集。
- 新增 `config_label.go`：`addLabel` + label key 校验逻辑。

预期收益：每个文件按主题组织，后续新增参数时不再挤压单文件行数。

### 2) `internal/publish/oci_registry.go`

问题：一次性包含引用解析、HTTP 请求、认证挑战解析、token 获取。

拆分建议：

- 保留 `oci_registry.go`：仅保留 `ociCopyByDigest` 入口与核心流程。
- 新增 `oci_registry_client.go`：`doRegistryRequest`、`doRawHTTP`、`registryResponse`。
- 新增 `oci_registry_auth.go`：`fetchBearerToken`、`parseBearerChallenge`、`parseChallengeParams`、`splitAuthParams`。
- 新增 `oci_registry_ref.go`：`ociReference`、`parseDigestReference`、`parseTagReference`、`splitRegistryRepository`。

预期收益：网络调用、鉴权、引用解析解耦，便于加单测。

### 3) `internal/version/command.go`

问题：命令分发、usage 输出、`get` 子命令细分输出全部耦合。

拆分建议：

- 保留 `command.go`：`Run`、`extractProjectRootFlag`、基础 usage。
- 新增 `command_get.go`：`cmdGet` 及 `--git` 分流逻辑。
- 新增 `command_get_print.go`：`printAllVersions`、`printChartVersion`、`printChartGitVersion`、`printAppVersion`、`printModuleVersion`、`printModuleGitVersion`。

预期收益：`get` 路径后续扩展（如额外输出格式）更易维护。

### 4) `internal/version/manager.go`

问题：配置加载、manager 构建、版本读取、chart image 解析混在一起。

拆分建议：

- 保留 `manager.go`：`Manager` 结构与对外方法（`ModuleVersion`、`ChartVersion`、`ModuleGitVersion` 等）。
- 新增 `manager_loader.go`：`NewManager`、`loadVersionControlConfig`。
- 新增 `manager_chart_images.go`：`extractImagesFromChart` 与 chart annotation 解析。
- 新增 `manager_lookup.go`：`VersionFilePath`、`findImageByName`、`chartByName`。

预期收益：初始化与查询逻辑分层，降低方法间耦合。

### 5) `internal/version/yaml_patch.go`

问题：YAML 路径查找、标量渲染、位置信息计算、token 扫描全部在一个文件里。

拆分建议：

- 保留 `yaml_patch.go`：`patchYAMLScalarValue` 主入口。
- 新增 `yaml_patch_keypath.go`：`splitKeyPath`、`findUniqueKeyPathNode`、`collectKeyPathNodes`。
- 新增 `yaml_patch_render.go`：`renderScalarLiteral`、`isSafePlainScalar`。
- 新增 `yaml_patch_locate.go`：`locateScalarTokenRange`、`lineColumnToOffset`、`bytesIndexByte`。
- 新增 `yaml_patch_scan.go`：`scanDoubleQuotedEnd`、`scanSingleQuotedEnd`、`scanPlainScalarEnd`、`findCommentStart`。

预期收益：语义边界更清晰，后续修复 YAML 边缘 case 时风险更低。

## 实施顺序（建议）

1. 先做 `internal/version/command.go`、`internal/version/manager.go`（纯结构拆分，风险最低）。
2. 再做 `internal/publish/config.go`（参数逻辑较多，但回归面可控）。
3. 再做 `internal/publish/oci_registry.go`（涉及网络与鉴权，需重点回归）。
4. 最后做 `internal/version/yaml_patch.go`（文本扫描逻辑最敏感，放最后单独验证）。

## 当前进度

- 已完成第一批重构：
  - `internal/version/command.go` 已拆分为：
    - `internal/version/command.go`
    - `internal/version/command_get.go`
    - `internal/version/command_get_print.go`
  - `internal/version/manager.go` 已拆分为：
    - `internal/version/manager.go`
    - `internal/version/manager_loader.go`
    - `internal/version/manager_lookup.go`
    - `internal/version/manager_chart_images.go`
- 已执行 `go build ./...`，当前可编译通过。
- 已完成第二批重构：
  - `internal/publish/config.go` 已拆分为：
    - `internal/publish/config.go`
    - `internal/publish/config_label.go`
    - `internal/publish/config_registry.go`
    - `internal/publish/config_container_env.go`
  - `internal/publish/oci_registry.go` 已拆分为：
    - `internal/publish/oci_registry.go`
    - `internal/publish/oci_registry_client.go`
    - `internal/publish/oci_registry_auth.go`
    - `internal/publish/oci_registry_ref.go`
- 已执行 `go build ./...`，当前可编译通过。
- 已完成第三批重构：
  - `internal/version/yaml_patch.go` 已拆分为：
    - `internal/version/yaml_patch.go`
    - `internal/version/yaml_patch_keypath.go`
    - `internal/version/yaml_patch_render.go`
    - `internal/version/yaml_patch_locate.go`
    - `internal/version/yaml_patch_scan.go`
- 已执行 `go build ./...`，当前可编译通过。

## 验证清单

- `go test ./...`
- `go run ./cmd/forgekit lint`（确保 `.go: 250` 全绿）
- 冒烟命令：
  - `go run ./cmd/forgekit version get`
  - `go run ./cmd/forgekit lint --help`
  - `go run ./cmd/forgekit publish --help`

## 风险与回避

- `yaml_patch` 扫描逻辑属于“字节级解析”，容易引入行为偏差；每次拆分后先跑现有 `yaml_patch_test`。
- `oci_registry` 涉及认证分支，建议补 1~2 个纯函数单测（challenge 解析）后再继续演进。
- `config` 解析顺序不能变（环境变量默认值 + CLI 覆盖），拆分时保持当前执行顺序。
