# 题库数据 · 微信云开发数据库导入包

> 目标数据库：**微信云开发数据库（CloudBase NoSQL，文档型）**
> 数据格式：**JSON Lines（JSONL）** — 每行一条 JSON 记录，集合名 = 文件名
> 生成工具：`scripts/export_cloudbase/export.py`（从 11 个 Word 题库自动解析）

---

## 一、数据文件清单

| 文件 | 对应集合 | 记录数 | 说明 |
|------|---------|--------|------|
| `chapters.jsonl` | `chapters` | 12 | 章节（第7~18章） |
| `questions.jsonl` | `questions` | 598 | 题目（355单选 + 243多选） |
| `exams.jsonl` | `exams` | 12 | 每章一套模考卷，内嵌题目ID数组 |

**数据质量**：答案覆盖率 595/598（3 题因源文件双编号错位，已标记 `[待核对]` 并清空答案）；选项完整；答案字母越界 0；253 题带真题年份；11 题带法条溯源。

---

## 二、数据结构（文档型设计）

### chapters（章节集合）
```json
{
  "_id": "ch07",
  "chapter_no": 7,
  "title": "民法总论",
  "category": "民商法篇",
  "sort_order": 1,
  "duration_minutes": 60,
  "total_score": 100,
  "pass_score": 60,
  "single_count": 44,
  "multi_count": 33,
  "total_count": 77
}
```

### questions（题目集合）— 选项内嵌为对象
```json
{
  "_id": "q00001",
  "chapter_no": 7,
  "question_no": 1,
  "type": 1,
  "stem": "下列法律原则中，属于民法基本原则的是（ ）。",
  "options": { "A": "公信原则", "B": "公序良俗原则", "C": "公示原则", "D": "等价有偿原则" },
  "answer": "B",
  "analysis": "本题考核民法的基本原则...",
  "law_citation": "",
  "year_tag": "",
  "is_modified": 0
}
```
> 相比 MySQL 方案，`options` 直接内嵌为对象（非 JSON 字符串），读取时无需二次解析。

### exams（试卷集合）— 题目ID内嵌为数组
```json
{
  "_id": "exam07",
  "chapter_id": "ch07",
  "chapter_no": 7,
  "title": "第七章 民法总论 真题模拟",
  "duration_minutes": 60,
  "total_score": 100,
  "pass_score": 60,
  "question_ids": ["q00001", "q00002", "..."],
  "question_count": 77,
  "mode": 1,
  "status": 1
}
```

---

## 三、导入步骤（微信开发者工具 / 云开发控制台）

1. 打开**微信开发者工具** → 云开发控制台 → **数据库** → 选择对应环境。
2. 依次创建 3 个集合：`chapters`、`questions`、`exams`（集合名必须完全一致）。
3. 选中集合 → 顶部 **导入** → 上传对应 `.jsonl` 文件：
   - 导入方式选择 **JSON**
   - 冲突处理选择 **Insert**（首次导入；若重复导入选 Upsert 并确认 `_id`）
   - 编码 UTF-8
4. 导入顺序建议：`chapters` → `questions` → `exams`（exams 引用了 question_id）。
5. 导入完成后抽查：集合 `questions` 记录数应为 598。

---

## 四、集合权限配置（重要）

在云开发控制台 → 数据库 → 集合权限，建议：

| 集合 | 权限 | 原因 |
|------|------|------|
| `chapters` | 所有用户可读，仅管理端可写 | 公开章节数据 |
| `questions` | **所有用户不可读**（仅云函数可读） | ⚠️ 保护答案！客户端直接读集合会暴露 `answer` 字段 |
| `exams` | 所有用户不可读，仅云函数可读 | 同上 |

> **安全提醒**：`questions` 集合含 `answer`/`analysis` 字段。若设为客户端可读，用户可在小程序里直接查到答案。
> 正确做法是**通过云函数代理查询**：答题时云函数只返回题干+选项（剔除 answer），交卷时云函数在服务端判分。这与设计方案中的后端判分逻辑一致。

---

## 五、与 Golang 后端的关系

当前提供两套后端方案的数据：

- **Golang + MySQL**（`backend/`）：用 `scripts/import_questions/` 导入 MySQL，表结构见 `backend/migrations/`。
- **微信云开发数据库**（本目录）：用本 JSONL 包导入，适合纯云开发（云函数代替 Golang 后端）。

如果最终选择云开发数据库 + 云函数方案，则 `questions` 等集合由云函数读写，`answer` 字段不会下发到客户端。Golang 后端方案可作为独立部署的备选。

---

## 六、重新生成数据

```bash
python3 scripts/export_cloudbase/export.py \
  "/Users/revivekwan/Library/Containers/com.tencent.xinWeChat/Data/Documents/xwechat_files/a32421002_b7d2/msg/file/2026-11" \
  "accounting-exam/database/cloudbase"
```

解析规则针对《磨笔头习题》格式（`序号.【题型】` + `序号、「正确答案」`）。如题库格式变动，调整脚本顶部的正则常量 `Q_RE` / `ANS_RE` / `OPT_RE`。

---

## 七、待核对题目（3 题）

| ID | 章节 | 原题 | 问题 |
|----|------|------|------|
| q00577 | 14 | 下列关于电子商务的说法正确的是 | 原答案 CDE 超出选项 A~D |
| q00585 | 14 | 根据电子商务法律制度关于经营者登记的规定 | 原答案 BCDE 超出选项 A~D |
| q00591 | 14 | 关于电子认证类型，下列表述正确的是 | 原答案 CE 超出选项 A~D |

原因：第14/15章合并文档采用"题型内重复编号 + 答案区连续编号"双轨制，导致 3 题答案错位。已在数据中清空答案并标记 `[待核对]`，请对照原始 Word 人工补录。其余 595 题答案完整准确。
