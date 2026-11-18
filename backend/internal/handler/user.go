package handler

import (
	"accounting-exam/internal/service"
	"accounting-exam/pkg/response"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) Login(c *gin.Context) {
	var req struct {
		Code string `json:"code"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Code == "" {
		response.Error(c, 4002, "请求参数格式错误")
		return
	}
	user, token, err := h.userService.LoginByWeChat(req.Code)
	if err != nil {
		response.Error(c, 5001, err.Error())
		return
	}
	response.Success(c, gin.H{
		"token": token,
		"user": gin.H{
			"id": user.ID, "nickname": user.Nickname,
			"avatar_url": user.AvatarURL, "is_guest": user.IsGuest,
		},
	})
}

func (h *UserHandler) GuestLogin(c *gin.Context) {
	user, token, err := h.userService.LoginAsGuest()
	if err != nil {
		response.Error(c, 5003, err.Error())
		return
	}
	response.Success(c, gin.H{
		"token": token,
		"user": gin.H{
			"id": user.ID, "nickname": user.Nickname,
			"avatar_url": user.AvatarURL, "is_guest": user.IsGuest,
		},
	})
}
