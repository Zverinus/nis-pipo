package participantslots

import "context"

type SlotCount struct {
	SlotIndex int
	Count     int
}

type SlotDetails struct {
	SlotIndex        int
	Count            int
	ParticipantNames []string
}

type Repository interface {
	SetSlots(ctx context.Context, participantID string, slotIndexes []int) error
	GetByParticipant(ctx context.Context, participantID string) ([]int, error)
	GetCountsByMeeting(ctx context.Context, meetingID string) ([]SlotCount, error)
	GetDetailsByMeeting(ctx context.Context, meetingID string) ([]SlotDetails, error)
	CountByMeetingAndSlot(ctx context.Context, meetingID string, slotIndex int) (int, error)
}
