package postgres

import (
	"context"
	"database/sql"

	"nis-pipo/internal/meeting"
)

type MeetingRepo struct {
	db *sql.DB
}

func NewMeetingRepo(db *sql.DB) *MeetingRepo {
	return &MeetingRepo{db: db}
}

func (repo *MeetingRepo) Create(ctx context.Context, m meeting.Meeting) (meeting.Meeting, error) {
	const query = `
		INSERT INTO meetings (owner_id, title, description, date_start, date_end, slot_minutes, status, final_slot_index)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id::text, owner_id::text, title, description, date_start, date_end, slot_minutes, status, final_slot_index, created_at, updated_at`
	var out meeting.Meeting
	err := repo.db.QueryRowContext(ctx, query,
		m.OwnerID, m.Title, m.Description, m.DateStart, m.DateEnd, m.SlotMinutes, m.Status, m.FinalSlotIndex,
	).Scan(&out.ID, &out.OwnerID, &out.Title, &out.Description, &out.DateStart, &out.DateEnd, &out.SlotMinutes, &out.Status, &out.FinalSlotIndex, &out.CreatedAt, &out.UpdatedAt)
	return out, err
}

func (repo *MeetingRepo) GetByID(ctx context.Context, id string) (meeting.Meeting, error) {
	const query = `
		SELECT id::text, owner_id::text, title, description, date_start, date_end, slot_minutes, status, final_slot_index, created_at, updated_at
		FROM meetings
		WHERE id = $1`
	var m meeting.Meeting
	err := repo.db.QueryRowContext(ctx, query, id).Scan(
		&m.ID, &m.OwnerID, &m.Title, &m.Description, &m.DateStart, &m.DateEnd,
		&m.SlotMinutes, &m.Status, &m.FinalSlotIndex, &m.CreatedAt, &m.UpdatedAt,
	)
	return m, err
}

func (repo *MeetingRepo) Update(ctx context.Context, id, title, description string) (meeting.Meeting, error) {
	const query = `
		UPDATE meetings SET title = $2, description = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING id::text, owner_id::text, title, description, date_start, date_end, slot_minutes, status, final_slot_index, created_at, updated_at`
	var m meeting.Meeting
	err := repo.db.QueryRowContext(ctx, query, id, title, description).Scan(
		&m.ID, &m.OwnerID, &m.Title, &m.Description, &m.DateStart, &m.DateEnd,
		&m.SlotMinutes, &m.Status, &m.FinalSlotIndex, &m.CreatedAt, &m.UpdatedAt,
	)
	return m, err
}

func (repo *MeetingRepo) Delete(ctx context.Context, id string) error {
	_, err := repo.db.ExecContext(ctx, `DELETE FROM meetings WHERE id = $1`, id)
	return err
}
