package repository

import (
	"accounting-exam/internal/model"
	"accounting-exam/pkg/database"
	"context"
	"time"
)

type ChapterRepo struct {
	store *database.CloudBaseStore
}

func NewChapterRepo(store *database.CloudBaseStore) *ChapterRepo {
	return &ChapterRepo{store: store}
}

func (r *ChapterRepo) ListAll() ([]model.Chapter, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	docs, err := r.store.Query(ctx, "chapters", map[string]interface{}{}, database.QueryOpts{
		Limit: 100,
		Sort:  map[string]int{"sort_order": 1, "chapter_no": 1},
	})
	if err != nil {
		return nil, err
	}
	out := make([]model.Chapter, 0, len(docs))
	for _, d := range docs {
		var ch model.Chapter
		if err := database.DecodeInto(d, &ch); err != nil {
			return nil, err
		}
		out = append(out, ch)
	}
	return out, nil
}

func (r *ChapterRepo) GetByID(id string) (*model.Chapter, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	doc, err := r.store.GetByID(ctx, "chapters", id)
	if err != nil {
		return nil, err
	}
	var ch model.Chapter
	if err := database.DecodeInto(doc, &ch); err != nil {
		return nil, err
	}
	return &ch, nil
}
