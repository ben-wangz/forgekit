# forgekit

`forgekit` 是一个独立维护的统一 CLI，当前聚合了 `version` 与 `publish` 能力。

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

```bash
# 列出所有 chart/image 版本
forgekit version get

# 获取模块语义化版本
forgekit version get catalog/ingest

# 获取模块 git-version
forgekit version get catalog/ingest --git

# 获取 chart git-version
forgekit version get chart astro-data-operator --git

# bump 模块版本
forgekit version bump catalog/ingest patch

# bump chart 版本并同步 image 版本到 values/appVersion
forgekit version bump-chart astro-data-operator minor --sync

# 同步所有 chart 的 image 版本
forgekit version sync
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

# 使用语义化版本发布（不带 commit，要求仓库 clean）
forgekit publish container build \
  --container-dir catalog/ingest/container \
  --module catalog/ingest \
  --push \
  --semver

# 打包 chart
forgekit publish chart build --chart-dir operator/chart

# 打包并推送 chart
forgekit publish chart build --chart-dir operator/chart --push

# 使用语义化版本发布 chart（不带 commit，要求仓库 clean）
forgekit publish chart build --chart-dir operator/chart --push --semver
```

`publish` 默认使用 git-version（带 commit，dirty 时可能带 `-dirty`）；传 `--semver` 则改为 semver 版本（不带 commit）。
当启用 `--semver` 且仓库 dirty 时，命令会报错退出。

## 配置与环境变量

### 通用

- `FORGEKIT_PROJECT_ROOT`：默认项目根目录（可被 `--project-root` 覆盖）

### publish 相关

- `IMAGE_NAME`：镜像名，默认 `astro-data/<module>`（`/` 转 `-`）
- `CONTAINER_REGISTRY`：目标 registry，未设置时自动探测 k3s registry，失败回落 `localhost:5000`
- `CONTAINER_REGISTRY_USERNAME` / `CONTAINER_REGISTRY_PASSWORD`：registry 认证
- `REGISTRY_PLAIN_HTTP`：是否使用 HTTP（`true/false`）
- `BUILD_ARG_*`：透传给 `podman build --build-arg`
- `KUBECONFIG`：k3s/cluster 访问配置路径

## 项目独立说明

- 统一二进制：`forgekit`
- `publish` 直接复用同仓库内 `internal/version` 逻辑
- 命令结构：
  - `forgekit version ...`
  - `forgekit publish ...`

## 项目接入指南

详细接入步骤见 `docs/integration.md`，包含：

- 指定版本下载二进制
- 校验 checksum
- 在 CI/脚本中通过绝对路径调用
- 在目标项目里用 `--project-root` 或 `FORGEKIT_PROJECT_ROOT` 绑定仓库根目录
