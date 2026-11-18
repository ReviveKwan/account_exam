// ============================================================
// 文档型数据库初始化脚本（MongoDB Shell / mongosh）
// 微信云开发数据库 或 腾讯云 MongoDB 均适用
//
// 用法:
//   mongosh "mongodb://localhost:27017/exam_db" --file 001_init.js
//
// 说明:
//   1. 集合结构由写入的文档自动决定（NoSQL 无表结构）
//   2. 题库数据请用 database/cloudbase/*.jsonl 导入（见该目录 README）
//   3. 本脚本仅创建集合与索引
// ============================================================

// 创建集合（若不存在）
const collections = [
  "users",           // 用户
  "chapters",        // 章节
  "questions",       // 题库
  "exams",           // 试卷
  "exam_records",    // 考试记录（answers 内嵌）
  "wrong_questions", // 错题本
];

const existing = db.getCollectionNames();
collections.forEach(function (name) {
  if (existing.indexOf(name) === -1) {
    db.createCollection(name);
    print("创建集合: " + name);
  }
});

// ---- 索引 ----

// users: openid 唯一
db.users.createIndex({ openid: 1 }, { unique: true });

// questions: 章节、题型、章节+题号
db.questions.createIndex({ chapter_no: 1 });
db.questions.createIndex({ type: 1 });
db.questions.createIndex({ chapter_no: 1, question_no: 1 });

// exams: 章节
db.exams.createIndex({ chapter_id: 1 });
db.exams.createIndex({ status: 1 });

// exam_records
db.exam_records.createIndex({ user_id: 1 });
db.exam_records.createIndex({ exam_id: 1 });
db.exam_records.createIndex({ user_id: 1, exam_id: 1, status: 1 });

// wrong_questions
db.wrong_questions.createIndex({ user_id: 1 });
db.wrong_questions.createIndex({ user_id: 1, chapter_no: 1, status: 1 });

print("✓ 集合与索引初始化完成");
print("题库文档数: " + db.questions.countDocuments({}));
