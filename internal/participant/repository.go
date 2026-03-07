package participant

import "context"

type Repository interface {
	Create(ctx context.Context, meetingID, displayName string) (Participant, error)
	GetByMeetingAndID(ctx context.Context, meetingID, participantID string) (Participant, error)
}
