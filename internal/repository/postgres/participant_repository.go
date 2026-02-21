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
	const query = `
		INSERT INTO participants (meeting_id, display_name)
		VALUES ($1, $2)
		RETURNING id::text, meeting_id::text, display_name, created_at`
	var p participant.Participant
	err := repo.db.QueryRowContext(ctx, query, meetingID, displayName).Scan(
		&p.ID, &p.MeetingID, &p.DisplayName, &p.CreatedAt,
	)
	return p, err
}

func (repo *ParticipantRepo) GetByID(ctx context.Context, id string) (participant.Participant, error) {
	const query = `
		SELECT id::text, meeting_id::text, display_name, created_at
		FROM participants
		WHERE id = $1`
	var p participant.Participant
	err := repo.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.MeetingID, &p.DisplayName, &p.CreatedAt,
	)
	return p, err
}

func (repo *ParticipantRepo) GetByMeetingAndToken(ctx context.Context, meetingID, token string) (participant.Participant, error) {
	const query = `
		SELECT id::text, meeting_id::text, display_name, created_at
		FROM participants
		WHERE meeting_id = $1 AND id = $2`
	var p participant.Participant
	err := repo.db.QueryRowContext(ctx, query, meetingID, token).Scan(
		&p.ID, &p.MeetingID, &p.DisplayName, &p.CreatedAt,
	)
	return p, err
}
