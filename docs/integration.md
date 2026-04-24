# forgekit 项目接入指南

本文档面向“将 `forgekit` 作为外部工具接入你的项目”。

## 1) 下载并校验指定版本

以 Linux amd64 为例：

```bash
VERSION=v0.3.1
OS=linux
ARCH=amd64
BIN="forgekit_${OS}_${ARCH}"

curl -fL -o "${BIN}" "https://github.com/ben-wangz/forgekit/releases/download/${VERSION}/${BIN}"
curl -fL -o checksums.txt "https://github.com/ben-wangz/forgekit/releases/download/${VERSION}/checksums.txt"

sha256sum --check checksums.txt --ignore-missing
chmod +x "${BIN}"
```

如果网络无法直连 GitHub，可改用 `https://files.m.daocloud.io/github.com/...`：

```bash
curl -fL -o "${BIN}" "https://files.m.daocloud.io/github.com/ben-wangz/forgekit/releases/download/${VERSION}/${BIN}"
curl -fL -o checksums.txt "https://files.m.daocloud.io/github.com/ben-wangz/forgekit/releases/download/${VERSION}/checksums.txt"
```

当前仅提供 Linux 与 macOS 平台产物。

## 2) 固定安装路径并通过绝对路径调用

```bash
mkdir -p /opt/forgekit/bin
mv "${BIN}" /opt/forgekit/bin/forgekit
```

推荐在 CI 与脚本中统一使用绝对路径，避免 PATH 与工作目录差异。

## 3) 项目根目录识别规则

`forgekit` 会按以下顺序定位项目根目录：

1. `--project-root <path>`
2. 环境变量 `FORGEKIT_PROJECT_ROOT`
3. 从当前目录向上查找，直到命中 `.git` 或 `version-control.yaml`

建议在自动化场景始终显式传入 `--project-root`。

## 4) 必要配置文件

### 4.1 `version-control.yaml`（`version` / `publish` 依赖）

最小示例（包含模块与可选 binary 声明）：

```yaml
charts:
  - name: demo-operator
    path: operator/chart

binaries:
  - name: your-tool
    path: .
    versionFile: VERSION
```

说明：

- `forgekit version` 读取此文件管理 chart/image/binary 版本。
- `forgekit publish` 内部复用 `version` 逻辑计算发布版本。

### 4.2 `lint.yaml`（可选，`lint` 命令依赖）

最小示例：

```yaml
include:
  - "*.go"

exclude:
  - "**/vendor/**"
  - "**/build/**"

max_lines_by_ext:
  .go: 250

commands:
  - name: "Go formatting check"
    cmd: "sh"
    args:
      - "-c"
      - "test -z \"$(gofmt -l .)\""
```

## 5) 常用调用示例

```bash
# 版本查询
/opt/forgekit/bin/forgekit --project-root /workspace/your-project version get
/opt/forgekit/bin/forgekit --project-root /workspace/your-project version get your-tool

# 代码规范检查
/opt/forgekit/bin/forgekit --project-root /workspace/your-project lint

# 镜像构建（默认 git-version）
/opt/forgekit/bin/forgekit --project-root /workspace/your-project publish container build \
  --container-dir catalog/ingest/container \
  --module catalog/ingest

# chart 打包并推送（semver）
CHART_REGISTRY="ghcr.io/acme/demo-charts" \
CHART_REGISTRY_USERNAME="$GITHUB_ACTOR" \
CHART_REGISTRY_PASSWORD="$GITHUB_TOKEN" \
/opt/forgekit/bin/forgekit --project-root /workspace/your-project publish chart build \
  --chart-dir operator/chart \
  --push \
  --semver
```

## 6) CI 最小模板

```bash
FORGEKIT=/opt/forgekit/bin/forgekit
ROOT="$REPO_ROOT"

${FORGEKIT} --project-root "${ROOT}" lint
${FORGEKIT} --project-root "${ROOT}" version get

${FORGEKIT} --project-root "${ROOT}" publish container build \
  --container-dir catalog/ingest/container \
  --module catalog/ingest \
  --push \
  --label org.opencontainers.image.source="https://github.com/acme/demo" \
  --label org.opencontainers.image.revision="${GIT_SHA}"
```

## 7) 关键行为与注意事项

- `publish` 默认使用 git-version；仅在需要语义化版本发布时传 `--semver`。
- `--multi-tag` 仅支持与 `--semver --push` 一起使用。
- semver 包含 `+` build metadata（如 `1.2.3+build.1`）会报错（OCI tag 不支持 `+`）。
- `CHART_REGISTRY` 不能带 `oci://` 前缀。
- 推荐显式设置 `CHART_REGISTRY` 与 chart 凭据，避免兼容回退路径。
