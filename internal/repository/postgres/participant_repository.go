package postgres

import (
	"context"
	"database/sql"

	"nis-pipo/internal/participant"
)

type ParticipantRepo struct {
	db *sql.DB
}

func NewParticipantRepo(db *sql.DB) *ParticipantRepo {
	return &ParticipantRepo{db: db}
}

func (repo *ParticipantRepo) Create(ctx context.Context, meetingID, displayName string) (participant.Participant, error) {
	const q = `INSERT INTO participants (meeting_id, display_name)
		VALUES ($1::uuid, $2)
		RETURNING id::text, meeting_id::text, display_name, created_at`
	var p participant.Participant
	err := repo.db.QueryRowContext(ctx, q, meetingID, displayName).
		Scan(&p.ID, &p.MeetingID, &p.DisplayName, &p.CreatedAt)
	return p, err
}

func (repo *ParticipantRepo) GetByMeetingAndID(ctx context.Context, meetingID, participantID string) (participant.Participant, error) {
	const q = `SELECT id::text, meeting_id::text, display_name, created_at
		FROM participants WHERE meeting_id = $1::uuid AND id = $2::uuid`
	var p participant.Participant
	err := repo.db.QueryRowContext(ctx, q, meetingID, participantID).
		Scan(&p.ID, &p.MeetingID, &p.DisplayName, &p.CreatedAt)
	return p, err
}
