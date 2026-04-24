# forgekit

`forgekit` 是一个独立维护的统一 CLI，当前聚合了 `lint`、`version` 与 `publish` 能力。

## 安装

### 本地构建

```bash
go build -o dist/bin/forgekit ./cmd/forgekit
```

或使用脚本（支持绝对路径调用）：

```bash
/path/to/forgekit/scripts/build.sh
```

### GitHub Releases 二进制

发布产物由 GitHub Actions 在 tag 触发后构建并上传到 Releases（不发布工具镜像）。

```bash
VERSION=v0.1.0
OS=linux
ARCH=amd64
BIN="forgekit_${OS}_${ARCH}"

curl -fL -o "${BIN}" "https://github.com/ben-wangz/forgekit/releases/download/${VERSION}/${BIN}"
curl -fL -o checksums.txt "https://github.com/ben-wangz/forgekit/releases/download/${VERSION}/checksums.txt"

sha256sum --check checksums.txt --ignore-missing
chmod +x "${BIN}"
mv "${BIN}" /usr/local/bin/forgekit
```

如果网络无法直连 GitHub，可改用 `https://files.m.daocloud.io/github.com/...` 形式，例如：

```bash
curl -fL -o "${BIN}" "https://files.m.daocloud.io/github.com/ben-wangz/forgekit/releases/download/${VERSION}/${BIN}"
curl -fL -o checksums.txt "https://files.m.daocloud.io/github.com/ben-wangz/forgekit/releases/download/${VERSION}/checksums.txt"
```

Windows 产物名为 `forgekit_windows_amd64.exe` / `forgekit_windows_arm64.exe`。

## CLI 结构

```text
forgekit
├── lint
├── version
│   ├── get
│   ├── bump
│   ├── bump-chart
│   └── sync
└── publish
    ├── container build
    └── chart build
```

## 使用示例

默认会从当前目录向上查找项目根目录；也可以显式指定：

```bash
forgekit --project-root /path/to/project version get
```

### version

`version-control.yaml` 可以为独立 binary 显式声明版本文件：

```yaml
binaries:
  - name: forgekit
    path: .
    versionFile: VERSION
```

```bash
# 列出所有 chart/image 版本
forgekit version get

# 获取 binary 语义化版本
forgekit version get forgekit

# 获取模块语义化版本
forgekit version get catalog/ingest

# 获取模块 git-version
forgekit version get catalog/ingest --git

# 获取 chart git-version
forgekit version get chart astro-data-operator --git

# bump 模块版本
forgekit version bump catalog/ingest patch

# bump binary 版本
forgekit version bump forgekit patch

# bump chart 版本并同步 image 版本到 values/appVersion
forgekit version bump-chart astro-data-operator minor --sync

# 同步所有 chart 的 image 版本
forgekit version sync
```

### lint

默认从当前目录向上发现 git 根目录，并读取 `<git-root>/lint.yaml`。

```bash
# 使用 <git-root>/lint.yaml
forgekit lint

# 显式指定配置文件（相对路径按 git 根目录解析）
forgekit lint --config lint.yaml
```

### publish

```bash
# 构建容器镜像（tag 自动走 version 逻辑）
forgekit publish container build \
  --container-dir catalog/ingest/container \
  --module catalog/ingest

# 构建并推送容器镜像
forgekit publish container build \
  --container-dir catalog/ingest/container \
  --module catalog/ingest \
  --push

# 显式传入 OCI labels（可重复）
forgekit publish container build \
  --container-dir catalog/ingest/container \
  --module catalog/ingest \
  --label org.opencontainers.image.source=https://github.com/acme/demo \
  --label org.opencontainers.image.revision=abc123

# 使用语义化版本发布（不带 commit，要求仓库 clean）
forgekit publish container build \
  --container-dir catalog/ingest/container \
  --module catalog/ingest \
  --push \
  --semver

# 语义化版本多标签发布（latest, major, major.minor, full）
forgekit publish container build \
  --container-dir catalog/ingest/container \
  --module catalog/ingest \
  --push \
  --semver \
  --multi-tag

# 打包 chart
forgekit publish chart build --chart-dir operator/chart

# 打包并推送 chart
forgekit publish chart build --chart-dir operator/chart --push

# 显式指定 chart registry（推荐，避免兼容回退 warning）
CHART_REGISTRY=ghcr.io/acme/demo-charts \
forgekit publish chart build --chart-dir operator/chart --push

# 使用语义化版本发布 chart（不带 commit，要求仓库 clean）
forgekit publish chart build --chart-dir operator/chart --push --semver

# 使用语义化版本多标签发布 chart（latest, major, major.minor, full）
forgekit publish chart build --chart-dir operator/chart --push --semver --multi-tag
```

