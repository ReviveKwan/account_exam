package handler

import (
	"accounting-exam/internal/model"
	"accounting-exam/internal/repository"
	"accounting-exam/internal/service"
	"accounting-exam/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ExamHandler struct {
	examService  *service.ExamService
	questionRepo *repository.QuestionRepo
}

func NewExamHandler(examService *service.ExamService, questionRepo *repository.QuestionRepo) *ExamHandler {
	return &ExamHandler{examService: examService, questionRepo: questionRepo}
}

func (h *ExamHandler) ListChapters(c *gin.Context) {
	userID := c.GetString("user_id")
	chapters, dashboard, err := h.examService.ListChapters(userID)
	if err != nil {
		response.Error(c, 1001, "获取章节列表失败")
		return
	}
	response.Success(c, gin.H{"chapters": chapters, "dashboard": dashboard})
}

func (h *ExamHandler) GetExamDetail(c *gin.Context) {
	id := c.Param("id")
	mode, _ := strconv.Atoi(c.DefaultQuery("mode", "1"))
	detail, err := h.examService.GetExamDetail(id, mode)
	if err != nil {
		response.Error(c, 1003, err.Error())
		return
	}
	response.Success(c, gin.H{"exam": detail})
}

func (h *ExamHandler) StartExam(c *gin.Context) {
	userID := c.GetString("user_id")
	record, err := h.examService.StartExam(userID, c.Param("id"))
	if err != nil {
		response.Error(c, 1004, err.Error())
		return
	}
	response.Success(c, gin.H{"record": record})
}

func (h *ExamHandler) ResumeExam(c *gin.Context) {
	userID := c.GetString("user_id")
	record, err := h.examService.ResumeExam(userID, c.Param("id"))
	if err != nil {
		response.Error(c, 1004, err.Error())
		return
	}
	response.Success(c, gin.H{"record": record})
}

func (h *ExamHandler) SaveProgress(c *gin.Context) {
	userID := c.GetString("user_id")
	var req struct {
		RecordID          string                     `json:"record_id"`
		CurrentQuestionNo int                        `json:"current_question_no"`
		RemainingSeconds  int                        `json:"remaining_seconds"`
		Answers           []service.AnswerSubmission `json:"answers"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 4002, "请求参数格式错误")
		return
	}
	err := h.examService.SaveProgress(userID, req.RecordID, req.CurrentQuestionNo, req.RemainingSeconds, req.Answers)
	if err != nil {
		response.Error(c, 1005, err.Error())
		return
	}
	response.Success(c, gin.H{"saved": true})
}

// GetRecordDetail 获取考试结果（补全题目题干/解析/法条）
func (h *ExamHandler) GetRecordDetail(c *gin.Context) {
	record, err := h.examService.GetRecordDetail(c.Param("id"))
	if err != nil {
		response.Error(c, 1007, err.Error())
		return
	}

	// 批量补全题目内容（成绩页解析用）
	qIDs := make([]string, 0, len(record.Answers))
	for _, a := range record.Answers {
		qIDs = append(qIDs, a.QuestionID)
	}
	qMap := make(map[string]*model.Question)
	if len(qIDs) > 0 {
		if qs, err := h.questionRepo.GetByIDs(qIDs); err == nil {
			for i := range qs {
				qMap[qs[i].ID] = &qs[i]
			}
		}
	}
	type answerVO struct {
		model.AnswerDetail
		Type          int               `json:"type"`
		Stem          string            `json:"stem"`
		Options       map[string]string `json:"options"`
		CorrectAnswer string            `json:"correct_answer"`
		Analysis      string            `json:"analysis"`
		LawCitation   string            `json:"law_citation"`
	}
	voList := make([]answerVO, 0, len(record.Answers))
	for _, a := range record.Answers {
		vo := answerVO{AnswerDetail: a}
		if q := qMap[a.QuestionID]; q != nil {
			vo.Type = q.Type
			vo.Stem = q.Stem
			vo.Options = q.Options
			vo.CorrectAnswer = q.Answer
			vo.Analysis = q.Analysis
			vo.LawCitation = q.LawCitation
		}
		voList = append(voList, vo)
	}

	response.Success(c, gin.H{"record": record, "answers": voList})
}

func (h *ExamHandler) ListRecords(c *gin.Context) {
	userID := c.GetString("user_id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	records, err := h.examService.ListRecords(userID, limit)
	if err != nil {
		response.Error(c, 1007, err.Error())
		return
	}
	response.Success(c, gin.H{"records": records})
}
