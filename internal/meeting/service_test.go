package meeting

import (
	"context"
	"errors"
	"testing"
	"time"

)

func mustDate(s string) time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return t
}

type mockMeetingRepo struct {
	getByID    func(ctx context.Context, id string) (Meeting, error)
	create     func(ctx context.Context, m Meeting) (Meeting, error)
	update     func(ctx context.Context, id, title, description string) (Meeting, error)
	delete     func(ctx context.Context, id string) error
	finalize   func(ctx context.Context, id string, finalSlotIndex int) error
	listByOwner func(ctx context.Context, ownerID string) ([]Meeting, error)
}

func (m *mockMeetingRepo) GetByID(ctx context.Context, id string) (Meeting, error) {
	if m.getByID != nil {
		return m.getByID(ctx, id)
	}
	return Meeting{}, errors.New("not found")
}

func (m *mockMeetingRepo) Create(ctx context.Context, meeting Meeting) (Meeting, error) {
	if m.create != nil {
		return m.create(ctx, meeting)
	}
	return Meeting{}, nil
}

func (m *mockMeetingRepo) Update(ctx context.Context, id, title, description string) (Meeting, error) {
	if m.update != nil {
		return m.update(ctx, id, title, description)
	}
	return Meeting{}, nil
}

func (m *mockMeetingRepo) Delete(ctx context.Context, id string) error {
	if m.delete != nil {
		return m.delete(ctx, id)
	}
	return nil
}

func (m *mockMeetingRepo) Finalize(ctx context.Context, id string, finalSlotIndex int) error {
	if m.finalize != nil {
		return m.finalize(ctx, id, finalSlotIndex)
	}
	return nil
}

func (m *mockMeetingRepo) ListByOwner(ctx context.Context, ownerID string) ([]Meeting, error) {
	if m.listByOwner != nil {
		return m.listByOwner(ctx, ownerID)
	}
	return nil, nil
}

type mockSlotsRepo struct {
	getDetailsByMeeting func(ctx context.Context, meetingID string) ([]SlotResult, error)
}

func (m *mockSlotsRepo) SetSlots(ctx context.Context, participantID string, slotIndexes []int) error {
	return nil
}

func (m *mockSlotsRepo) GetDetailsByMeeting(ctx context.Context, meetingID string) ([]SlotResult, error) {
	if m.getDetailsByMeeting != nil {
		return m.getDetailsByMeeting(ctx, meetingID)
	}
	return nil, nil
}

func TestCreate(t *testing.T) {
	ctx := context.Background()

	t.Run("valid", func(t *testing.T) {
		var passed Meeting
		r := &mockMeetingRepo{
			create: func(_ context.Context, m Meeting) (Meeting, error) {
				passed = m
				m.ID = "m1"
				return m, nil
			},
		}
		m, err := NewService(r, &mockSlotsRepo{}).Create(ctx, "o1", "Meet", "Desc", mustDate("2025-03-01"), mustDate("2025-03-03"), 30)
		if err != nil {
			t.Fatal(err)
		}
		if m.Status != "active" || passed.SlotMinutes != 30 {
			t.Fail()
		}
	})

	t.Run("date_end before date_start", func(t *testing.T) {
		_, err := NewService(&mockMeetingRepo{}, &mockSlotsRepo{}).Create(ctx, "o1", "x", "x", mustDate("2025-03-05"), mustDate("2025-03-01"), 30)
		if err != ErrInvalidDates {
			t.Fatalf("got %v, want ErrInvalidDates", err)
		}
	})

	t.Run("slot_minutes must be 15/30/60", func(t *testing.T) {
		_, err := NewService(&mockMeetingRepo{}, &mockSlotsRepo{}).Create(ctx, "o1", "x", "x", mustDate("2025-03-01"), mustDate("2025-03-03"), 45)
		if err != ErrInvalidSlotMin {
			t.Fatalf("got %v, want ErrInvalidSlotMin", err)
		}
	})
}

func TestUpdate(t *testing.T) {
	r := &mockMeetingRepo{
		getByID: func(_ context.Context, id string) (Meeting, error) { return Meeting{ID: id, OwnerID: "alice"}, nil },
		update:  func(_ context.Context, id, title, _ string) (Meeting, error) { return Meeting{ID: id, Title: title}, nil },
	}
	svc := NewService(r, &mockSlotsRepo{})

	if _, err := svc.Update(context.Background(), "m1", "alice", "New", "d"); err != nil {
		t.Fatal("owner should update:", err)
	}
	if _, err := svc.Update(context.Background(), "m1", "bob", "New", "d"); err != ErrForbidden {
		t.Fatalf("got %v, want ErrForbidden", err)
	}
}

func TestDelete(t *testing.T) {
	r := &mockMeetingRepo{
		getByID: func(_ context.Context, id string) (Meeting, error) { return Meeting{ID: id, OwnerID: "alice"}, nil },
	}
	err := NewService(r, &mockSlotsRepo{}).Delete(context.Background(), "m1", "bob")
	if err != ErrForbidden {
		t.Fatalf("got %v, want ErrForbidden", err)
	}
}

func TestFinalize(t *testing.T) {
	oneDay := func(id string) Meeting {
		return Meeting{ID: id, OwnerID: "o1", DateStart: mustDate("2025-03-01"), DateEnd: mustDate("2025-03-01"), SlotMinutes: 30}
	}

	t.Run("valid slot", func(t *testing.T) {
		r := &mockMeetingRepo{
			getByID:  func(_ context.Context, id string) (Meeting, error) { return oneDay(id), nil },
			finalize: func(_ context.Context, _ string, _ int) error { return nil },
		}
		if err := NewService(r, &mockSlotsRepo{}).Finalize(context.Background(), "m1", "o1", 0); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("slot index out of range", func(t *testing.T) {
		r := &mockMeetingRepo{getByID: func(_ context.Context, id string) (Meeting, error) { return oneDay(id), nil }}
		err := NewService(r, &mockSlotsRepo{}).Finalize(context.Background(), "m1", "o1", 999)
		if err != ErrInvalidSlotIndex {
			t.Fatalf("got %v, want ErrInvalidSlotIndex", err)
		}
	})
}

func TestGetResults(t *testing.T) {
	r := &mockMeetingRepo{
		getByID: func(_ context.Context, id string) (Meeting, error) { return Meeting{ID: id, OwnerID: "alice"}, nil },
	}
	_, err := NewService(r, &mockSlotsRepo{}).GetResults(context.Background(), "m1", "bob")
	if err != ErrForbidden {
		t.Fatalf("got %v, want ErrForbidden", err)
	}
}
