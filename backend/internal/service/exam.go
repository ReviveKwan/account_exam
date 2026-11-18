package service

import (
	"accounting-exam/internal/model"
	"accounting-exam/internal/repository"
	"fmt"
)

type ExamService struct {
	examRepo     *repository.ExamRepo
	recordRepo   *repository.RecordRepo
	questionRepo *repository.QuestionRepo
	chapterRepo  *repository.ChapterRepo
}

func NewExamService(examRepo *repository.ExamRepo, recordRepo *repository.RecordRepo, questionRepo *repository.QuestionRepo, chapterRepo *repository.ChapterRepo) *ExamService {
	return &ExamService{examRepo: examRepo, recordRepo: recordRepo, questionRepo: questionRepo, chapterRepo: chapterRepo}
}

// ChapterWithStats 章节统计
type ChapterWithStats struct {
	ID               string   `json:"id"`
	ChapterNo        int      `json:"chapter_no"`
	Title            string   `json:"title"`
	Category         string   `json:"category"`
	SortOrder        int      `json:"sort_order"`
	DurationMinutes  int      `json:"duration_minutes"`
	TotalScore       int      `json:"total_score"`
	PassScore        int      `json:"pass_score"`
	ExamID           string   `json:"exam_id"`
	SingleCount      int      `json:"single_count"`
	MultiCount       int      `json:"multi_count"`
	TotalCount       int      `json:"total_count"`
	BestScore        float64  `json:"best_score"`
	HasResume        bool     `json:"has_resume"`
	ResumeQuestionNo int      `json:"resume_question_no"`
	ResumeRemaining  int      `json:"resume_remaining"`
	Badges           []string `json:"badges"`
}

type Dashboard struct {
	PassedChapters int     `json:"passed_chapters"`
	TotalChapters  int     `json:"total_chapters"`
	RecentAccuracy float64 `json:"recent_accuracy"`
	HighestScore   float64 `json:"highest_score"`
}

// ListChapters 章节列表（含统计 + 断点续答）
func (s *ExamService) ListChapters(userID string) ([]ChapterWithStats, *Dashboard, error) {
	chapters, err := s.chapterRepo.ListAll()
	if err != nil {
		return nil, nil, err
	}

	result := make([]ChapterWithStats, 0, len(chapters))
	for _, ch := range chapters {
		single, multi, _ := s.questionRepo.CountByChapterNo(ch.ChapterNo)
		exam, _ := s.examRepo.GetByChapterID(ch.ID)
		bestScore := float64(0)
		hasResume := false
		resumeQ, resumeR := 0, 0
		examID := ""
		if exam != nil {
			examID = exam.ID
			bestScore, _ = s.recordRepo.GetBestScore(userID, exam.ID)
			if active, _ := s.recordRepo.GetActiveByUserExam(userID, exam.ID); active != nil {
				hasResume = true
				resumeQ = active.CurrentQuestionNo
				resumeR = active.RemainingSeconds
			}
		}

		badges := buildBadges(ch.ChapterNo, bestScore, single+multi)
		result = append(result, ChapterWithStats{
			ID: ch.ID, ChapterNo: ch.ChapterNo, Title: ch.Title, Category: ch.Category,
			SortOrder: ch.SortOrder, DurationMinutes: ch.DurationMinutes,
			TotalScore: ch.TotalScore, PassScore: ch.PassScore,
			ExamID: examID, SingleCount: single, MultiCount: multi, TotalCount: single + multi,
			BestScore: bestScore, HasResume: hasResume,
			ResumeQuestionNo: resumeQ, ResumeRemaining: resumeR, Badges: badges,
		})
	}

	dashboard := &Dashboard{TotalChapters: len(chapters)}
	passed, _ := s.recordRepo.CountPassedChapters(userID)
	dashboard.PassedChapters = int(passed)
	dashboard.RecentAccuracy, _ = s.recordRepo.GetRecentAccuracy(userID)
	dashboard.HighestScore, _ = s.recordRepo.GetHighestScore(userID)

	return result, dashboard, nil
}

func buildBadges(chapterNo int, bestScore float64, totalQ int) []string {
	badges := []string{}
	if bestScore >= 60 {
		badges = append(badges, fmt.Sprintf("pass:%.1f", bestScore))
	}
	if totalQ > 60 {
		badges = append(badges, "hot")
	}
	if chapterNo == 13 {
		badges = append(badges, "new")
	}
	return badges
}

