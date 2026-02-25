package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"nis-pipo/internal/participantslots"
)

type ParticipantSlotsRepo struct {
	db *sql.DB
}

func NewParticipantSlotsRepo(db *sql.DB) *ParticipantSlotsRepo {
	return &ParticipantSlotsRepo{db: db}
}

func (repo *ParticipantSlotsRepo) SetSlots(ctx context.Context, participantID string, slotIndexes []int) error {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err = tx.ExecContext(ctx, `DELETE FROM participant_slots WHERE participant_id = $1`, participantID); err != nil {
		return err
	}
	if len(slotIndexes) == 0 {
		return tx.Commit()
	}

	args := make([]interface{}, 0, len(slotIndexes)+1)
	args = append(args, participantID)
	placeholders := make([]string, len(slotIndexes))
	for i, idx := range slotIndexes {
		args = append(args, idx)
		placeholders[i] = fmt.Sprintf("($1, $%d)", i+2)
	}
	query := `INSERT INTO participant_slots (participant_id, slot_index) VALUES ` + strings.Join(placeholders, ", ")
	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (repo *ParticipantSlotsRepo) GetCountsByMeeting(ctx context.Context, meetingID string) ([]participantslots.SlotCount, error) {
	const query = `
		SELECT a.slot_index, COUNT(*)
		FROM participant_slots a
		JOIN participants p ON a.participant_id = p.id
		WHERE p.meeting_id = $1
		GROUP BY a.slot_index
		ORDER BY a.slot_index`
	rows, err := repo.db.QueryContext(ctx, query, meetingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []participantslots.SlotCount
	for rows.Next() {
		var row participantslots.SlotCount
		if err := rows.Scan(&row.SlotIndex, &row.Count); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (repo *ParticipantSlotsRepo) GetDetailsByMeeting(ctx context.Context, meetingID string) ([]participantslots.SlotDetails, error) {
	const query = `
		SELECT
			a.slot_index,
			COUNT(*) AS cnt,
			COALESCE(json_agg(p.display_name ORDER BY p.display_name), '[]'::json) AS participant_names
		FROM participant_slots a
		JOIN participants p ON a.participant_id = p.id
		WHERE p.meeting_id = $1
		GROUP BY a.slot_index
		ORDER BY a.slot_index`
	rows, err := repo.db.QueryContext(ctx, query, meetingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []participantslots.SlotDetails
	for rows.Next() {
		var row participantslots.SlotDetails
		var namesJSON []byte
		if err := rows.Scan(&row.SlotIndex, &row.Count, &namesJSON); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(namesJSON, &row.ParticipantNames); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (repo *ParticipantSlotsRepo) GetByParticipant(ctx context.Context, participantID string) ([]int, error) {
	rows, err := repo.db.QueryContext(ctx,
		`SELECT slot_index FROM participant_slots WHERE participant_id = $1 ORDER BY slot_index`,
		participantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []int
	for rows.Next() {
		var idx int
		if err := rows.Scan(&idx); err != nil {
			return nil, err
		}
		out = append(out, idx)
	}
	return out, rows.Err()
}

func (repo *ParticipantSlotsRepo) CountByMeetingAndSlot(ctx context.Context, meetingID string, slotIndex int) (int, error) {
	const query = `
		SELECT COUNT(*) FROM participant_slots a
		JOIN participants p ON a.participant_id = p.id
		WHERE p.meeting_id = $1 AND a.slot_index = $2`
	var n int
	err := repo.db.QueryRowContext(ctx, query, meetingID, slotIndex).Scan(&n)
	return n, err
}
