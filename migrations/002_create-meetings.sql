-- +goose Up
CREATE TABLE meetings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(500) NOT NULL,
    description TEXT,
    date_start DATE NOT NULL,
    date_end DATE NOT NULL,
    slot_minutes INT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    final_slot_index INT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_meetings_owner_id ON meetings(owner_id);

-- +goose Down
DROP TABLE IF EXISTS meetings;
