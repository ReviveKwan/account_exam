# 后端接入「云开发自带实例」说明

后端已改为通过 **CloudBase 文档库 HTTP API** 访问自带实例，**不再使用** `mongodb://` 直连。

## 1. 控制台准备

### 创建服务端 API Key

1. 打开 [云开发平台](https://tcb.cloud.tencent.com/) → 你的环境 `wechat-app`
2. 进入 **API Key 管理**
3. 创建 **服务端 API Key**（管理员权限，长期有效）
4. 复制 Key（只显示一次，妥善保存）

### 确认默认实例

文档型数据库 → 连接管理：当前使用 **云开发自带实例**  
后端环境变量 `CLOUDBASE_INSTANCE=(default)` 即指向该实例。

### 补齐业务集合

你已有：`chapters` / `exams` / `questions`  
还需新建（权限建议「仅管理端可读写」）：

| 集合 | 用途 |
|------|------|
| `users` | 用户 |
| `exam_records` | 考试记录 / 断点续答 |
| `wrong_questions` | 错题本 |

## 2. 云托管环境变量（必填）

| 变量 | 示例 | 说明 |
|------|------|------|
| `CLOUDBASE_ENV_ID` | `wechat-app-d6gjceiz966031f7f` | 环境 ID |
| `CLOUDBASE_APIKEY` | `你的服务端 Key` | **必填**，勿提交到 Git |
| `CLOUDBASE_INSTANCE` | `(default)` | 自带实例（括号要保留） |
| `CLOUDBASE_DATABASE` | `(default)` | 默认库 |
| `PORT` / `SERVER_PORT` | `80` | 与控制台「容器端口」一致 |
| `JWT_SECRET` | 随机长串 | JWT 密钥 |
| `WX_APPID` / `WX_SECRET` | 小程序凭证 | 正式登录用 |

控制台 **容器端口务必填 80**。

## 3. 重新部署

1. 用更新后的代码重新构建镜像并发布
2. 看容器日志应出现：`已连接云开发自带文档库（默认实例）` 和 `题库已就绪: xxx 题`
3. 若探活失败，优先检查 API Key 是否配错、集合名是否一致

## 4. 原理（一句话）

`Go 后端 → HTTPS + Bearer API Key → tcloudbasegateway → 云开发自带文档库(default)`  
不是 MongoDB 驱动，也不是 `localhost:27017`。
