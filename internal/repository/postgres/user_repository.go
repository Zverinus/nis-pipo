package postgres

import (
	"context"
	"database/sql"

	"nis-pipo/internal/user"
)

type UserRepo struct{ 
	db *sql.DB 
}

func NewUserRepo(db *sql.DB) *UserRepo { 
	return &UserRepo{db: db} 
}

func (repo *UserRepo) Create(ctx context.Context, username, email, passwordHash, firstName, lastName string) (user.User, error) {
	const query = `
		INSERT INTO users (username, email, password_hash, first_name, last_name)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id::text, username, email, password_hash, first_name, last_name, created_at, updated_at`
	var u user.User
	err := repo.db.QueryRowContext(ctx, query,
		username, email, passwordHash, firstName, lastName,
	).Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.FirstName, &u.LastName, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}

func (repo *UserRepo) GetByID(ctx context.Context, id string) (user.User, error) {
	const query = `
		SELECT id::text, username, email, password_hash, first_name, last_name, created_at, updated_at
		FROM users
		WHERE id = $1`
	var u user.User
	err := repo.db.QueryRowContext(ctx, query, id).Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.FirstName, &u.LastName, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}
	
func (repo *UserRepo) List(ctx context.Context, limit, offset int) ([]user.User, error) {
	const query = `
		SELECT id::text, username, email, password_hash, first_name, last_name, created_at, updated_at
		FROM users
		LIMIT $1 OFFSET $2`
	rows, err := repo.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []user.User
	for rows.Next() {
		var u user.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.FirstName, &u.LastName, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}