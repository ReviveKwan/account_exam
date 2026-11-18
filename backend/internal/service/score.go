package service

import (
	"accounting-exam/internal/model"
	"accounting-exam/internal/repository"
	"fmt"
	"sort"
	"strings"
	"time"
)

type ScoreService struct {
	recordRepo   *repository.RecordRepo
	questionRepo *repository.QuestionRepo
	wrongRepo    *repository.WrongQuestionRepo
}

func NewScoreService(recordRepo *repository.RecordRepo, questionRepo *repository.QuestionRepo, wrongRepo *repository.WrongQuestionRepo) *ScoreService {
	return &ScoreService{recordRepo: recordRepo, questionRepo: questionRepo, wrongRepo: wrongRepo}
}

// SubmitAnswers 提交考试并判分
func (s *ScoreService) SubmitAnswers(userID, recordID string, durationSeconds int, answers []AnswerSubmission) (*SubmitResult, error) {
	record, err := s.recordRepo.GetByID(recordID)
	if err != nil {
		return nil, fmt.Errorf("考试记录不存在")
	}
	if record.Status != 0 {
		return nil, fmt.Errorf("考试已提交，无法重复操作")
	}
	if record.UserID != userID {
		return nil, fmt.Errorf("无权操作此记录")
	}

	qIDs := make([]string, len(answers))
	for i, a := range answers {
		qIDs[i] = a.QuestionID
	}
	questions, err := s.questionRepo.GetByIDs(qIDs)
	if err != nil {
		return nil, fmt.Errorf("获取题目失败")
	}
	qMap := make(map[string]*model.Question)
	for i := range questions {
		qMap[questions[i].ID] = &questions[i]
	}

	var details []model.AnswerDetail
	var correctCount, partialCount, wrongCount, unansweredCount int
	var totalScore float64

	for _, a := range answers {
		q, ok := qMap[a.QuestionID]
		if !ok {
			continue
		}
		detail := model.AnswerDetail{
			QuestionID: a.QuestionID, UserAnswer: a.UserAnswer, IsMarked: boolToInt(a.IsMarked),
		}
		if strings.TrimSpace(a.UserAnswer) == "" {
			detail.IsAnswered, detail.IsCorrect, detail.Score = 0, 0, 0
			unansweredCount++
		} else {
			detail.IsAnswered = 1
			isCorrect, isPartial, score := s.judgeQuestion(q, a.UserAnswer)
			detail.Score = score
			if isCorrect {
				detail.IsCorrect = 1
				correctCount++
				totalScore += score
			} else if isPartial {
				detail.IsCorrect = 2
				partialCount++
				totalScore += score
			} else {
				detail.IsCorrect = 0
				wrongCount++
				s.wrongRepo.Upsert(userID, q.ID, q.ChapterNo)
			}
		}
		details = append(details, detail)
	}

	now := time.Now()
	record.Score = totalScore
	record.CorrectCount = correctCount
	record.PartialCount = partialCount
	record.WrongCount = wrongCount
	record.UnansweredCount = unansweredCount
	record.DurationSeconds = durationSeconds
	record.Status = 1
	record.SubmittedAt = &now
	record.Answers = details

	if err := s.recordRepo.Update(record); err != nil {
		return nil, fmt.Errorf("保存考试记录失败: %v", err)
	}

	result := &SubmitResult{
		RecordID: recordID, Score: totalScore, TotalScore: record.TotalScore,
		Passed: totalScore >= float64(record.TotalScore)*0.6,
		CorrectCount: correctCount, PartialCount: partialCount,
		WrongCount: wrongCount, UnansweredCount: unansweredCount,
		DurationSeconds: durationSeconds,
		Details: s.buildDetails(details, qMap),
	}
	return result, nil
}

func (s *ScoreService) judgeQuestion(q *model.Question, userAnswer string) (bool, bool, float64) {
	if q.Type == 1 {
		isCorrect := strings.EqualFold(strings.TrimSpace(userAnswer), strings.TrimSpace(q.Answer))
		if isCorrect {
			return true, false, 1.5
		}
		return false, false, 0
	}
	return s.judgeMulti(userAnswer, q.Answer, 2.0)
}

// judgeMulti 多选题判分：全对满分，少选每项0.5分，错选0分
func (s *ScoreService) judgeMulti(userAnswer, correctAnswer string, fullScore float64) (bool, bool, float64) {
	userLetters := sortAnswerLetters(userAnswer)
	correctLetters := sortAnswerLetters(correctAnswer)
	if userLetters == correctLetters {
		return true, false, fullScore
	}
	correctMap := make(map[string]bool)
	for _, l := range strings.Split(correctLetters, "") {
		correctMap[l] = true
	}
	hasWrong := false
	correctSelected := 0
	for _, l := range strings.Split(userLetters, "") {
		if l == "" {
			continue
		}
		if correctMap[l] {
			correctSelected++
		} else {
			hasWrong = true
		}
	}
	if hasWrong || correctSelected == 0 {
		return false, false, 0
	}
	partial := float64(correctSelected) * 0.5
	if partial > fullScore {
		partial = fullScore
	}
	return false, true, partial
}

func (s *ScoreService) buildDetails(details []model.AnswerDetail, qMap map[string]*model.Question) []DetailResult {
	result := make([]DetailResult, 0, len(details))
	for _, d := range details {
		q := qMap[d.QuestionID]
		if q == nil {
			continue
		}
		result = append(result, DetailResult{
			QuestionID: d.QuestionID, Type: q.Type, UserAnswer: d.UserAnswer,
			CorrectAnswer: q.Answer, IsCorrect: d.IsCorrect, Score: d.Score,
			Analysis: q.Analysis, LawCitation: q.LawCitation,
			Stem: q.Stem, Options: q.Options,
		})
	}
	return result
}

// AnswerSubmission 前端提交的答案
type AnswerSubmission struct {
	QuestionID string `json:"question_id"`
	UserAnswer string `json:"user_answer"`
	IsMarked   bool   `json:"is_marked"`
}

type SubmitResult struct {
	RecordID        string         `json:"record_id"`
	Score           float64        `json:"score"`
	TotalScore      int            `json:"total_score"`
	Passed          bool           `json:"passed"`
	CorrectCount    int            `json:"correct_count"`
	PartialCount    int            `json:"partial_count"`
	WrongCount      int            `json:"wrong_count"`
	UnansweredCount int            `json:"unanswered_count"`
	DurationSeconds int            `json:"duration_seconds"`
	Details         []DetailResult `json:"details"`
}

type DetailResult struct {
	QuestionID    string            `json:"question_id"`
	Type          int               `json:"type"`
	UserAnswer    string            `json:"user_answer"`
	CorrectAnswer string            `json:"correct_answer"`
	IsCorrect     int               `json:"is_correct"`
	Score         float64           `json:"score"`
	Analysis      string            `json:"analysis"`
	LawCitation   string            `json:"law_citation"`
	Stem          string            `json:"stem"`
	Options       map[string]string `json:"options"`
}

func sortAnswerLetters(answer string) string {
	answer = strings.ToUpper(strings.TrimSpace(answer))
	if answer == "" {
		return ""
	}
	letters := strings.Split(answer, "")
	sort.Strings(letters)
	return strings.Join(letters, "")
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
