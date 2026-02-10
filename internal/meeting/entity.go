package meeting

import "time"

type Meeting struct {
	ID             string
	OwnerID        string
	Title          string
	Description    string
	DateStart      time.Time
	DateEnd        time.Time
	SlotMinutes    int
	Status         string
	FinalSlotIndex *int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
