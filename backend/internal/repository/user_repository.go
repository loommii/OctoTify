package repository

import (
	"context"

	"gorm.io/gorm"

	"octotify/internal/model"
	"octotify/internal/query"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	q := query.Use(r.db)
	return q.WithContext(ctx).User.Create(user)
}

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	q := query.Use(r.db)
	return q.WithContext(ctx).User.Where(q.User.Username.Eq(username)).First()
}

func (r *UserRepository) FindByID(ctx context.Context, id int64) (*model.User, error) {
	q := query.Use(r.db)
	return q.WithContext(ctx).User.Where(q.User.ID.Eq(id)).First()
}
