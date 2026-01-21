CREATE TABLE encrypted_states (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    device_id UUID REFERENCES devices(id) ON DELETE SET NULL,
    key VARCHAR(255) NOT NULL,
    state BYTEA NOT NULL,
    nonce BYTEA NOT NULL,
    version BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    UNIQUE(account_id, key)
);

CREATE INDEX idx_encrypted_states_account_id ON encrypted_states(account_id);
CREATE INDEX idx_encrypted_states_device_id ON encrypted_states(device_id);