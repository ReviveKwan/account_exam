# 小程序上传交接说明（miniprogram-ci 路径）

> 记录时间：2026-11-17。**最终采用 wechatide 路径并已上传成功**（miniprogram-ci 因沙箱故障弃用，脚本保留备用）。

## ✅ 上传已完成（wechatide 路径）

- **结果**：体验版上传成功，版本 `1.0.0`，描述「V1.0 题库598题 断点续答 错题本 法条溯源」，包大小 79174 字节
- **下一步（需你在网页操作）**：微信公众平台 mp.weixin.qq.com → 版本管理 → 选择该体验版 → 提交审核 → 审核通过后发布
- **wechatide 工具链**：
  - Nightly Electron 版 2.02.2611162，安装于 `$HOME/wechatwebdevtools-nightly.app`
  - CLI：`$HOME/wechatwebdevtools-nightly.app/Contents/MacOS/wechatide`
  - 命令格式：`wechatide -c WorkBuddy <tool> [--flags]`（注意 `-c` 在前）
  - 授权已完成（用户已确认），登录用户 🌳Jovair Kwan💻🌊🐬，skill v0.3.17
  - 常用：`check_wechatide_status`（登录态）、`project_import --project <路径>`、`open_project_window --project <路径> --window-mode fullMode`、`upload --project <路径> --upload-version 1.0.0 --desc "..."`
  - 异步授权/任务：`polling_task_result --task-id <id>` 轮询

## 已完成（miniprogram-ci 备用路径）

1. **AppID 已写入** `miniprogram/project.config.json`：`wxe161a5d9e2c0f148`
2. **上传脚本已就绪**：
   - `scripts/upload.js` —— CI 上传体验版
   - `scripts/preview.js` —— 生成真机预览二维码
   - 两者都从环境变量 `WX_PRIVATE_KEY_PATH` 读取私钥路径，**不硬编码密钥**
3. **依赖已声明** `package.json`：`miniprogram-ci ^2.1.48`
4. **私钥已确认有效**：`/Users/revivekwan/Downloads/private.wxe161a5d9e2c0f148.key`（RSA PRIVATE KEY，1675 字节，文件名与 AppID 吻合）

## 密钥文件辨析（重要）

用户在 Downloads 提供三个文件，用途不同：

| 文件 | 用途 | 本次是否使用 |
|------|------|------------|
| `private.wxe161a5d9e2c0f148.key` | 小程序**代码上传私钥** | ✅ 是（CI 上传用） |
| `小程序对称密钥.txt` | CloudBase 服务端对称密钥 | ❌ 否（后端/云函数用，勿入前端） |
| `小程序非对称密钥.txt` | CloudBase 服务端 RSA 私钥 | ❌ 否（后端/云函数用，勿入前端） |

⚠️ 三个都是敏感凭据，均不应提交到代码仓库、不应写入前端代码、不应出现在对话记录中。

## 备用路径：miniprogram-ci（仅当 wechatide 不可用时）

**背景**：本会话中 npm install 因沙箱故障（`sandbox-exec: pattern serialization length 65633 exceeds maximum`）未能完成，故改用 wechatide。以下命令在环境恢复后仍可执行。

### 步骤 1：安装依赖

```bash
cd "/Users/revivekwan/WorkBuddy AI/2026-11-17-13-13-53/accounting-exam"
npm install --no-audit --no-fund
# 若报 sandbox-exec 超长错误，关闭沙箱或在系统终端里执行
node -e "console.log(require('./node_modules/miniprogram-ci/package.json').version)"
```

### 步骤 2：配置 IP 白名单

当前公网出口 IP：`198.40.51.242`（若变化，用 `curl https://ifconfig.me/ip` 重新获取）。

在微信公众平台 → 开发管理 → 开发设置 → 小程序代码上传 → IP 白名单，加入该 IP。**私钥和 IP 白名单必须在同一账号下匹配**，否则上传报 401/白名单错误。

### 步骤 3：生成预览二维码（可选，先验证）

```bash
cd "/Users/revivekwan/WorkBuddy AI/2026-11-17-13-13-53/accounting-exam"
export WX_PRIVATE_KEY_PATH="/Users/revivekwan/Downloads/private.wxe161a5d9e2c0f148.key"
node scripts/preview.js
# 二维码：scripts/output/preview-qrcode.png，微信扫码真机预览（约25分钟有效）
```

### 步骤 4：上传体验版

```bash
export WX_PRIVATE_KEY_PATH="/Users/revivekwan/Downloads/private.wxe161a5d9e2c0f148.key"
node scripts/upload.js "V1.0 题库598题 断点续答 错题本"
```

上传成功后：微信公众平台 → 版本管理 → 选择该版本 → 提交审核 → 审核通过后发布。

## 常见错误

- `invalid ip` / `ip out of whitelist` → 步骤 2 白名单 IP 不对或出口 IP 变了
- `private key error` / 签名失败 → 私钥与 AppID 不匹配，或私钥文件有多余空格/换行的字符串转义（脚本用 `fs.readFileSync(...,'utf8')` 直接读文件，一般无此问题）
- `80001` 等 → 多为公众平台后台权限/主体问题，需登录 mp.weixin.qq.com 检查

## 相关：另外两条部署路径（未变）

- **数据库**：已导入 CloudBase 文档数据库（chapters 12 / exams 12 / questions 598），环境 `wechat-app-d6gjceiz966031f7f`
- **后端 Golang**：仍待部署到 CloudBase 云托管（CloudRun），配置 `MONGODB_URI` / `JWT_SECRET` / `WX_APPID` / `WX_SECRET`
- 上传小程序代码包 ≠ 部署到 CloudBase；前端走微信公众平台，后端与数据库走 CloudBase
