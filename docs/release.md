# forgekit Release Playbook

本文档描述 `forgekit` 项目自身的发布流程。

## 发布模型

- 版本来源：仓库根目录 `VERSION`
- 版本管理：`forgekit version`
- 发布锚点：Git tag `vX.Y.Z`
- 发布载体：GitHub Releases

说明：tag 仅用于触发与锚定 release，版本真相始终来自 `VERSION`。

## 发布前准备

1. 在仓库根目录执行命令。
2. 确保本地分支与远端同步。
3. 确保有发布权限（push main 与 push tag）。

## 标准发布步骤

```bash
# 1) 质量检查
forgekit lint

# 2) 查看当前版本
forgekit version get forgekit

# 3) bump 版本（patch/minor/major）
forgekit version bump forgekit patch

# 4) 再次确认版本
NEW_VERSION="$(forgekit version get forgekit)"
echo "$NEW_VERSION"

# 5) 提交版本变更
git add VERSION
git commit -m "Bump forgekit to v${NEW_VERSION}"

# 6) 创建并推送 tag
git tag "v${NEW_VERSION}"
git push origin main
git push origin "v${NEW_VERSION}"
```

## Workflow 行为

`release.yml` 会在 tag 推送后执行以下关键校验：

1. 读取仓库 `VERSION`
2. 解析 tag 版本（去掉前缀 `v`）
3. 若两者不一致，workflow 直接失败

校验通过后，workflow 执行多平台构建、生成 `checksums.txt` 并上传到 GitHub Release。

## 回滚策略

若误 bump 或误打 tag：

1. 先停止继续发布。
2. 在新提交中修正 `VERSION`。
3. 重新创建正确版本的 tag 并推送。

不建议改写已发布 tag 历史。
