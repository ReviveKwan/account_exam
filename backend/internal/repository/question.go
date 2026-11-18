package repository

import (
	"accounting-exam/internal/model"
	"accounting-exam/pkg/database"
	"context"
	"time"
)

type QuestionRepo struct {
	store *database.CloudBaseStore
}

func NewQuestionRepo(store *database.CloudBaseStore) *QuestionRepo {
	return &QuestionRepo{store: store}
}

func (r *QuestionRepo) GetByChapterNo(chapterNo int) ([]model.Question, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	docs, err := r.store.Query(ctx, "questions", map[string]interface{}{"chapter_no": chapterNo}, database.QueryOpts{
		Limit: 500,
		Sort:  map[string]int{"question_no": 1},
	})
	if err != nil {
		return nil, err
	}
	return decodeQuestions(docs)
}

func (r *QuestionRepo) GetByIDs(ids []string) ([]model.Question, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	docs, err := r.store.Query(ctx, "questions", map[string]interface{}{
		"_id": map[string]interface{}{"$in": ids},
	}, database.QueryOpts{Limit: len(ids) + 10})
	if err != nil {
		return nil, err
	}
	return decodeQuestions(docs)
}

func (r *QuestionRepo) GetByID(id string) (*model.Question, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	doc, err := r.store.GetByID(ctx, "questions", id)
	if err != nil {
		return nil, err
	}
	var q model.Question
	if err := database.DecodeInto(doc, &q); err != nil {
		return nil, err
	}
	return &q, nil
}

func (r *QuestionRepo) CountByChapterNo(chapterNo int) (single int, multi int, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	s, err := r.store.Count(ctx, "questions", map[string]interface{}{"chapter_no": chapterNo, "type": 1})
	if err != nil {
		return 0, 0, err
	}
	m, err := r.store.Count(ctx, "questions", map[string]interface{}{"chapter_no": chapterNo, "type": 2})
	if err != nil {
		return 0, 0, err
	}
	return int(s), int(m), nil
}

func decodeQuestions(docs []map[string]interface{}) ([]model.Question, error) {
	out := make([]model.Question, 0, len(docs))
	for _, d := range docs {
		var q model.Question
		if err := database.DecodeInto(d, &q); err != nil {
			return nil, err
		}
		out = append(out, q)
	}
	return out, nil
}
