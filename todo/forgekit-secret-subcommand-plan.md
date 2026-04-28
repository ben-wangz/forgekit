# Forgekit `secret` 子命令集成方案

## 1. 背景与目标

当前 `forgekit` 不包含 secret 管理能力。我们希望将 `build/notebook/tools/secret` 的核心能力集成到 `forgekit secret` 子命令中，用于在 Git 仓库内管理敏感文件，避免明文提交。

本次集成范围：

- 保留并集成以下能力：
  - 单文件加密/解密
  - SSH 密钥发现与校验、私钥保护提示

## 2. 现有实现能力映射

来源工具当前能力（`build/notebook/tools/secret`）：

- `encrypt <file>`：加密单个 `*.secret.*` 明文文件，输出 `*.enc`
- `decrypt <file.enc>`：解密为明文，写入权限 `0600`
- 密钥逻辑：支持环境变量覆盖，否则默认 `~/.ssh/id_ed25519(.pub)`，仅支持 ed25519
- 私钥保护检查：发现无口令私钥时警告，不阻断流程

## 3. 集成后的 CLI 设计

### 3.1 顶层命令

在 `forgekit` 主入口增加子命令：

- `forgekit secret <command> [args]`

并在全局 usage 中加入 `secret` 描述。

### 3.2 子命令列表

- `forgekit secret encrypt <file>`
- `forgekit secret decrypt <file.enc>`
- `forgekit secret help`

说明：保持子命令集简洁，聚焦单文件加解密能力。

### 3.3 命令行为约束

- `encrypt`：
  - 输入必须匹配 `*.secret.*` 且不能以 `.enc` 结尾
  - 输出 `<file>.enc`
- `decrypt`：
  - 输入必须以 `.enc` 结尾
  - 输出去除 `.enc` 后缀，权限 `0600`

## 4. 技术方案

### 4.1 目录与模块划分

新增目录：`internal/secret/`

建议文件拆分：

- `command.go`：`Run(args []string) error` 与参数解析
- `config.go`：密钥路径解析、环境变量读取、路径校验
- `crypto.go`：ed25519 到 curve25519 派生、文件格式读写、加解密逻辑
- `keycheck.go`：私钥保护检查

保持与现有子命令一致的错误处理和 help 风格，降低主命令接入复杂度。

### 4.2 文件格式与加密算法

沿用现有格式，避免迁移成本：

- Header: `SECRET-V1` + `\n`
- 结构：`ephemeralPub(32)` + `nonce(24)` + `dataLen(uint64 little-endian)` + `ciphertext`
- 算法：NaCl box（curve25519 + XSalsa20-Poly1305）

兼容性目标：旧 `secret` 工具产物可被 `forgekit secret decrypt` 直接解密。

### 4.3 密钥与配置策略

优先级：

1. `SECRET_PRIVATE_KEY` / `SECRET_PUBLIC_KEY`
2. 默认 `~/.ssh/id_ed25519` / `~/.ssh/id_ed25519.pub`

约束：

- 仅支持 ed25519（与现有实现一致）
- 私钥存在但无口令时仅告警，不阻断

### 4.4 路径语义

`secret` 子命令不支持 `--project-root`。文件路径按当前工作目录解析，遵循常见 CLI 直觉行为。

### 4.5 错误处理与返回码

- 参数错误、密钥不可用、解密失败返回非 0

### 4.6 输出与可观测性

- 单文件操作输出 `Encrypted:` / `Decrypted:` 行

## 5. 与现有 forgekit 的接入点

### 5.1 主命令入口改造

修改 `cmd/forgekit/main.go`：

- 增加 `internal/secret` import
- `switch rest[0]` 添加 `case "secret": return secretcmd.Run(rest[1:])`
- `printUsage()` 中加入 `secret` 命令说明和示例

### 5.2 子命令参数

- 支持 `help/-h/--help` 显示帮助
- 不支持 `--project-root`，保持命令简洁

## 6. 安全与风险评估

### 6.1 风险

- 用户使用未加密 SSH 私钥，导致 secret 解密风险增加
- 错误的命名规则使用（未遵循 `*.secret.*`）导致遗漏

### 6.2 缓解

- 保留私钥保护告警并强化提示文案
- 在命令帮助中明确命名规则与建议

## 7. 验证方案

- 手工验证 `encrypt -> decrypt` 往返一致性（内容一致、权限正确）
- 手工验证与旧工具生成文件互通解密

## 8. 分阶段实施计划

### 阶段 1：开发

- 新建 `internal/secret/` 并实现 `encrypt`、`decrypt` 核心能力
- 迁移并整理密钥发现/校验与私钥保护提示逻辑
- 在 `cmd/forgekit/main.go` 注册 `secret` 子命令并完善 help 文案
- 保持 `SECRET-V1` 格式兼容，确保与旧工具产物互通

### 阶段 2：集成测试

- 执行端到端验证：`encrypt -> decrypt` 往返一致性（内容一致、权限正确）
- 验证与旧工具生成文件的解密兼容性
- 验证异常场景（错误文件名、错误后缀、缺失密钥）报错符合预期
- 根据测试结果修正文档（命令帮助/README）

## 9. 预估改动清单（代码级）

- 新增：`internal/secret/*.go`
- 修改：`cmd/forgekit/main.go`
- 修改：项目文档（命令帮助/README）

---

文档版本：v1
创建时间：2026-04-28
