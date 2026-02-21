package participantslots

import "context"

type Repository interface {
	SetSlots(ctx context.Context, participantID string, slotIndexes []int) error
	GetByParticipant(ctx context.Context, participantID string) ([]int, error)
	CountByMeetingAndSlot(ctx context.Context, meetingID string, slotIndex int) (int, error)
}
