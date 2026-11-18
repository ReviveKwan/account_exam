package repository

import (
	"accounting-exam/pkg/database"
	"context"
	"time"
)

type WrongQuestionRepo struct {
	store *database.CloudBaseStore
}

func NewWrongQuestionRepo(store *database.CloudBaseStore) *WrongQuestionRepo {
	return &WrongQuestionRepo{store: store}
}

func (r *WrongQuestionRepo) Upsert(userID, questionID string, chapterNo int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	now := time.Now()
	filter := map[string]interface{}{"user_id": userID, "question_id": questionID}

	existing, err := r.store.FindOne(ctx, "wrong_questions", filter, nil)
	if err != nil && !IsNotFound(err) {
		return err
	}
	if IsNotFound(err) || existing == nil {
		doc := map[string]interface{}{
			"_id":            "wq_" + newUUID(),
			"user_id":        userID,
			"question_id":    questionID,
			"chapter_no":     chapterNo,
			"error_count":    1,
			"correct_streak": 0,
			"last_wrong_at":  now,
			"status":         1,
		}
		return r.store.Insert(ctx, "wrong_questions", doc)
	}

	id, _ := existing["_id"].(string)
	if id == "" {
		if s, ok := existing["id"].(string); ok {
			id = s
		}
	}
	return r.store.UpdateOne(ctx, "wrong_questions", id, map[string]interface{}{
		"$set": map[string]interface{}{
			"chapter_no":     chapterNo,
			"correct_streak": 0,
			"last_wrong_at":  now,
			"status":         1,
		},
		"$inc": map[string]interface{}{"error_count": 1},
	})
}

func (r *WrongQuestionRepo) MarkCorrect(userID, questionID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	now := time.Now()
	doc, err := r.store.FindOne(ctx, "wrong_questions", map[string]interface{}{
		"user_id": userID, "question_id": questionID, "status": 1,
	}, nil)
	if err != nil {
		if IsNotFound(err) {
			return nil
		}
		return err
	}
	id, _ := doc["_id"].(string)
	streak := toInt(doc["correct_streak"]) + 1
	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"correct_streak": streak,
			"last_review_at": now,
		},
	}
	if streak >= 2 {
		update["$set"].(map[string]interface{})["status"] = 2
	}
	return r.store.UpdateOne(ctx, "wrong_questions", id, update)
}

func (r *WrongQuestionRepo) CountByUser(userID string) (total, pending, resolved int64) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	total, _ = r.store.Count(ctx, "wrong_questions", map[string]interface{}{"user_id": userID})
	pending, _ = r.store.Count(ctx, "wrong_questions", map[string]interface{}{"user_id": userID, "status": 1})
	resolved, _ = r.store.Count(ctx, "wrong_questions", map[string]interface{}{"user_id": userID, "status": 2})
	return total, pending, resolved
}

func (r *WrongQuestionRepo) ListChapterStats(userID string) ([]ChapterWrongStat, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	docs, err := r.store.Query(ctx, "wrong_questions", map[string]interface{}{
		"user_id": userID, "status": 1,
	}, database.QueryOpts{Limit: 1000})
	if err != nil {
		return nil, err
	}
	counter := map[int]int64{}
	for _, d := range docs {
		ch := toInt(d["chapter_no"])
		counter[ch]++
	}
	out := make([]ChapterWrongStat, 0, len(counter))
	for ch, n := range counter {
		out = append(out, ChapterWrongStat{ChapterNo: ch, WrongCount: n})
	}
	return out, nil
}

type ChapterWrongStat struct {
	ChapterNo  int   `json:"chapter_no"`
	WrongCount int64 `json:"wrong_count"`
}

func toInt(v interface{}) int {
	switch t := v.(type) {
	case int:
		return t
	case int32:
		return int(t)
	case int64:
		return int(t)
	case float64:
		return int(t)
	default:
		return 0
	}
}
