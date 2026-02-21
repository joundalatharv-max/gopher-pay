-- transactions table
CREATE TABLE IF NOT EXISTS transactions (
    id BIGSERIAL PRIMARY KEY,
    from_account_id BIGINT NOT NULL,
    to_account_id BIGINT NOT NULL,
    amount BIGINT NOT NULL CHECK (amount > 0),
    status TEXT NOT NULL,
    request_id TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_from_account
        FOREIGN KEY(from_account_id)
        REFERENCES accounts(id),

    CONSTRAINT fk_to_account
        FOREIGN KEY(to_account_id)
        REFERENCES accounts(id)
);

