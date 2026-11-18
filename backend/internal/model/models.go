package model

import "time"

// User 用户集合 users
type User struct {
	ID        string    `bson:"_id" json:"id"`
	OpenID    string    `bson:"openid" json:"openid"`
	UnionID   string    `bson:"union_id" json:"union_id"`
	Nickname  string    `bson:"nickname" json:"nickname"`
	AvatarURL string    `bson:"avatar_url" json:"avatar_url"`
	IsGuest   int       `bson:"is_guest" json:"is_guest"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

// Chapter 章节集合 chapters
type Chapter struct {
	ID              string    `bson:"_id" json:"id"`            // ch07
	ChapterNo       int       `bson:"chapter_no" json:"chapter_no"`
	Title           string    `bson:"title" json:"title"`
	Category        string    `bson:"category" json:"category"`
	SortOrder       int       `bson:"sort_order" json:"sort_order"`
	DurationMinutes int       `bson:"duration_minutes" json:"duration_minutes"`
	TotalScore      int       `bson:"total_score" json:"total_score"`
	PassScore       int       `bson:"pass_score" json:"pass_score"`
	SingleCount     int       `bson:"single_count" json:"single_count"`
	MultiCount      int       `bson:"multi_count" json:"multi_count"`
	TotalCount      int       `bson:"total_count" json:"total_count"`
	CreatedAt       time.Time `bson:"created_at" json:"created_at"`
}

// Question 题目集合 questions（options 内嵌为对象）
type Question struct {
	ID          string            `bson:"_id" json:"id"`        // q00001
	ChapterNo   int               `bson:"chapter_no" json:"chapter_no"`
	QuestionNo  int               `bson:"question_no" json:"question_no"`
	Type        int               `bson:"type" json:"type"`     // 1=单选 2=多选
	Stem        string            `bson:"stem" json:"stem"`
	Options     map[string]string `bson:"options" json:"options"`
	Answer      string            `bson:"answer" json:"answer"`
	Analysis    string            `bson:"analysis" json:"analysis"`
	LawCitation string            `bson:"law_citation" json:"law_citation"`
	YearTag     string            `bson:"year_tag" json:"year_tag"`
	IsModified  int               `bson:"is_modified" json:"is_modified"`
}

// Exam 考试集合 exams（question_ids 内嵌为数组）
type Exam struct {
	ID              string   `bson:"_id" json:"id"`          // exam07
	ChapterID       string   `bson:"chapter_id" json:"chapter_id"`
	ChapterNo       int      `bson:"chapter_no" json:"chapter_no"`
	Title           string   `bson:"title" json:"title"`
	Description     string   `bson:"description" json:"description"`
	DurationMinutes int      `bson:"duration_minutes" json:"duration_minutes"`
	TotalScore      int      `bson:"total_score" json:"total_score"`
	PassScore       int      `bson:"pass_score" json:"pass_score"`
	QuestionIDs     []string `bson:"question_ids" json:"question_ids"`
	QuestionCount   int      `bson:"question_count" json:"question_count"`
	Mode            int      `bson:"mode" json:"mode"`       // 1=模考 2=背题
	Status          int      `bson:"status" json:"status"`
}

// AnswerDetail 答题明细（内嵌子文档）
type AnswerDetail struct {
	QuestionID string  `bson:"question_id" json:"question_id"`
	UserAnswer string  `bson:"user_answer" json:"user_answer"`
	IsCorrect  int     `bson:"is_correct" json:"is_correct"` // 0=错/未答 1=全对 2=少选半对
	IsAnswered int     `bson:"is_answered" json:"is_answered"`
	IsMarked   int     `bson:"is_marked" json:"is_marked"`
	Score      float64 `bson:"score" json:"score"`
}

// ExamRecord 考试记录集合 exam_records（answers 内嵌数组）
type ExamRecord struct {
	ID                string         `bson:"_id" json:"id"`
	UserID            string         `bson:"user_id" json:"user_id"`
	ExamID            string         `bson:"exam_id" json:"exam_id"`
	Score             float64        `bson:"score" json:"score"`
	TotalScore        int            `bson:"total_score" json:"total_score"`
	CorrectCount      int            `bson:"correct_count" json:"correct_count"`
	PartialCount      int            `bson:"partial_count" json:"partial_count"`
	WrongCount        int            `bson:"wrong_count" json:"wrong_count"`
	UnansweredCount   int            `bson:"unanswered_count" json:"unanswered_count"`
	DurationSeconds   int            `bson:"duration_seconds" json:"duration_seconds"`
	RemainingSeconds  int            `bson:"remaining_seconds" json:"remaining_seconds"`
	CurrentQuestionNo int            `bson:"current_question_no" json:"current_question_no"`
	MarkedQuestions   []string       `bson:"marked_questions" json:"marked_questions"`
	Status            int            `bson:"status" json:"status"` // 0=进行中 1=已完成 2=已放弃
	StartedAt         time.Time      `bson:"started_at" json:"started_at"`
	SubmittedAt       *time.Time     `bson:"submitted_at" json:"submitted_at"`
	Answers           []AnswerDetail `bson:"answers" json:"answers,omitempty"`
}

// WrongQuestion 错题本集合 wrong_questions
type WrongQuestion struct {
	ID            string     `bson:"_id" json:"id"`
	UserID        string     `bson:"user_id" json:"user_id"`
	QuestionID    string     `bson:"question_id" json:"question_id"`
	ChapterNo     int        `bson:"chapter_no" json:"chapter_no"`
	ErrorCount    int        `bson:"error_count" json:"error_count"`
	CorrectStreak int        `bson:"correct_streak" json:"correct_streak"`
	LastWrongAt   time.Time  `bson:"last_wrong_at" json:"last_wrong_at"`
	LastReviewAt  *time.Time `bson:"last_review_at" json:"last_review_at"`
	Status        int        `bson:"status" json:"status"` // 1=待攻克 2=已斩题
}
