# forgekit 项目接入指南

以下示例演示其他项目如何按版本接入 `forgekit` 二进制。

## 1) 下载指定版本

以 Linux amd64 为例：

```bash
VERSION=v0.1.0
OS=linux
ARCH=amd64
BIN="forgekit_${OS}_${ARCH}"

curl -fL -o "$BIN" "https://github.com/ben-wangz/forgekit/releases/download/${VERSION}/${BIN}"
curl -fL -o checksums.txt "https://github.com/ben-wangz/forgekit/releases/download/${VERSION}/checksums.txt"
```

Windows 产物名称为 `forgekit_windows_amd64.exe` / `forgekit_windows_arm64.exe`。

## 2) 校验完整性

```bash
sha256sum --check checksums.txt --ignore-missing
chmod +x "$BIN"
```

## 3) 放置到固定工具目录

```bash
mkdir -p /opt/forgekit/bin
mv "$BIN" /opt/forgekit/bin/forgekit
```

## 4) 在项目中调用

推荐显式传入项目根目录，避免依赖当前工作目录：

```bash
/opt/forgekit/bin/forgekit --project-root /workspace/your-project version get
/opt/forgekit/bin/forgekit --project-root /workspace/your-project publish chart build --chart-dir operator/chart
```

也可以设置环境变量：

```bash
export FORGEKIT_PROJECT_ROOT=/workspace/your-project
/opt/forgekit/bin/forgekit version get catalog/ingest --git
```

## 5) CI 中的最小示例

```bash
/opt/forgekit/bin/forgekit --project-root "$REPO_ROOT" version get
/opt/forgekit/bin/forgekit --project-root "$REPO_ROOT" publish container build \
  --container-dir catalog/ingest/container \
  --module catalog/ingest
```
