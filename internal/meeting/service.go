package meeting

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalidDates   = errors.New("date_end must be >= date_start")
	ErrInvalidSlotMin = errors.New("slot_minutes must be 15, 30 or 60")
	ErrNotFound       = errors.New("meeting not found")
	ErrForbidden      = errors.New("forbidden")
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, ownerID, title, description string, dateStart, dateEnd time.Time, slotMinutes int) (Meeting, error) {
	if dateEnd.Before(dateStart) {
		return Meeting{}, ErrInvalidDates
	}
	if slotMinutes != 15 && slotMinutes != 30 && slotMinutes != 60 {
		return Meeting{}, ErrInvalidSlotMin
	}
	m := Meeting{
		OwnerID:    ownerID,
		Title:      title,
		Description: description,
		DateStart:  dateStart,
		DateEnd:    dateEnd,
		SlotMinutes: slotMinutes,
		Status:     "active",
	}
	return s.repo.Create(ctx, m)
}

func (s *Service) GetByID(ctx context.Context, id string) (Meeting, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id, ownerID, title, description string) (Meeting, error) {
	m, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return Meeting{}, ErrNotFound
	}
	if m.OwnerID != ownerID {
		return Meeting{}, ErrForbidden
	}
	return s.repo.Update(ctx, id, title, description)
}

func (s *Service) Delete(ctx context.Context, id, ownerID string) error {
	m, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return ErrNotFound
	}
	if m.OwnerID != ownerID {
		return ErrForbidden
	}
	return s.repo.Delete(ctx, id)
}
