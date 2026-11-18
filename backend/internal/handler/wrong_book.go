package handler

import (
	"accounting-exam/internal/service"
	"accounting-exam/pkg/response"

	"github.com/gin-gonic/gin"
)

type WrongBookHandler struct {
	wrongBookService *service.WrongBookService
}

func NewWrongBookHandler(wrongBookService *service.WrongBookService) *WrongBookHandler {
	return &WrongBookHandler{wrongBookService: wrongBookService}
}

func (h *WrongBookHandler) GetWrongBook(c *gin.Context) {
	userID := c.GetString("user_id")
	summary, err := h.wrongBookService.GetSummary(userID)
	if err != nil {
		response.Error(c, 1008, err.Error())
		return
	}
	response.Success(c, summary)
}

func (h *WrongBookHandler) MarkCorrect(c *gin.Context) {
	userID := c.GetString("user_id")
	var req struct {
		QuestionID string `json:"question_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 4002, "请求参数格式错误")
		return
	}
	if err := h.wrongBookService.MarkCorrect(userID, req.QuestionID); err != nil {
		response.Error(c, 1009, err.Error())
		return
	}
	response.Success(c, gin.H{"marked": true})
}
