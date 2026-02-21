package participant

import "time"

type Participant struct {
	ID          string
	MeetingID   string
	DisplayName string
	CreatedAt   time.Time
}
