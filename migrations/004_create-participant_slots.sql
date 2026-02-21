-- +goose Up
CREATE TABLE participant_slots (
    participant_id UUID NOT NULL REFERENCES participants(id) ON DELETE CASCADE,
    slot_index INT NOT NULL,
    PRIMARY KEY (participant_id, slot_index)
);

CREATE INDEX idx_participant_slots_participant_id ON participant_slots(participant_id);

-- +goose Down
DROP TABLE IF EXISTS participant_slots;