`publish` 默认使用 git-version（带 commit，dirty 时可能带 `-dirty`）；传 `--semver` 则改为 semver 版本（不带 commit）。
当启用 `--semver` 且仓库 dirty 时，命令会报错退出。

`--multi-tag` 仅支持与 `--semver --push` 一起使用；不支持自定义 `--tag` 参数。
当 semver 为 `0.x.y` 或 prerelease（如 `1.2.3-alpha.1`）时，`--multi-tag` 会降级为单 tag 并输出 warning（container/chart 一致）。
当 semver 含 `+` build metadata（如 `1.2.3+build.1`）时，命令会报错退出（OCI tag 不支持 `+`）。

## 配置与环境变量

### 通用

- `FORGEKIT_PROJECT_ROOT`：默认项目根目录（可被 `--project-root` 覆盖）

### publish 相关

- `IMAGE_NAME`：镜像名，默认 `astro-data/<module>`（`/` 转 `-`）
- `CONTAINER_REGISTRY`：目标 registry，未设置时自动探测 k3s registry，失败回落 `localhost:5000`
- `CONTAINER_REGISTRY_USERNAME` / `CONTAINER_REGISTRY_PASSWORD`：registry 认证
- `CHART_REGISTRY`：chart 推送地址（不带 `oci://`，推荐显式设置）
- `CHART_REGISTRY_USERNAME` / `CHART_REGISTRY_PASSWORD`：chart registry 认证（推荐显式设置）
- `REGISTRY_PLAIN_HTTP`：是否使用 HTTP（`true/false`）
- `BUILD_ARG_*`：透传给 `podman build --build-arg`
- `KUBECONFIG`：k3s/cluster 访问配置路径

未设置 `CHART_REGISTRY` 时会走兼容回退路径（基于 `CONTAINER_REGISTRY` 或自动探测）并输出 deprecation warning，后续版本可能移除。

## 项目独立说明

- 统一二进制：`forgekit`
- `publish` 直接复用同仓库内 `internal/version` 逻辑
- 命令结构：
  - `forgekit lint ...`
  - `forgekit version ...`
  - `forgekit publish ...`

## forgekit 自身发布

`forgekit` 自身采用 **VERSION 主导 + tag 锚点** 模式：

- 版本真相来自仓库根目录 `VERSION`
- `version-control.yaml` 声明了 `forgekit` binary 的版本文件映射
- GitHub Release 仍由 `v*` tag 触发（仅作为发布锚点）
- release workflow 会强校验 `VERSION` 与 tag 一致

常用命令：

```bash
# 查看 forgekit 自身版本
forgekit version get forgekit

# bump 版本（示例：patch）
forgekit version bump forgekit patch
```

标准发布步骤：

```bash
# 1) 校验仓库
forgekit lint

# 2) bump 版本（patch/minor/major 按需替换）
forgekit version bump forgekit patch

# 3) 提交版本变更
git add VERSION
git commit -m "Bump forgekit version"

# 4) 创建并推送与 VERSION 对齐的 tag
git tag "v$(forgekit version get forgekit)"
git push origin main --tags
```

详细发布说明见 `docs/release.md`。

## 项目接入指南

详细接入步骤见 `docs/integration.md`，包含：

- 指定版本下载二进制
- 校验 checksum
- 在 CI/脚本中通过绝对路径调用
- 在目标项目里用 `--project-root` 或 `FORGEKIT_PROJECT_ROOT` 绑定仓库根目录
