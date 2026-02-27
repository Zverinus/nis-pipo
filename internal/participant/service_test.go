package participant

import (
	"context"
	"errors"
	"testing"
	"time"

	"nis-pipo/internal/meeting"
	"nis-pipo/internal/participantslots"
)

func mustDate(s string) time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return t
}

type mockParticipantRepo struct {
	create             func(ctx context.Context, meetingID, displayName string) (Participant, error)
	getByID            func(ctx context.Context, id string) (Participant, error)
	getByMeetingAndToken func(ctx context.Context, meetingID, token string) (Participant, error)
}

func (m *mockParticipantRepo) Create(ctx context.Context, meetingID, displayName string) (Participant, error) {
	if m.create != nil {
		return m.create(ctx, meetingID, displayName)
	}
	return Participant{}, nil
}

func (m *mockParticipantRepo) GetByID(ctx context.Context, id string) (Participant, error) {
	if m.getByID != nil {
		return m.getByID(ctx, id)
	}
	return Participant{}, nil
}

func (m *mockParticipantRepo) GetByMeetingAndToken(ctx context.Context, meetingID, token string) (Participant, error) {
	if m.getByMeetingAndToken != nil {
		return m.getByMeetingAndToken(ctx, meetingID, token)
	}
	return Participant{}, errors.New("not found")
}

type mockMeetingRepo struct {
	getByID func(ctx context.Context, id string) (meeting.Meeting, error)
}

func (m *mockMeetingRepo) GetByID(ctx context.Context, id string) (meeting.Meeting, error) {
	if m.getByID != nil {
		return m.getByID(ctx, id)
	}
	return meeting.Meeting{}, errors.New("not found")
}

func (m *mockMeetingRepo) Create(ctx context.Context, _ meeting.Meeting) (meeting.Meeting, error) {
	return meeting.Meeting{}, nil
}

func (m *mockMeetingRepo) Update(ctx context.Context, id, title, description string) (meeting.Meeting, error) {
	return meeting.Meeting{}, nil
}

func (m *mockMeetingRepo) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *mockMeetingRepo) Finalize(ctx context.Context, id string, finalSlotIndex int) error {
	return nil
}

func (m *mockMeetingRepo) ListByOwner(ctx context.Context, ownerID string) ([]meeting.Meeting, error) {
	return nil, nil
}

type mockSlotsRepo struct {
	setSlots func(ctx context.Context, participantID string, slotIndexes []int) error
}

func (m *mockSlotsRepo) SetSlots(ctx context.Context, participantID string, slotIndexes []int) error {
	if m.setSlots != nil {
		return m.setSlots(ctx, participantID, slotIndexes)
	}
	return nil
}

func (m *mockSlotsRepo) GetByParticipant(ctx context.Context, participantID string) ([]int, error) {
	return nil, nil
}

func (m *mockSlotsRepo) GetCountsByMeeting(ctx context.Context, meetingID string) ([]participantslots.SlotCount, error) {
	return nil, nil
}

func (m *mockSlotsRepo) GetDetailsByMeeting(ctx context.Context, meetingID string) ([]participantslots.SlotDetails, error) {
	return nil, nil
}

func (m *mockSlotsRepo) CountByMeetingAndSlot(ctx context.Context, meetingID string, slotIndex int) (int, error) {
	return 0, nil
}

func activeMeeting(id string) meeting.Meeting {
	return meeting.Meeting{
		ID:          id,
		DateStart:   mustDate("2025-03-01"),
		DateEnd:     mustDate("2025-03-01"),
		SlotMinutes: 30,
		Status:      "active",
	}
}

func finalizedMeeting(id string) meeting.Meeting {
	m := activeMeeting(id)
	m.Status = "finalized"
	return m
}

func TestCreate(t *testing.T) {
	ctx := context.Background()

	t.Run("add participant to active meeting", func(t *testing.T) {
		mRepo := &mockMeetingRepo{getByID: func(_ context.Context, id string) (meeting.Meeting, error) { return activeMeeting(id), nil }}
		pRepo := &mockParticipantRepo{
			create: func(_ context.Context, mid, name string) (Participant, error) {
				return Participant{ID: "p1", MeetingID: mid, DisplayName: name, CreatedAt: time.Now()}, nil
			},
		}
		p, err := NewService(pRepo, mRepo, &mockSlotsRepo{}).Create(ctx, "m1", "Alice")
		if err != nil {
			t.Fatal(err)
		}
		if p.DisplayName != "Alice" {
			t.Fail()
		}
	})

	t.Run("cannot add to finalized meeting", func(t *testing.T) {
		mRepo := &mockMeetingRepo{getByID: func(_ context.Context, id string) (meeting.Meeting, error) { return finalizedMeeting(id), nil }}
		_, err := NewService(&mockParticipantRepo{}, mRepo, &mockSlotsRepo{}).Create(ctx, "m1", "Alice")
		if err != ErrMeetingFinalized {
			t.Fatalf("got %v, want ErrMeetingFinalized", err)
		}
	})
}

func TestSetSlots(t *testing.T) {
	ctx := context.Background()

	t.Run("save slots", func(t *testing.T) {
		var savedSlots []int
		mRepo := &mockMeetingRepo{getByID: func(_ context.Context, id string) (meeting.Meeting, error) { return activeMeeting(id), nil }}
		pRepo := &mockParticipantRepo{
			getByMeetingAndToken: func(_ context.Context, _, _ string) (Participant, error) { return Participant{ID: "p1"}, nil },
		}
		sRepo := &mockSlotsRepo{
			setSlots: func(_ context.Context, _ string, idx []int) error { savedSlots = idx; return nil },
		}
		err := NewService(pRepo, mRepo, sRepo).SetSlots(ctx, "m1", "p1", []int{0, 1})
		if err != nil {
			t.Fatal(err)
		}
		if len(savedSlots) != 2 {
			t.Fatalf("expected 2 slots, got %v", savedSlots)
		}
	})

	t.Run("slot_index out of range", func(t *testing.T) {
		mRepo := &mockMeetingRepo{getByID: func(_ context.Context, id string) (meeting.Meeting, error) { return activeMeeting(id), nil }}
		pRepo := &mockParticipantRepo{
			getByMeetingAndToken: func(_ context.Context, _, _ string) (Participant, error) { return Participant{ID: "p1"}, nil },
		}
		err := NewService(pRepo, mRepo, &mockSlotsRepo{}).SetSlots(ctx, "m1", "p1", []int{0, 999})
		if err != ErrSlotOutOfRange {
			t.Fatalf("got %v, want ErrSlotOutOfRange", err)
		}
	})

	t.Run("cannot change slots in finalized meeting", func(t *testing.T) {
		mRepo := &mockMeetingRepo{getByID: func(_ context.Context, id string) (meeting.Meeting, error) { return finalizedMeeting(id), nil }}
		err := NewService(&mockParticipantRepo{}, mRepo, &mockSlotsRepo{}).SetSlots(ctx, "m1", "p1", []int{0})
		if err != ErrMeetingFinalized {
			t.Fatalf("got %v, want ErrMeetingFinalized", err)
		}
	})
}
