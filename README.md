# 法考磨笔头 · 税务师涉税法律刷题小程序

> 基于 V2.0 高保真 UI 稿落地的完整项目：微信小程序前端 + Golang 后端 + **文档型 NoSQL 数据库（微信云开发数据库 / 腾讯云 MongoDB）**，部署于微信云托管（CloudBase）。

---

## 一、本地项目在哪里

项目根目录：

```
/Users/revivekwan/WorkBuddy AI/2026-11-17-13-13-53/
```

| 路径 | 说明 |
|------|------|
| `会计刷题小程序设计方案.md` | 设计方案 V2.0 |
| `开发文档.md` | 开发文档 V2.0 |
| `UI原型设计说明.md` | UI 原型设计说明 |
| `高保真UI稿.html` | V1 高保真稿 |
| `accounting-exam/` | **完整可运行项目代码（本次交付）** |

---

## 二、项目目录结构

```
accounting-exam/
├── miniprogram/                        # 微信小程序前端（V2.0 UI稿）
│   ├── app.js / app.json / app.wxss
│   ├── project.config.json
│   ├── pages/                          # 6 页面（login/exam-list/exam-page/exam-result/wrong-book/profile）
│   ├── components/                     # answer-sheet / submit-confirm
│   └── utils/                          # request / auth / exam-state
│
├── backend/                            # Golang 后端（Gin + MongoDB Driver）
│   ├── cmd/server/main.go              # 程序入口（连接NoSQL+建集合+索引）
│   ├── config/config.yaml
│   ├── internal/
│   │   ├── config/                     # 环境变量配置（MONGODB_URI 等）
│   │   ├── router/                     # 路由
│   │   ├── handler/                    # HTTP 处理器
│   │   ├── service/                    # 业务逻辑 + 判分引擎
│   │   ├── repository/                 # MongoDB 数据访问层
│   │   └── model/                      # BSON 文档模型
│   ├── pkg/
│   │   ├── database/                   # MongoDB 连接 + 索引
│   │   ├── response/                   # 统一响应
│   │   └── middleware/                 # JWT + CORS
│   ├── migrations/001_init.js          # NoSQL 集合/索引初始化脚本
│   └── go.mod / go.sum
│
├── database/cloudbase/                 # 题库导入包（JSONL，可直接导入云数据库）
│   ├── chapters.jsonl / questions.jsonl / exams.jsonl
│   └── README.md                       # 导入步骤 + 权限 + 安全提醒
│
├── scripts/
│   ├── import_questions/               # Word → 旧 MySQL 导入工具（保留兼容，可弃用）
│   └── export_cloudbase/export.py      # Word → JSONL 导出工具（当前使用）
│
└── deploy/
    ├── Dockerfile / docker-compose.yml # 用 MongoDB 7.0
    └── container.config.json           # 云托管配置（MONGODB_URI）
```

---

## 三、数据库设计（文档型 NoSQL）

### 集合（Collection）一览

| 集合 | 说明 | 关键字段 |
|------|------|---------|
| `users` | 用户 | `_id`(u_xxx), openid, nickname, is_guest |
| `chapters` | 章节 | `_id`(ch07), chapter_no, title, category |
| `questions` | 题目 | `_id`(q00001), options **内嵌对象**, answer, analysis, law_citation |
| `exams` | 试卷 | `_id`(exam07), question_ids **内嵌数组**, chapter_id |
| `exam_records` | 考试记录 | `_id`, answers **内嵌子文档数组**, status, remaining_seconds |
| `wrong_questions` | 错题本 | `_id`, user_id, question_id, correct_streak, status |

### 文档型设计要点（对比原 MySQL）
- **无表结构、无 JOIN**：options 内嵌为对象，answers 内嵌为数组，读取零解析
- **字符串 _id**：`ch07` / `q00001` / `exam07`，集合间直接引用
- **保留关系型等价索引**：openid 唯一、chapter_no+question_no 复合、user_id+exam_id+status 复合
- **聚合管道**替代 SQL GROUP BY（统计掌握度、及格章节数等）

