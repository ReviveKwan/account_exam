package handler

import (
	"accounting-exam/internal/service"
	"accounting-exam/pkg/response"

	"github.com/gin-gonic/gin"
)

type ScoreHandler struct {
	scoreService *service.ScoreService
}

func NewScoreHandler(scoreService *service.ScoreService) *ScoreHandler {
	return &ScoreHandler{scoreService: scoreService}
}

func (h *ScoreHandler) SubmitExam(c *gin.Context) {
	userID := c.GetString("user_id")
	var req struct {
		RecordID        string                     `json:"record_id"`
		DurationSeconds int                        `json:"duration_seconds"`
		Answers         []service.AnswerSubmission `json:"answers"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 4002, "请求参数格式错误")
		return
	}
	result, err := h.scoreService.SubmitAnswers(userID, req.RecordID, req.DurationSeconds, req.Answers)
	if err != nil {
		response.Error(c, 1006, err.Error())
		return
	}
	response.Success(c, result)
}
