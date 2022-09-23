package repo

import (
	"context"
	"fmt"
	"user-balance-api/internal/entity"
	"user-balance-api/pkg/postgres"
)

type AuthRepo struct {
	*postgres.Postgres
}

func NewAuthRepo(pg *postgres.Postgres) *AuthRepo {
	return &AuthRepo{pg}
}

func (r *AuthRepo) CreateUser(ctx context.Context, user entity.User) (int, error) {
	sql, args, err := r.Builder.
		Insert("users").
		Columns("username", "password_hash").
		Values(user.Username, user.Password).
		Suffix("RETURNING id").
		ToSql()

	if err != nil {
		return 0, fmt.Errorf("repo - AuthRepo - CreateUser - r.Builder: %w", err)
	}

	var id int
	err = r.Pool.QueryRow(ctx, sql, args...).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("repo - AuthRepo - CreateUser - r.Pool.QueryRow: %w", err)
	}

	return id, nil
}

func (r *AuthRepo) GetUser(ctx context.Context, username, passwordHash string) (entity.User, error) {
	sql, args, err := r.Builder.
		Select("id", "username", "password_hash").
		From("users").
		Where("username = ? AND password_hash = ?", username, passwordHash).
		ToSql()

	if err != nil {
		return entity.User{}, fmt.Errorf("repo - AuthRepo - GetUser - r.Builder: %w", err)
	}

	var user entity.User
	err = r.Pool.QueryRow(ctx, sql, args...).Scan(&user.Id, &user.Username, &user.Password)
	if err != nil {
		return entity.User{}, fmt.Errorf("repo - AuthRepo - GetUser - r.Pool.QueryRow: %w", err)
	}

	return user, nil
}
