package repository

import (
	"accounting-exam/internal/model"
	"accounting-exam/pkg/database"
	"context"
	"errors"
	"time"
)

type UserRepo struct {
	store *database.CloudBaseStore
}

func NewUserRepo(store *database.CloudBaseStore) *UserRepo {
	return &UserRepo{store: store}
}

func (r *UserRepo) FindByOpenID(openid string) (*model.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	doc, err := r.store.FindOne(ctx, "users", map[string]interface{}{"openid": openid}, nil)
	if err != nil {
		return nil, err
	}
	var user model.User
	if err := database.DecodeInto(doc, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) Create(user *model.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	doc, err := database.ToDoc(user)
	if err != nil {
		return err
	}
	return r.store.Insert(ctx, "users", doc)
}

func (r *UserRepo) GetByID(id string) (*model.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	doc, err := r.store.GetByID(ctx, "users", id)
	if err != nil {
		return nil, err
	}
	var user model.User
	if err := database.DecodeInto(doc, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func IsNotFound(err error) bool {
	return errors.Is(err, database.ErrNotFound)
}
