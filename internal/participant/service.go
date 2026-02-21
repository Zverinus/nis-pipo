package participant

import (
	"context"
	"errors"
	"time"

	"nis-pipo/internal/meeting"
	"nis-pipo/internal/participantslots"
)

var (
	ErrMeetingNotFound   = errors.New("meeting not found")
	ErrMeetingFinalized  = errors.New("meeting is finalized")
	ErrParticipantNotFound = errors.New("participant not found")
	ErrSlotOutOfRange    = errors.New("slot_index out of range")
)

type Service struct {
	repo       Repository
	meetingRepo meeting.Repository
	slotsRepo  participantslots.Repository
}

func NewService(repo Repository, meetingRepo meeting.Repository, slotsRepo participantslots.Repository) *Service {
	return &Service{repo: repo, meetingRepo: meetingRepo, slotsRepo: slotsRepo}
}

func slotCountFromMeeting(m meeting.Meeting) int {
	start := time.Date(m.DateStart.Year(), m.DateStart.Month(), m.DateStart.Day(), 0, 0, 0, 0, time.UTC)
	end := time.Date(m.DateEnd.Year(), m.DateEnd.Month(), m.DateEnd.Day(), 0, 0, 0, 0, time.UTC)
	days := int(end.Sub(start).Hours()/24) + 1
	slotsPerDay := 24 * 60 / m.SlotMinutes
	return days * slotsPerDay
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

func (s *Service) SetSlots(ctx context.Context, meetingID, participantToken string, slotIndexes []int) error {
	m, err := s.meetingRepo.GetByID(ctx, meetingID)
	if err != nil {
		return ErrMeetingNotFound
	}
	if m.Status == "finalized" {
		return ErrMeetingFinalized
	}
	p, err := s.repo.GetByMeetingAndToken(ctx, meetingID, participantToken)
	if err != nil {
		return ErrParticipantNotFound
	}
	n := slotCountFromMeeting(m)
	for _, idx := range slotIndexes {
		if idx < 0 || idx >= n {
			return ErrSlotOutOfRange
		}
	}
	return s.slotsRepo.SetSlots(ctx, p.ID, slotIndexes)
}
