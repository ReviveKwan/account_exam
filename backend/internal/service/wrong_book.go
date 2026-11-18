package service

import (
	"accounting-exam/internal/repository"
	"fmt"
)

type WrongBookService struct {
	wrongRepo    *repository.WrongQuestionRepo
	questionRepo *repository.QuestionRepo
	chapterRepo  *repository.ChapterRepo
}

func NewWrongBookService(wrongRepo *repository.WrongQuestionRepo, questionRepo *repository.QuestionRepo, chapterRepo *repository.ChapterRepo) *WrongBookService {
	return &WrongBookService{wrongRepo: wrongRepo, questionRepo: questionRepo, chapterRepo: chapterRepo}
}

type WrongBookSummary struct {
	TotalWrong  int64                `json:"total_wrong"`
	Resolved    int64                `json:"resolved"`
	Pending     int64                `json:"pending"`
	TodayReview int64                `json:"today_review"`
	Chapters    []ChapterWrongDetail `json:"chapters"`
}

type ChapterWrongDetail struct {
	ChapterNo  int    `json:"chapter_no"`
	Title      string `json:"title"`
	WrongCount int64  `json:"wrong_count"`
	Mastery    int    `json:"mastery"`
}

func (s *WrongBookService) GetSummary(userID string) (*WrongBookSummary, error) {
	total, pending, resolved := s.wrongRepo.CountByUser(userID)
	stats, _ := s.wrongRepo.ListChapterStats(userID)

	chapters := make([]ChapterWrongDetail, 0, len(stats))
	for _, st := range stats {
		title := ""
		if ch, err := s.chapterRepo.GetByID(chIDByNo(st.ChapterNo)); err == nil {
			title = ch.Title
		}
		single, multi, _ := s.questionRepo.CountByChapterNo(st.ChapterNo)
		totalQ := single + multi
		mastery := 100
		if totalQ > 0 {
			mastery = 100 - int(float64(st.WrongCount)/float64(totalQ)*100)
			if mastery < 0 {
				mastery = 0
			}
		}
		chapters = append(chapters, ChapterWrongDetail{
			ChapterNo: st.ChapterNo, Title: title,
			WrongCount: st.WrongCount, Mastery: mastery,
		})
	}

	return &WrongBookSummary{
		TotalWrong: total, Resolved: resolved, Pending: pending,
		TodayReview: pending, Chapters: chapters,
	}, nil
}

func (s *WrongBookService) MarkCorrect(userID, questionID string) error {
	return s.wrongRepo.MarkCorrect(userID, questionID)
}

// chIDByNo 章节号 → 章节 _id（ch07 格式）
func chIDByNo(no int) string {
	return fmt.Sprintf("ch%02d", no)
}

func fmtNo(no int) string {
	return fmt.Sprintf("%02d", no)
}
