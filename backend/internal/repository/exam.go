package repository

import (
	"accounting-exam/internal/model"
	"accounting-exam/pkg/database"
	"context"
	"time"
)

type ExamRepo struct {
	store *database.CloudBaseStore
}

func NewExamRepo(store *database.CloudBaseStore) *ExamRepo {
	return &ExamRepo{store: store}
}

func (r *ExamRepo) List() ([]model.Exam, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	docs, err := r.store.Query(ctx, "exams", map[string]interface{}{"status": 1}, database.QueryOpts{
		Limit: 100,
		Sort:  map[string]int{"chapter_no": 1},
	})
	if err != nil {
		return nil, err
	}
	out := make([]model.Exam, 0, len(docs))
	for _, d := range docs {
		var exam model.Exam
		if err := database.DecodeInto(d, &exam); err != nil {
			return nil, err
		}
		out = append(out, exam)
	}
	return out, nil
}

func (r *ExamRepo) GetByID(id string) (*model.Exam, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	doc, err := r.store.GetByID(ctx, "exams", id)
	if err != nil {
		return nil, err
	}
	var exam model.Exam
	if err := database.DecodeInto(doc, &exam); err != nil {
		return nil, err
	}
	return &exam, nil
}

func (r *ExamRepo) GetByChapterID(chapterID string) (*model.Exam, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	doc, err := r.store.FindOne(ctx, "exams", map[string]interface{}{
		"chapter_id": chapterID, "status": 1,
	}, nil)
	if err != nil {
		return nil, err
	}
	var exam model.Exam
	if err := database.DecodeInto(doc, &exam); err != nil {
		return nil, err
	}
	return &exam, nil
}

type RecordRepo struct {
	store *database.CloudBaseStore
}

func NewRecordRepo(store *database.CloudBaseStore) *RecordRepo {
	return &RecordRepo{store: store}
}

func (r *RecordRepo) Create(record *model.ExamRecord) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if record.ID == "" {
		record.ID = "rec_" + newUUID()
	}
	record.StartedAt = time.Now()
	doc, err := database.ToDoc(record)
	if err != nil {
		return err
	}
	return r.store.Insert(ctx, "exam_records", doc)
}

func (r *RecordRepo) GetByID(id string) (*model.ExamRecord, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	doc, err := r.store.GetByID(ctx, "exam_records", id)
	if err != nil {
		return nil, err
	}
	var record model.ExamRecord
	if err := database.DecodeInto(doc, &record); err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *RecordRepo) GetActiveByUserExam(userID, examID string) (*model.ExamRecord, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	doc, err := r.store.FindOne(ctx, "exam_records", map[string]interface{}{
		"user_id": userID, "exam_id": examID, "status": 0,
	}, nil)
	if err != nil {
		return nil, err
	}
	var record model.ExamRecord
	if err := database.DecodeInto(doc, &record); err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *RecordRepo) Update(record *model.ExamRecord) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	doc, err := database.ToDoc(record)
	if err != nil {
		return err
	}
	return r.store.Replace(ctx, "exam_records", record.ID, doc)
}

func (r *RecordRepo) UpdateAnswers(id string, answers []model.AnswerDetail, record *model.ExamRecord) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return r.store.UpdateOne(ctx, "exam_records", id, map[string]interface{}{
		"$set": map[string]interface{}{
			"answers":          answers,
			"score":            record.Score,
			"correct_count":    record.CorrectCount,
			"partial_count":    record.PartialCount,
			"wrong_count":      record.WrongCount,
			"unanswered_count": record.UnansweredCount,
			"duration_seconds": record.DurationSeconds,
			"status":           record.Status,
			"submitted_at":     record.SubmittedAt,
		},
	})
}

func (r *RecordRepo) ListByUser(userID string, limit int) ([]model.ExamRecord, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if limit <= 0 {
		limit = 20
	}
	docs, err := r.store.Query(ctx, "exam_records", map[string]interface{}{
		"user_id": userID, "status": 1,
	}, database.QueryOpts{
		Limit: limit,
		Sort:  map[string]int{"submitted_at": -1},
	})
	if err != nil {
		return nil, err
	}
	out := make([]model.ExamRecord, 0, len(docs))
	for _, d := range docs {
		var rec model.ExamRecord
		if err := database.DecodeInto(d, &rec); err != nil {
			return nil, err
		}
		out = append(out, rec)
	}
	return out, nil
}

func (r *RecordRepo) GetBestScore(userID, examID string) (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	doc, err := r.store.FindOne(ctx, "exam_records", map[string]interface{}{
		"user_id": userID, "exam_id": examID, "status": 1,
	}, map[string]int{"score": -1})
	if err != nil {
		if IsNotFound(err) {
			return 0, nil
		}
		return 0, err
	}
	var record model.ExamRecord
	if err := database.DecodeInto(doc, &record); err != nil {
		return 0, err
	}
	return record.Score, nil
}

func (r *RecordRepo) CountPassedChapters(userID string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	docs, err := r.store.Query(ctx, "exam_records", map[string]interface{}{
		"user_id": userID,
		"status":  1,
		"score":   map[string]interface{}{"$gte": 60},
	}, database.QueryOpts{Limit: 500})
	if err != nil {
		return 0, err
	}
	seen := map[string]struct{}{}
	for _, d := range docs {
		if id, ok := d["exam_id"].(string); ok {
			seen[id] = struct{}{}
		}
	}
	return int64(len(seen)), nil
}

func (r *RecordRepo) GetRecentAccuracy(userID string) (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	docs, err := r.store.Query(ctx, "exam_records", map[string]interface{}{
		"user_id": userID, "status": 1,
	}, database.QueryOpts{Limit: 3, Sort: map[string]int{"submitted_at": -1}})
	if err != nil {
		return 0, err
	}
	if len(docs) == 0 {
		return 0, nil
	}
	var sum float64
	for _, d := range docs {
		var rec model.ExamRecord
		if err := database.DecodeInto(d, &rec); err != nil {
			return 0, err
		}
		if rec.TotalScore > 0 {
			sum += rec.Score / float64(rec.TotalScore) * 100
		}
	}
	return sum / float64(len(docs)), nil
}

func (r *RecordRepo) GetHighestScore(userID string) (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	doc, err := r.store.FindOne(ctx, "exam_records", map[string]interface{}{
		"user_id": userID, "status": 1,
	}, map[string]int{"score": -1})
	if err != nil {
		if IsNotFound(err) {
			return 0, nil
		}
		return 0, err
	}
	var record model.ExamRecord
	if err := database.DecodeInto(doc, &record); err != nil {
		return 0, err
	}
	return record.Score, nil
}
