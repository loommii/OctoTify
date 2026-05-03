package repository

import (
	"context"

	"gorm.io/gorm"

	"octotify/internal/model"
	"octotify/internal/query"
)

type RefreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) Create(ctx context.Context, token *model.RefreshToken) error {
	q := query.Use(r.db)
	return q.WithContext(ctx).RefreshToken.Create(token)
}

func (r *RefreshTokenRepository) FindByJTI(ctx context.Context, jti string) (*model.RefreshToken, error) {
	q := query.Use(r.db)
	return q.WithContext(ctx).RefreshToken.Where(q.RefreshToken.JTI.Eq(jti)).First()
}

func (r *RefreshTokenRepository) Revoke(ctx context.Context, jti string) error {
	q := query.Use(r.db)
	_, err := q.WithContext(ctx).RefreshToken.
		Where(q.RefreshToken.JTI.Eq(jti)).
		Update(q.RefreshToken.Revoked, true)
	return err
}
