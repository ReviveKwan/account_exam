#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
把 database/cloudbase/*.jsonl 通过 tcb db nosql execute 导入到微信云开发数据库。
每次 insert 一批（默认 50 条），避免命令过长。
"""
import json
import subprocess
import sys
import os

DATA_DIR = sys.argv[1] if len(sys.argv) > 1 else "/Users/revivekwan/WorkBuddy AI/2026-11-17-13-13-53/accounting-exam/database/cloudbase"
BATCH = 50

COLLECTIONS = ["chapters", "exams", "questions"]


def run_insert(collection, docs):
    """构造 MgoCommands insert 命令并执行"""
    cmd_obj = {
        "insert": collection,
        "documents": docs
    }
    command_payload = [{
        "TableName": collection,
        "CommandType": "INSERT",
        "Command": json.dumps(cmd_obj, ensure_ascii=False, separators=(",", ":"))
    }]
    result = subprocess.run(
        ["tcb", "db", "nosql", "execute", "--command", json.dumps(command_payload, ensure_ascii=False, separators=(",", ":")), "--json"],
        capture_output=True, text=True,
        cwd="/Users/revivekwan/WorkBuddy AI/2026-11-17-13-13-53/accounting-exam"
    )
    if result.returncode != 0:
        return False, (result.stderr or result.stdout)[:300]
    # 成功标志: 结果含 "n":{"$numberInt":"..."} 且 ok 为 1.*（用正则避免被截断误判）
    import re
    ok_match = re.search(r'"ok":\{\s*"\$numberDouble":\s*"1\.', out)
    if result.returncode == 0 and ok_match and '"n":{' in out:
        return True, "ok"
    return False, (result.stdout or result.stderr)[:300]


def main():
    for coll in COLLECTIONS:
        path = os.path.join(DATA_DIR, coll + ".jsonl")
        docs = []
        with open(path, encoding="utf-8") as f:
            for line in f:
                line = line.strip()
                if line:
                    docs.append(json.loads(line))
        total = len(docs)
        imported = 0
        failed = 0
        print(f"\n=== 导入 {coll}: {total} 条 ===")
        for i in range(0, total, BATCH):
            batch = docs[i:i+BATCH]
            ok, msg = run_insert(coll, batch)
            if ok:
                imported += len(batch)
                print(f"  [{i//BATCH+1}] 已导入 {len(batch)} 条 (累计 {imported}/{total})")
            else:
                failed += len(batch)
                print(f"  [{i//BATCH+1}] ✗ 失败: {msg}")
        print(f"  → {coll}: 成功 {imported}, 失败 {failed}")
    print("\n导入完成")


if __name__ == "__main__":
    main()
