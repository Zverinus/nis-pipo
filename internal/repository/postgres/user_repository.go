package postgres

import (
	"context"
	"database/sql"
	"errors"

	"nis-pipo/internal/user"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (repo *UserRepo) Create(ctx context.Context, username, email, passwordHash, firstName, lastName string) (user.User, error) {
	const q = `INSERT INTO users (username, email, password_hash, first_name, last_name)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id::text, username, email, password_hash, COALESCE(first_name,''), COALESCE(last_name,''), created_at, updated_at`
	var u user.User
	err := repo.db.QueryRowContext(ctx, q, username, email, passwordHash, firstName, lastName).
		Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.FirstName, &u.LastName, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}

func (repo *UserRepo) GetByID(ctx context.Context, id string) (user.User, error) {
	const q = `SELECT id::text, username, email, password_hash, COALESCE(first_name,''), COALESCE(last_name,''), created_at, updated_at
		FROM users WHERE id = $1::uuid`
	var u user.User
	err := repo.db.QueryRowContext(ctx, q, id).
		Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.FirstName, &u.LastName, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return user.User{}, user.ErrNotFound
	}
	return u, err
}

func (repo *UserRepo) GetByEmail(ctx context.Context, email string) (user.User, error) {
	const q = `SELECT id::text, username, email, password_hash, COALESCE(first_name,''), COALESCE(last_name,''), created_at, updated_at
		FROM users WHERE email = $1`
	var u user.User
	err := repo.db.QueryRowContext(ctx, q, email).
		Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.FirstName, &u.LastName, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return user.User{}, user.ErrNotFound
	}
	return u, err
}
