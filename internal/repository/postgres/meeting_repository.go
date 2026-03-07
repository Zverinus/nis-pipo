package postgres

import (
	"context"
	"database/sql"
	"errors"

	"nis-pipo/internal/meeting"
)

type MeetingRepo struct {
	db *sql.DB
}

func NewMeetingRepo(db *sql.DB) *MeetingRepo {
	return &MeetingRepo{db: db}
}

func (repo *MeetingRepo) Create(ctx context.Context, m meeting.Meeting) (meeting.Meeting, error) {
	const q = `INSERT INTO meetings (owner_id, title, description, date_start, date_end, slot_minutes, status, final_slot_index)
		VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id::text, owner_id::text, title, COALESCE(description,''), date_start, date_end, slot_minutes, status, final_slot_index, created_at, updated_at`
	var finalSlot interface{}
	if m.FinalSlotIndex != nil {
		finalSlot = *m.FinalSlotIndex
	}
	var out meeting.Meeting
	var finalSlotOut sql.NullInt32
	err := repo.db.QueryRowContext(ctx, q,
		m.OwnerID, m.Title, m.Description, m.DateStart, m.DateEnd, m.SlotMinutes, m.Status, finalSlot,
	).Scan(&out.ID, &out.OwnerID, &out.Title, &out.Description, &out.DateStart, &out.DateEnd,
		&out.SlotMinutes, &out.Status, &finalSlotOut, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return meeting.Meeting{}, err
	}
	if finalSlotOut.Valid {
		v := int(finalSlotOut.Int32)
		out.FinalSlotIndex = &v
	}
	return out, nil
}

func (repo *MeetingRepo) ListByOwner(ctx context.Context, ownerID string) ([]meeting.Meeting, error) {
	const q = `SELECT id::text, owner_id::text, title, COALESCE(description,''), date_start, date_end, slot_minutes, status, final_slot_index, created_at, updated_at
		FROM meetings WHERE owner_id = $1::uuid ORDER BY created_at DESC`
	rows, err := repo.db.QueryContext(ctx, q, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []meeting.Meeting
	for rows.Next() {
		var m meeting.Meeting
		var finalSlot sql.NullInt32
		if err := rows.Scan(&m.ID, &m.OwnerID, &m.Title, &m.Description, &m.DateStart, &m.DateEnd,
			&m.SlotMinutes, &m.Status, &finalSlot, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		if finalSlot.Valid {
			v := int(finalSlot.Int32)
			m.FinalSlotIndex = &v
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (repo *MeetingRepo) GetByID(ctx context.Context, id string) (meeting.Meeting, error) {
	const q = `SELECT id::text, owner_id::text, title, COALESCE(description,''), date_start, date_end, slot_minutes, status, final_slot_index, created_at, updated_at
		FROM meetings WHERE id = $1::uuid`
	var m meeting.Meeting
	var finalSlot sql.NullInt32
	err := repo.db.QueryRowContext(ctx, q, id).Scan(
		&m.ID, &m.OwnerID, &m.Title, &m.Description, &m.DateStart, &m.DateEnd,
		&m.SlotMinutes, &m.Status, &finalSlot, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return meeting.Meeting{}, meeting.ErrNotFound
		}
		return meeting.Meeting{}, err
	}
	if finalSlot.Valid {
		v := int(finalSlot.Int32)
		m.FinalSlotIndex = &v
	}
	return m, nil
}

func (repo *MeetingRepo) Update(ctx context.Context, id, title, description string) (meeting.Meeting, error) {
	const q = `UPDATE meetings SET title = $2, description = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1::uuid
		RETURNING id::text, owner_id::text, title, COALESCE(description,''), date_start, date_end, slot_minutes, status, final_slot_index, created_at, updated_at`
	var m meeting.Meeting
	var finalSlot sql.NullInt32
	err := repo.db.QueryRowContext(ctx, q, id, title, description).Scan(
		&m.ID, &m.OwnerID, &m.Title, &m.Description, &m.DateStart, &m.DateEnd,
		&m.SlotMinutes, &m.Status, &finalSlot, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		return meeting.Meeting{}, err
	}
	if finalSlot.Valid {
		v := int(finalSlot.Int32)
		m.FinalSlotIndex = &v
	}
	return m, nil
}

func (repo *MeetingRepo) Finalize(ctx context.Context, id string, finalSlotIndex int) error {
	_, err := repo.db.ExecContext(ctx,
		`UPDATE meetings SET status = 'finalized', final_slot_index = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $1::uuid`,
		id, finalSlotIndex,
	)
	return err
}

func (repo *MeetingRepo) Delete(ctx context.Context, id string) error {
	_, err := repo.db.ExecContext(ctx, `DELETE FROM meetings WHERE id = $1::uuid`, id)
	return err
}
