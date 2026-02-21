package participant

import "context"

type Repository interface {
	Create(ctx context.Context, meetingID, displayName string) (Participant, error)
	GetByID(ctx context.Context, id string) (Participant, error)
	GetByMeetingAndToken(ctx context.Context, meetingID, token string) (Participant, error)
}
