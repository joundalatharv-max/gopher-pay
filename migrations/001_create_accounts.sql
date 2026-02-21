-- accounts table
CREATE TABLE IF NOT EXISTS accounts (
	id BIGSERIAL PRIMARY KEY,
	account_number TEXT UNIQUE NOT NULL,
	name TEXT,
	email TEXT,
	phone TEXT,
	dob DATE,
	balance BIGINT NOT NULL DEFAULT 0,
	created_at TIMESTAMP NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_accounts_account_number ON accounts(account_number);
