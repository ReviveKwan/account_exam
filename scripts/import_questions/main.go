package main

import (
	"archive/zip"
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// docx 章节映射：文件名关键词 → 章节号
var chapterMapping = []struct {
	keyword   string
	chapterNo int
}{
	{"第七章", 7}, {"第八章", 8}, {"第九章", 9}, {"第十章", 10},
	{"第十一章", 11}, {"第十二章", 12}, {"第十三章", 13},
	{"第十四章", 14}, {"第十五章", 15}, {"第十六章", 16},
	{"第十七章", 17}, {"第十八章", 18},
}

func main() {
	dir := flag.String("dir", "", "Word文档目录")
	dsn := flag.String("dsn", "root:root123@tcp(localhost:3306)/exam_db?charset=utf8mb4&parseTime=True&loc=Local", "MySQL DSN")
	dryRun := flag.Bool("dry-run", false, "仅解析不写入数据库")
	flag.Parse()

	if *dir == "" {
		fmt.Println("用法: import_tool -dir <Word目录> [-dsn <数据库连接>] [-dry-run]")
		os.Exit(1)
	}

	// 遍历 docx 文件
	files, _ := filepath.Glob(filepath.Join(*dir, "*.docx"))
	sort.Strings(files)
	fmt.Printf("找到 %d 个 Word 文档\n", len(files))

	// 按章节汇总
	chapterQuestions := make(map[int][]ParsedQuestion)

	for _, file := range files {
		chapterNo := detectChapter(filepath.Base(file))
		if chapterNo == 0 {
			fmt.Printf("  [跳过] 无法识别章节: %s\n", filepath.Base(file))
			continue
		}

		text, err := extractDocxText(file)
		if err != nil {
			fmt.Printf("  [错误] 读取失败 %s: %v\n", filepath.Base(file), err)
			continue
		}

		questions, err := ParseDocument(text)
		if err != nil {
			fmt.Printf("  [错误] 解析失败 %s: %v\n", filepath.Base(file), err)
			continue
		}

		chapterQuestions[chapterNo] = append(chapterQuestions[chapterNo], questions...)
		fmt.Printf("  [OK] 第%d章 %s → %d 题\n", chapterNo, filepath.Base(file), len(questions))
	}

	if *dryRun {
		printDryRun(chapterQuestions)
		return
	}

	// 写入数据库
	db, err := sql.Open("mysql", *dsn)
	if err != nil {
		fmt.Printf("数据库连接失败: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		fmt.Printf("数据库不可用: %v\n", err)
		os.Exit(1)
	}

	totalImported := 0
	for chapterNo, questions := range chapterQuestions {
		chapterID := getOrCreateChapter(db, chapterNo)
		ids := importQuestions(db, chapterID, questions)
		updateExamQuestionIDs(db, chapterID, ids)
		totalImported += len(ids)
		fmt.Printf("  [DB] 第%d章导入 %d 题, exam题目ID已回填\n", chapterNo, len(ids))
	}

	fmt.Printf("\n✅ 导入完成，共 %d 题\n", totalImported)
}

// extractDocxText 从 docx 提取纯文本（解压 word/document.xml）
func extractDocxText(path string) (string, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return "", err
	}
	defer r.Close()

	for _, f := range r.File {
		if f.Name != "word/document.xml" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return "", err
		}
		defer rc.Close()

		data, err := io.ReadAll(rc)
		if err != nil {
			return "", err
		}
		return parseDocumentXML(data), nil
	}
	return "", fmt.Errorf("未找到 word/document.xml")
}

// parseDocumentXML 从 document.xml 提取段落文本
func parseDocumentXML(data []byte) string {
	type textNode struct {
		Text string `xml:",chardata"`
	}
	var doc struct {
		Body struct {
			Paragraphs []struct {
				Texts []textNode `xml:"r>t"`
			} `xml:"p"`
		} `xml:"body"`
	}
	xml.Unmarshal(data, &doc)

	var lines []string
	for _, p := range doc.Body.Paragraphs {
		var sb strings.Builder
		for _, t := range p.Texts {
			sb.WriteString(t.Text)
		}
		line := strings.TrimSpace(sb.String())
		if line != "" {
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, "\n")
}

func detectChapter(filename string) int {
	for _, m := range chapterMapping {
		if strings.Contains(filename, m.keyword) {
			return m.chapterNo
		}
	}
	return 0
}

func getOrCreateChapter(db *sql.DB, chapterNo int) int64 {
	var id int64
	err := db.QueryRow("SELECT id FROM chapters WHERE chapter_no = ?", chapterNo).Scan(&id)
	if err == nil {
		return id
	}
	result, _ := db.Exec("INSERT INTO chapters (chapter_no, title, duration_minutes, total_score, pass_score, sort_order) VALUES (?, ?, 60, 100, 60, ?)",
		chapterNo, fmt.Sprintf("第%d章", chapterNo), chapterNo)
	id, _ = result.LastInsertId()
	return id
}

func importQuestions(db *sql.DB, chapterID int64, questions []ParsedQuestion) []int64 {
	stmt, err := db.Prepare(`INSERT INTO questions
		(chapter_id, question_no, type, stem, options, answer, analysis, law_citation, year_tag, is_modified)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		fmt.Printf("  [错误] 准备语句失败: %v\n", err)
		return nil
	}
	defer stmt.Close()

	ids := make([]int64, 0, len(questions))
	for _, q := range questions {
		optJSON, _ := json.Marshal(q.Options)
		modified := 0
		if q.IsModified {
			modified = 1
		}
		result, err := stmt.Exec(chapterID, q.QuestionNo, q.Type, q.Stem, string(optJSON),
			q.Answer, q.Analysis, q.LawCitation, q.YearTag, modified)
		if err != nil {
			fmt.Printf("  [错误] 插入第%d题失败: %v\n", q.QuestionNo, err)
			continue
		}
		id, _ := result.LastInsertId()
		ids = append(ids, id)
	}
	return ids
}

func updateExamQuestionIDs(db *sql.DB, chapterID int64, ids []int64) {
	if len(ids) == 0 {
		return
	}
	idsJSON, _ := json.Marshal(ids)
	db.Exec("UPDATE exams SET question_ids = ? WHERE chapter_id = ?", string(idsJSON), chapterID)
}

func printDryRun(chapterQuestions map[int][]ParsedQuestion) {
	fmt.Println("\n===== DRY RUN 预览 =====")
	keys := make([]int, 0, len(chapterQuestions))
	for k := range chapterQuestions {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	for _, k := range keys {
		qs := chapterQuestions[k]
		fmt.Printf("\n第%d章: %d 题\n", k, len(qs))
		if len(qs) > 0 {
			fmt.Println(qs[0].ToJSON())
		}
	}
}

var _ = strconv.Itoa