> 库表结构与 `database/cloudbase/*.jsonl` 完全对齐，导入即可用。

---

## 四、快速启动

### 4.1 启动后端（本地，MongoDB）

```bash
cd accounting-exam
docker-compose -f deploy/docker-compose.yml up -d
# MongoDB 27017 + 后端 8080
curl http://localhost:8080/health   # {"status":"ok"}
```

或本地已有 MongoDB：

```bash
cd accounting-exam/backend
export MONGODB_URI="mongodb://localhost:27017"
export MONGODB_DB="exam_db"
go run cmd/server/main.go
```

### 4.2 导入题库（关键）

后端启动时会检查题库，空库会提示。导入方式见 `database/cloudbase/README.md`：
- **微信云开发数据库**：控制台导入 3 个 JSONL 文件到对应集合
- **本地/腾讯云 MongoDB**：
  ```bash
  mongosh "mongodb://localhost:27017/exam_db" --file backend/migrations/001_init.js
  mongoimport --uri "mongodb://localhost:27017/exam_db" --collection questions --file database/cloudbase/questions.jsonl
  mongoimport --uri "mongodb://localhost:27017/exam_db" --collection chapters --file database/cloudbase/chapters.jsonl
  mongoimport --uri "mongodb://localhost:27017/exam_db" --collection exams --file database/cloudbase/exams.jsonl
  ```

### 4.3 运行小程序前端

1. 微信开发者工具 → 导入项目 → 选 `accounting-exam/miniprogram`
2. 本地调试：`utils/request.js` 的 `USE_CLOUD=false`，`app.js` 的 `apiBase` 指向后端
3. 云托管：`USE_CLOUD=true`，`envId` 填云环境 ID

---

## 五、关键技术点

| 项 | 实现 |
|----|------|
| 数据库 | 文档型 NoSQL（MongoDB Driver），内嵌文档设计 |
| 前端 → 后端 | `wx.cloud.callContainer`（云托管）/ `wx.request`（本地） |
| 登录 | wx.login → JWT（user_id 为文档 _id 字符串）；支持游客试刷 |
| 判分 | 单选精确匹配；多选全对满分、**少选每项 0.5 分**、错选 0 分 |
| 断点续答 | `exam_records` 存 remaining_seconds + current_question_no + answers 内嵌数组 |
| 错题本 | 答错入池，连续答对 2 次 status=2 移出 |
| 安全 | questions 集合仅云函数/后端可读，answer 不下发客户端（模考模式） |

---

## 六、已完成功能对照 V2.0 UI 稿

| 屏幕 | 页面 | 状态 |
|------|------|------|
| 01 登录页 | pages/login | ✅ |
| 02 题库首页 | pages/exam-list | ✅ |
| 03 答题页 | pages/exam-page | ✅ |
| 04 答题卡抽屉 | components/answer-sheet | ✅ |
| 05 交卷确认 | components/submit-confirm | ✅ |
| 06 成绩诊断 | pages/exam-result | ✅ |
| 07 错题攻坚 | pages/wrong-book | ✅ |
| — 我的学情 | pages/profile | ✅ |

后端（MongoDB 版）`go build ./...` 与 `go vet ./...` 均通过。

---

## 七、待办 / 注意事项

1. **连接信息**：在 `MONGODB_URI`（`deploy/container.config.json` / 环境变量）填入真实地址；微信云开发数据库用内网连接或云函数代理。
2. **题库导入**：598 题需先导入（见 4.2），其中 3 题因源文档双编号已标记 `[待核对]`。
3. **微信登录**：`WX_APPID`/`WX_SECRET` 填入后，`user.go` 的 `code2OpenID` 需补上 jscode2session 真实调用（当前开发模式用 mock）。
4. **答案保护**：云数据库 `questions` 集合权限务必设为"仅云函数可读"，防止客户端直查答案。
