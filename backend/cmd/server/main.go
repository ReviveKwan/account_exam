package main

import (
	"accounting-exam/internal/config"
	"accounting-exam/internal/handler"
	"accounting-exam/internal/repository"
	"accounting-exam/internal/router"
	"accounting-exam/internal/service"
	"accounting-exam/pkg/database"
	"context"
	"fmt"
	"log"
	"time"
)

func main() {
	cfg := config.Load()

	log.Printf("云开发环境: %s | 实例: %s | 监听端口: %d",
		cfg.CloudBase.EnvID, cfg.CloudBase.Instance, cfg.Server.Port)

	store, err := database.ConnectCloudBase(database.StoreConfig{
		EnvID:    cfg.CloudBase.EnvID,
		APIKey:   cfg.CloudBase.APIKey,
		Instance: cfg.CloudBase.Instance,
		Database: cfg.CloudBase.Database,
	})
	if err != nil {
		log.Fatalf("云开发配置错误: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	store.EnsureCollections(ctx, "users", "exam_records", "wrong_questions", "chapters", "questions", "exams")
	if pingErr := store.Ping(ctx); pingErr != nil {
		log.Printf("⚠️  云开发文档库探活失败（请确认环境变量名是 CLOUDBASE_APIKEY）: %v", pingErr)
		log.Printf("   当前 APIKey 长度=%d env=%s", len(cfg.CloudBase.APIKey), cfg.CloudBase.EnvID)
	} else {
		log.Printf("已连接云开发自带文档库（默认实例）")
		checkQuestions(ctx, store)
	}
	cancel()

	userRepo := repository.NewUserRepo(store)
	chapterRepo := repository.NewChapterRepo(store)
	questionRepo := repository.NewQuestionRepo(store)
	examRepo := repository.NewExamRepo(store)
	recordRepo := repository.NewRecordRepo(store)
	wrongRepo := repository.NewWrongQuestionRepo(store)

	userService := service.NewUserService(userRepo, cfg)
	examService := service.NewExamService(examRepo, recordRepo, questionRepo, chapterRepo)
	scoreService := service.NewScoreService(recordRepo, questionRepo, wrongRepo)
	wrongBookService := service.NewWrongBookService(wrongRepo, questionRepo, chapterRepo)

	userHandler := handler.NewUserHandler(userService)
	examHandler := handler.NewExamHandler(examService, questionRepo)
	scoreHandler := handler.NewScoreHandler(scoreService)
	wrongBookHandler := handler.NewWrongBookHandler(wrongBookService)

	r := router.Setup(cfg, userHandler, examHandler, scoreHandler, wrongBookHandler)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("法考磨笔头后端启动 → http://0.0.0.0%s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}

func checkQuestions(ctx context.Context, store *database.CloudBaseStore) {
	n, err := store.Count(ctx, "questions", map[string]interface{}{})
	if err != nil {
		log.Printf("⚠️  读取 questions 失败: %v（请确认集合已创建且 API Key 有权限）", err)
		return
	}
	if n == 0 {
		log.Printf("⚠️  题库为空，请导入 database/cloudbase/*.jsonl")
	} else {
		log.Printf("题库已就绪: %d 题", n)
	}
}
