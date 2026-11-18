package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// ParsedQuestion 解析后的题目结构
type ParsedQuestion struct {
	QuestionNo  int               `json:"question_no"`
	Type        int               `json:"type"` // 1=单选 2=多选
	Stem        string            `json:"stem"`
	Options     map[string]string `json:"options"`
	Answer      string            `json:"answer"`
	Analysis    string            `json:"analysis"`
	LawCitation string            `json:"law_citation"`
	YearTag     string            `json:"year_tag"`
	IsModified  bool              `json:"is_modified"`
}

var (
	// 题目正则: 序号、【单选题/多选题】（年份）（改）题干 + 选项
	questionRegex = regexp.MustCompile(
		`(\d+)\s*【(单选题|多选题|判断题)】(?:（(\d{4})年)）?(?:（改）)?\s*([\s\S]*?)(?=(?:\d+\s*【)|$)`,
	)
	// 选项正则: A. 选项内容
	optionRegex = regexp.MustCompile(
		`(?:^|\s)([A-Z])\s*[.．、]\s*(.*?)(?=\s+[A-Z]\s*[.．、]|$)`,
	)
	// 答案正则: 序号、「正确答案」 X「答案解析」 解析内容
	answerRegex = regexp.MustCompile(
		`(\d+)、「正确答案」\s*([A-Z]+)\s*「答案解析」\s*([\s\S]*?)(?=\d+、「正确答案」|$)`,
	)
	// 法条正则: 《...》第X条
	lawRegex = regexp.MustCompile(`(《[^》]+》第[^\s，。；]+条(?:至第?[^\s，。；]+条)?)`)
)

// ParseDocument 解析单个文档的纯文本，返回题目列表
func ParseDocument(text string) ([]ParsedQuestion, error) {
	// 归一化文本
	text = normalizeText(text)

	// 先解析答案区
	answers := parseAnswers(text)

	// 解析题目区
	qMatches := questionRegex.FindAllStringSubmatch(text, -1)
	questions := make([]ParsedQuestion, 0, len(qMatches))

	for _, m := range qMatches {
		q := ParsedQuestion{
			QuestionNo:  atoi(m[1]),
			Type:        typeToInt(m[2]),
			YearTag:     m[3],
			IsModified:  strings.Contains(m[0], "（改）"),
			Options:     make(map[string]string),
		}

		// 提取题干和选项
		body := strings.TrimSpace(m[4])
		q.Stem, q.Options = splitStemAndOptions(body)

		// 关联答案
		if ans, ok := answers[q.QuestionNo]; ok {
			q.Answer = ans.answer
			q.Analysis = ans.analysis
			q.LawCitation = extractLawCitation(ans.analysis)
		}

		questions = append(questions, q)
	}

	return questions, nil
}

type answerInfo struct {
	answer  string
	analysis string
}

func parseAnswers(text string) map[int]answerInfo {
	result := make(map[int]answerInfo)
	matches := answerRegex.FindAllStringSubmatch(text, -1)
	for _, m := range matches {
		result[atoi(m[1])] = answerInfo{
			answer:   strings.TrimSpace(m[2]),
			analysis: strings.TrimSpace(m[3]),
		}
	}
	return result
}

// splitStemAndOptions 分离题干和选项
func splitStemAndOptions(body string) (string, map[string]string) {
	options := make(map[string]string)
	optMatches := optionRegex.FindAllStringSubmatch(body, -1)

	stem := body
	if len(optMatches) > 0 {
		// 题干 = 第一个选项之前的内容
		firstOptIdx := strings.Index(body, optMatches[0][0])
		if firstOptIdx > 0 {
			stem = strings.TrimSpace(body[:firstOptIdx])
		}
		// 收集选项
		for _, om := range optMatches {
			key := strings.TrimSpace(om[1])
			val := strings.TrimSpace(om[2])
			options[key] = val
		}
	}

	return cleanText(stem), options
}

// extractLawCitation 从解析中提取法条引用
func extractLawCitation(analysis string) string {
	m := lawRegex.FindString(analysis)
	if m != "" {
		return m
	}
	return ""
}

// normalizeText 归一化文本
func normalizeText(text string) string {
	// 统一换行
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	// 去除多余空白行
	lines := strings.Split(text, "\n")
	var cleaned []string
	for _, line := range lines {
		line = strings.TrimRightFunc(line, unicode.IsSpace)
		if line != "" {
			cleaned = append(cleaned, line)
		}
	}
	return strings.Join(cleaned, "\n")
}

func cleanText(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\u00a0", " ") // 不间断空格
	s = regexp.MustCompile(`\s+`).ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

func typeToInt(s string) int {
	switch s {
	case "单选题":
		return 1
	case "多选题":
		return 2
	default:
		return 1
	}
}

func atoi(s string) int {
	var n int
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		}
	}
	return n
}

// ToJSON 用于调试
func (q ParsedQuestion) ToJSON() string {
	b, _ := json.MarshalIndent(q, "", "  ")
	return fmt.Sprintf("%s", b)
}
