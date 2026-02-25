package meeting

import "context"

type Repository interface {
	Create(ctx context.Context, m Meeting) (Meeting, error)
	GetByID(ctx context.Context, id string) (Meeting, error)
	ListByOwner(ctx context.Context, ownerID string) ([]Meeting, error)
	Update(ctx context.Context, id, title, description string) (Meeting, error)
	Finalize(ctx context.Context, id string, finalSlotIndex int) error
	Delete(ctx context.Context, id string) error
}