// GetExamDetail 考试详情（始终返回答案/解析，供前端单题确认后判对错；mode 仅影响前端交互文案）
func (s *ExamService) GetExamDetail(examID string, mode int) (*ExamDetail, error) {
	_ = mode
	exam, err := s.examRepo.GetByID(examID)
	if err != nil {
		return nil, fmt.Errorf("考试不存在")
	}
	questions, err := s.questionRepo.GetByIDs(exam.QuestionIDs)
	if err != nil {
		return nil, fmt.Errorf("获取题目失败")
	}

	// 按 exam.QuestionIDs 顺序重排
	qMap := make(map[string]*model.Question)
	for i := range questions {
		qMap[questions[i].ID] = &questions[i]
	}
	qList := make([]QuestionBrief, 0, len(exam.QuestionIDs))
	for _, qid := range exam.QuestionIDs {
		q, ok := qMap[qid]
		if !ok {
			continue
		}
		brief := QuestionBrief{
			ID: q.ID, QuestionNo: q.QuestionNo, Type: q.Type, Stem: q.Stem,
			Options: q.Options, YearTag: q.YearTag, IsModified: q.IsModified,
			Answer: q.Answer, Analysis: q.Analysis, LawCitation: q.LawCitation,
		}
		qList = append(qList, brief)
	}

	return &ExamDetail{Exam: *exam, Questions: qList}, nil
}

// StartExam 开始考试（有进行中记录则复用）
func (s *ExamService) StartExam(userID, examID string) (*model.ExamRecord, error) {
	exam, err := s.examRepo.GetByID(examID)
	if err != nil {
		return nil, fmt.Errorf("考试不存在")
	}
	if existing, _ := s.recordRepo.GetActiveByUserExam(userID, examID); existing != nil {
		return existing, nil
	}
	record := &model.ExamRecord{
		UserID: userID, ExamID: examID, TotalScore: exam.TotalScore,
		RemainingSeconds: exam.DurationMinutes * 60, Status: 0,
		MarkedQuestions: []string{},
	}
	if err := s.recordRepo.Create(record); err != nil {
		return nil, fmt.Errorf("创建考试记录失败")
	}
	return record, nil
}

// ResumeExam 断点续答
func (s *ExamService) ResumeExam(userID, examID string) (*model.ExamRecord, error) {
	record, err := s.recordRepo.GetActiveByUserExam(userID, examID)
	if err != nil {
		return nil, fmt.Errorf("无进行中的考试")
	}
	return record, nil
}

// SaveProgress 保存答题进度
func (s *ExamService) SaveProgress(userID, recordID string, currentQuestionNo, remainingSeconds int, answers []AnswerSubmission) error {
	record, err := s.recordRepo.GetByID(recordID)
	if err != nil {
		return fmt.Errorf("考试记录不存在")
	}
	if record.UserID != userID || record.Status != 0 {
		return fmt.Errorf("无权操作或考试已结束")
	}
	record.CurrentQuestionNo = currentQuestionNo
	record.RemainingSeconds = remainingSeconds
	marked := make([]string, 0)
	for _, a := range answers {
		if a.IsMarked {
			marked = append(marked, a.QuestionID)
		}
	}
	record.MarkedQuestions = marked
	return s.recordRepo.Update(record)
}

// GetRecordDetail 结果详情（answers 已内嵌题目信息需在 handler 层补全）
func (s *ExamService) GetRecordDetail(recordID string) (*model.ExamRecord, error) {
	record, err := s.recordRepo.GetByID(recordID)
	if err != nil {
		return nil, fmt.Errorf("记录不存在")
	}
	return record, nil
}

// ListRecords 历史记录
func (s *ExamService) ListRecords(userID string, limit int) ([]model.ExamRecord, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.recordRepo.ListByUser(userID, limit)
}

type ExamDetail struct {
	model.Exam
	Questions []QuestionBrief `json:"questions"`
}

type QuestionBrief struct {
	ID          string            `json:"id"`
	QuestionNo  int               `json:"question_no"`
	Type        int               `json:"type"`
	Stem        string            `json:"stem"`
	Options     map[string]string `json:"options"`
	Answer      string            `json:"answer,omitempty"`
	Analysis    string            `json:"analysis,omitempty"`
	LawCitation string            `json:"law_citation,omitempty"`
	YearTag     string            `json:"year_tag"`
	IsModified  int               `json:"is_modified"`
}
