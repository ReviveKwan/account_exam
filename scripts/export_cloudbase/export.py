#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
题库导出工具 - 生成微信云开发数据库(CloudBase NoSQL)导入文件
从 11 个 Word 文档解析题目，输出 JSON Lines 格式：
  chapters.jsonl  - 章节集合
  questions.jsonl - 题目集合
  exams.jsonl     - 考试集合
每行一条 JSON 记录，微信云开发数据库导入时"导入方式"选 JSON，冲突处理选 Insert。
"""
import zipfile
import re
import html
import json
import os
import sys
from collections import OrderedDict

# 章节映射：文件名关键词 -> (章节号, 章节标题, 分类, 排序)
CHAPTER_MAP = [
    ("第七章", 7, "民法总论", "民商法篇", 1),
    ("第八章", 8, "物权法律制度", "民商法篇", 2),
    ("第九章", 9, "债权法律制度", "民商法篇", 3),
    ("第十章", 10, "婚姻家庭和继承法", "民商法篇", 4),
    ("第十一章", 11, "个人独资企业法", "企业法篇", 5),
    ("第十二章", 12, "合伙企业法", "企业法篇", 6),
    ("第十三章", 13, "公司法", "企业法篇", 7),
    ("第十四章", 14, "破产法", "程序法篇", 8),
    ("第十五章", 15, "电子商务法", "程序法篇", 9),
    ("第十六章", 16, "民事诉讼法", "程序法篇", 10),
    ("第十七章", 17, "刑法", "刑事法篇", 11),
    ("第十八章", 18, "刑事诉讼法", "刑事法篇", 12),
]

# 用于匹配文件名的章节（注意"第十四章/第十五章"合并在一个文件）
def detect_chapters(filename):
    found = []
    for kw, no, title, cat, order in CHAPTER_MAP:
        if kw in filename:
            found.append((no, title, cat, order))
    return found


def extract_text(docx_path):
    """从 docx 提取纯文本"""
    with zipfile.ZipFile(docx_path) as z:
        xml = z.read("word/document.xml").decode("utf-8", errors="ignore")
    # 提取所有 w:t 内容，段落用换行连接
    parts = re.split(r"</w:p>", xml)
    lines = []
    for part in parts:
        texts = re.findall(r"<w:t[^>]*>(.*?)</w:t>", part, re.S)
        line = "".join(texts)
        line = html.unescape(line)
        line = line.replace("\u3000", " ").replace("\xa0", " ")
        line = re.sub(r"\s+", " ", line).strip()
        if line:
            lines.append(line)
    return "\n".join(lines)


# 题目正则：序号. 【题型】 （年份）（改）题干 ... 选项
Q_RE = re.compile(
    r"(\d+)\s*[、.]\s*【(单选题|多选题)】"
    r"(?:（(\d{4})年）)?\s*(?:（改）)?\s*"
    r"(.*?)(?=\d+\s*[、.]\s*【(?:单选题|多选题)】|习题参考答案|参考答案：|$)",
    re.S,
)

# 选项正则：A. xxx B. xxx ...（选项在题干文本内，A~L）
OPT_RE = re.compile(r"(?:^|\s)([A-L])\s*[.．、]\s*")

# 答案正则：序号、 「正确答案」 X 「答案解析」 解析
ANS_RE = re.compile(
    r"(\d+)\s*、\s*「正确答案」\s*([A-Z]+)\s*「答案解析」\s*"
    r"(.*?)(?=\d+\s*、\s*「正确答案」|$)",
    re.S,
)

# 法条正则
LAW_RE = re.compile(r"(《[^》]+》第[^\s，。；、]+条(?:至第?[^\s，。；、]+条)?)")


def parse_answers(full_text):
    """解析答案区，返回按顺序的答案列表 [{answer, analysis, law}]。
    注意：部分章节题目按题型内编号（多选从1重新开始），题号会重复，
    因此用出现顺序而非题号关联。"""
    answers = []
    for m in ANS_RE.finditer(full_text):
        answer = m.group(2).strip()
        analysis = re.sub(r"\s+", " ", m.group(3)).strip()
        law_m = LAW_RE.search(analysis)
        law = law_m.group(1) if law_m else ""
        answers.append({"answer": answer, "analysis": analysis, "law": law})
    return answers


def parse_questions(full_text, answers, chapter_no):
    """解析题目区（按出现顺序关联答案）"""
    questions = []
    for idx, m in enumerate(Q_RE.finditer(full_text)):
        no = int(m.group(1))
        qtype = 1 if m.group(2) == "单选题" else 2
        year = m.group(3) or ""
        body = m.group(4).strip()
        is_modified = "（改）" in m.group(0)

        # 找第一个选项位置，之前是题干
        opt_matches = list(OPT_RE.finditer(body))
        if opt_matches:
            first = opt_matches[0].start()
            stem = body[:first].strip()
            options = OrderedDict()
            for i, om in enumerate(opt_matches):
                key = om.group(1)
                start = om.end()
                end = opt_matches[i + 1].start() if i + 1 < len(opt_matches) else len(body)
                val = body[start:end].strip()
                # 去掉尾部残留的下一题序号（如 "等价有偿原则 2、"）
                val = re.sub(r"\s+\d+\s*[、.]\s*$", "", val).strip()
                val = re.sub(r"\s+\d+\s*[、.]\s*【.*$", "", val).strip()
                options[key] = val
        else:
            stem = body
            options = OrderedDict()

        stem = re.sub(r"\s+", " ", stem).strip()
        # 按出现顺序关联答案（题号在题型内会重复，不能按题号）
        ans = answers[idx] if idx < len(answers) else {}
        # 校验：答案字母必须都在选项范围内，否则判定为错位，清空并记录
        opt_keys = set(options.keys())
        raw_answer = ans.get("answer", "")
        answer_valid = all(a in opt_keys for a in raw_answer) if raw_answer else False
        if raw_answer and not answer_valid:
            ans = dict(ans)
            ans["answer"] = ""
            ans["analysis"] = "[待核对] 原答案 " + raw_answer + " 与选项不符，疑似源文件答案错位，请人工核对"
        questions.append(
            OrderedDict([
                ("chapter_no", chapter_no),
                ("question_no", no),
                ("type", qtype),
                ("stem", stem),
                ("options", options),
                ("answer", ans.get("answer", "")),
                ("analysis", ans.get("analysis", "")),
                ("law_citation", ans.get("law", "")),
                ("year_tag", year),
                ("is_modified", 1 if is_modified else 0),
            ])
        )
    return questions


def main():
    docx_dir = sys.argv[1] if len(sys.argv) > 1 else "/Users/revivekwan/Library/Containers/com.tencent.xinWeChat/Data/Documents/xwechat_files/a32421002_b7d2/msg/file/2026-11"
    out_dir = sys.argv[2] if len(sys.argv) > 2 else "/Users/revivekwan/WorkBuddy AI/2026-11-17-13-13-53/accounting-exam/database/cloudbase"
    os.makedirs(out_dir, exist_ok=True)

    files = sorted([f for f in os.listdir(docx_dir) if f.lower().endswith(".docx")])
    print(f"找到 {len(files)} 个 Word 文档")

    all_questions = []
    chapter_stats = OrderedDict()

    for fname in files:
        chapters = detect_chapters(fname)
        if not chapters:
            print(f"  [跳过] 无法识别章节: {fname}")
            continue
        try:
            text = extract_text(os.path.join(docx_dir, fname))
        except Exception as e:
            print(f"  [错误] {fname}: {e}")
            continue

        answers = parse_answers(text)
        # 文件名含多个章节时，题目都归到主章节（第一个），拆分需人工
        main_ch = chapters[0]
        ch_no = main_ch[0]
        qs = parse_questions(text, answers, ch_no)
        # 校验：答案覆盖率
        matched = sum(1 for q in qs if q["answer"])
        all_questions.extend(qs)
        ch_stats = chapter_stats.setdefault(ch_no, {"single": 0, "multi": 0, "no_answer": 0})
        for q in qs:
            if q["type"] == 1:
                ch_stats["single"] += 1
            else:
                ch_stats["multi"] += 1
            if not q["answer"]:
                ch_stats["no_answer"] += 1
        extra = f" (含第{chapters[1][0]}章)" if len(chapters) > 1 else ""
        print(f"  [OK] 第{ch_no}章{extra}: 解析 {len(qs)} 题, 有答案 {matched}/{len(qs)}")

    # 写 questions.jsonl（文档型：options 内嵌为对象）
    q_path = os.path.join(out_dir, "questions.jsonl")
    qid_by_ch = OrderedDict()
    with open(q_path, "w", encoding="utf-8") as f:
        for idx, q in enumerate(all_questions, 1):
            qid = f"q{idx:05d}"
            q["_id"] = qid
            qid_by_ch.setdefault(q["chapter_no"], []).append(qid)
            # 云开发自动生成 _id，这里用自定义 _id 便于关联
            rec = OrderedDict([
                ("_id", qid),
                ("chapter_no", q["chapter_no"]),
                ("question_no", q["question_no"]),
                ("type", q["type"]),
                ("stem", q["stem"]),
                ("options", q["options"]),   # 内嵌对象，非JSON字符串
                ("answer", q["answer"]),
                ("analysis", q["analysis"]),
                ("law_citation", q["law_citation"]),
                ("year_tag", q["year_tag"]),
                ("is_modified", q["is_modified"]),
            ])
            f.write(json.dumps(rec, ensure_ascii=False) + "\n")

    # 写 chapters.jsonl
    ch_path = os.path.join(out_dir, "chapters.jsonl")
    with open(ch_path, "w", encoding="utf-8") as f:
        for kw, no, title, cat, order in CHAPTER_MAP:
            st = chapter_stats.get(no, {"single": 0, "multi": 0})
            rec = OrderedDict([
                ("_id", f"ch{no:02d}"),
                ("chapter_no", no),
                ("title", title),
                ("category", cat),
                ("sort_order", order),
                ("duration_minutes", 60),
                ("total_score", 100),
                ("pass_score", 60),
                ("single_count", st["single"]),
                ("multi_count", st["multi"]),
                ("total_count", st["single"] + st["multi"]),
            ])
            f.write(json.dumps(rec, ensure_ascii=False) + "\n")

    # 写 exams.jsonl：每章题目 _id 列表（qid_by_ch 已在写题目时构建）
    ex_path = os.path.join(out_dir, "exams.jsonl")
    with open(ex_path, "w", encoding="utf-8") as f:
        for kw, no, title, cat, order in CHAPTER_MAP:
            qids = qid_by_ch.get(no, [])
            rec = OrderedDict([
                ("_id", f"exam{no:02d}"),
                ("chapter_id", f"ch{no:02d}"),
                ("chapter_no", no),
                ("title", f"第{kw} {title} 真题模拟"),
                ("duration_minutes", 60),
                ("total_score", 100),
                ("pass_score", 60),
                ("question_ids", qids),   # 内嵌数组
                ("question_count", len(qids)),
                ("mode", 1),
                ("status", 1),
            ])
            f.write(json.dumps(rec, ensure_ascii=False) + "\n")

    print(f"\n总计解析 {len(all_questions)} 题")
    print(f"输出文件：")
    print(f"  {q_path}")
    print(f"  {ch_path}")
    print(f"  {ex_path}")


if __name__ == "__main__":
    main()
