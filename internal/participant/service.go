package participant

import (
	"context"
	"errors"

	"nis-pipo/internal/meeting"
)

var (
	ErrMeetingNotFound     = errors.New("meeting not found")
	ErrMeetingFinalized    = errors.New("meeting is finalized")
	ErrParticipantNotFound = errors.New("participant not found")
	ErrSlotOutOfRange      = errors.New("slot_index out of range")
)

type Service struct {
	repo        Repository
	meetingRepo meeting.Repository
	slotsRepo   meeting.SlotsRepo
}

func NewService(repo Repository, meetingRepo meeting.Repository, slotsRepo meeting.SlotsRepo) *Service {
	return &Service{repo: repo, meetingRepo: meetingRepo, slotsRepo: slotsRepo}
}

func (s *Service) Create(ctx context.Context, meetingID, displayName string) (Participant, error) {
	m, err := s.meetingRepo.GetByID(ctx, meetingID)
	if err != nil {
		return Participant{}, ErrMeetingNotFound
	}
	if m.Status != "active" {
		return Participant{}, ErrMeetingFinalized
	}
	return s.repo.Create(ctx, meetingID, displayName)
}

func (s *Service) SetSlots(ctx context.Context, meetingID, participantID string, slotIndexes []int) error {
	m, err := s.meetingRepo.GetByID(ctx, meetingID)
	if err != nil {
		return ErrMeetingNotFound
	}
	if m.Status == "finalized" {
		return ErrMeetingFinalized
	}
	p, err := s.repo.GetByMeetingAndID(ctx, meetingID, participantID)
	if err != nil {
		return ErrParticipantNotFound
	}
	n := meeting.SlotCount(m)
	seen := make(map[int]bool)
	var unique []int
	for _, idx := range slotIndexes {
		if idx < 0 || idx >= n {
			return ErrSlotOutOfRange
		}
		if !seen[idx] {
			seen[idx] = true
			unique = append(unique, idx)
		}
	}
	return s.slotsRepo.SetSlots(ctx, p.ID, unique)
}
