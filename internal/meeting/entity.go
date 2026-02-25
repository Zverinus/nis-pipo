package meeting

import "time"

type Meeting struct {
	ID             string    `json:"id"`
	OwnerID        string    `json:"owner_id"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	DateStart      time.Time `json:"date_start"`
	DateEnd        time.Time `json:"date_end"`
	SlotMinutes    int       `json:"slot_minutes"`
	Status         string    `json:"status"`
	FinalSlotIndex *int      `json:"final_slot_index,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type SlotResult struct {
	SlotIndex        int      `json:"slot_index"`
	Count            int      `json:"count"`
	ParticipantNames []string `json:"participant_names,omitempty"`
}
