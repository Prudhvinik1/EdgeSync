CREATE TABLE sync_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    device_id UUID REFERENCES devices(id) ON DELETE SET NULL,
    event_type VARCHAR(50) NOT NULL,
    state_key VARCHAR(255),
    sequence_number BIGSERIAL,
    payload BYTEA,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_sync_events_account_id ON sync_events(account_id, sequence_number);
CREATE INDEX idx_sync_events_device_id ON sync_events(device_id, sequence_number);

