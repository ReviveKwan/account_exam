package repository

import "github.com/google/uuid"

// newUUID 生成短 UUID（用于 record / wrong_question 的 _id）
func newUUID() string {
	return uuid.New().String()
}
