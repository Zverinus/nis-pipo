package meeting

import "context"

type Repository interface {
	Create(ctx context.Context, m Meeting) (Meeting, error)
	GetByID(ctx context.Context, id string) (Meeting, error)
	Update(ctx context.Context, id, title, description string) (Meeting, error)
	Delete(ctx context.Context, id string) error
}
