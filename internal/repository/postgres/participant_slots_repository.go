package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"nis-pipo/internal/meeting"
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

	if _, err = tx.ExecContext(ctx, `DELETE FROM participant_slots WHERE participant_id = $1::uuid`, participantID); err != nil {
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
		placeholders[i] = fmt.Sprintf("($1::uuid, $%d)", i+2)
	}
	q := `INSERT INTO participant_slots (participant_id, slot_index) VALUES ` + strings.Join(placeholders, ", ")
	if _, err = tx.ExecContext(ctx, q, args...); err != nil {
		return err
	}
	return tx.Commit()
}

func (repo *ParticipantSlotsRepo) GetDetailsByMeeting(ctx context.Context, meetingID string) ([]meeting.SlotResult, error) {
	const q = `SELECT a.slot_index, COUNT(*) AS cnt,
		COALESCE(json_agg(p.display_name ORDER BY p.display_name), '[]'::json)::text AS participant_names
		FROM participant_slots a
		JOIN participants p ON a.participant_id = p.id
		WHERE p.meeting_id = $1::uuid
		GROUP BY a.slot_index ORDER BY a.slot_index`
	rows, err := repo.db.QueryContext(ctx, q, meetingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []meeting.SlotResult
	for rows.Next() {
		var row meeting.SlotResult
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
