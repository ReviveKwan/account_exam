package router

import (
	"accounting-exam/internal/config"
	"accounting-exam/internal/handler"
	"accounting-exam/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func Setup(cfg *config.Config, userHandler *handler.UserHandler, examHandler *handler.ExamHandler, scoreHandler *handler.ScoreHandler, wrongBookHandler *handler.WrongBookHandler) *gin.Engine {
	gin.SetMode(cfg.Server.Mode)
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.CORSMiddleware())

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := r.Group("/api/v1")

	// 无需鉴权
	api.POST("/auth/login", userHandler.Login)
	api.POST("/auth/guest", userHandler.GuestLogin)

	// 需要鉴权
	auth := api.Group("")
	auth.Use(middleware.JWTAuth(cfg.JWT.Secret))
	{
		auth.GET("/chapters", examHandler.ListChapters)
		auth.GET("/exams", examHandler.ListChapters) // 复用，实际可分离
		auth.GET("/exams/:id", examHandler.GetExamDetail)
		auth.POST("/exams/:id/start", examHandler.StartExam)
		auth.GET("/exams/:id/resume", examHandler.ResumeExam)
		auth.PUT("/exams/:id/progress", examHandler.SaveProgress)
		auth.POST("/exams/:id/submit", scoreHandler.SubmitExam)
		auth.GET("/records/:id", examHandler.GetRecordDetail)
		auth.GET("/records", examHandler.ListRecords)

		auth.GET("/wrong-questions", wrongBookHandler.GetWrongBook)
		auth.POST("/wrong-questions/practice", wrongBookHandler.MarkCorrect)
	}

	return r
}
