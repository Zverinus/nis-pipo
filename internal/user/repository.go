package user

import "context"

type Repository interface {
	Create(ctx context.Context, username, email, passwordHash, firstName, lastName string) (User, error)
	GetByID(ctx context.Context, id string) (User, error)
	List(ctx context.Context, limit, offset int) ([]User, error)
}