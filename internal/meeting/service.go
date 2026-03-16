package meeting

import (
	"context"
	"errors"
	"time"
)

type SlotsRepo interface {
	SetSlots(ctx context.Context, meetingID, participantID string, slotIndexes []int) error
	GetDetailsByMeeting(ctx context.Context, meetingID string) ([]SlotResult, error)
}

var (
	ErrInvalidDates     = errors.New("date_end must be >= date_start")
	ErrInvalidSlotMin   = errors.New("slot_minutes must be 15, 30 or 60")
	ErrNotFound         = errors.New("meeting not found")
	ErrMeetingFinalized = errors.New("meeting is finalized")
	ErrForbidden        = errors.New("forbidden")
	ErrInvalidSlotIndex = errors.New("final_slot_index out of range")
)

type Service struct {
	repo      Repository
	slotsRepo SlotsRepo
}

func NewService(repo Repository, slotsRepo SlotsRepo) *Service {
	return &Service{repo: repo, slotsRepo: slotsRepo}
}

func (s *Service) Create(ctx context.Context, ownerID, title, description string, dateStart, dateEnd time.Time, slotMinutes int) (Meeting, error) {
	if dateEnd.Before(dateStart) {
		return Meeting{}, ErrInvalidDates
	}
	if slotMinutes != 15 && slotMinutes != 30 && slotMinutes != 60 {
		return Meeting{}, ErrInvalidSlotMin
	}
	m := Meeting{
		OwnerID:     ownerID,
		Title:       title,
		Description: description,
		DateStart:   dateStart,
		DateEnd:     dateEnd,
		SlotMinutes: slotMinutes,
		Status:      "active",
	}
	return s.repo.Create(ctx, m)
}

func (s *Service) GetByID(ctx context.Context, id string) (Meeting, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) ListByOwner(ctx context.Context, ownerID string) ([]Meeting, error) {
	return s.repo.ListByOwner(ctx, ownerID)
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

func (s *Service) GetResults(ctx context.Context, meetingID, ownerID string) ([]SlotResult, error) {
	m, err := s.repo.GetByID(ctx, meetingID)
	if err != nil {
		return nil, ErrNotFound
	}
	if m.OwnerID != ownerID {
		return nil, ErrForbidden
	}
	return s.slotsRepo.GetDetailsByMeeting(ctx, meetingID)
}

func SlotCount(m Meeting) int {
	start := time.Date(m.DateStart.Year(), m.DateStart.Month(), m.DateStart.Day(), 0, 0, 0, 0, time.UTC)
	end := time.Date(m.DateEnd.Year(), m.DateEnd.Month(), m.DateEnd.Day(), 0, 0, 0, 0, time.UTC)
	days := int(end.Sub(start).Hours()/24) + 1
	return days * (24 * 60 / m.SlotMinutes)
}

func (s *Service) Finalize(ctx context.Context, meetingID, ownerID string, finalSlotIndex int) error {
	m, err := s.repo.GetByID(ctx, meetingID)
	if err != nil {
		return ErrNotFound
	}
	if m.OwnerID != ownerID {
		return ErrForbidden
	}
	n := SlotCount(m)
	if finalSlotIndex < 0 || finalSlotIndex >= n {
		return ErrInvalidSlotIndex
	}
	return s.repo.Finalize(ctx, meetingID, finalSlotIndex)
}
