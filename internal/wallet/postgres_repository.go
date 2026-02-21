package wallet

import (
	"context"
	"database/sql"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// Get account using account_number (outside transaction)
func (r *PostgresRepository) GetAccountByNumber(
	ctx context.Context,
	accountNumber string,
) (*Account, error) {

	query := `
	SELECT id, account_number, name, email, phone, dob, balance, created_at, updated_at
	FROM accounts
	WHERE account_number = $1
	`

	row := r.db.QueryRowContext(ctx, query, accountNumber)

	var acc Account

	err := row.Scan(
		&acc.ID,
		&acc.AccountNumber,
		&acc.Name,
		&acc.Email,
		&acc.Phone,
		&acc.DOB,
		&acc.Balance,
		&acc.CreatedAt,
		&acc.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &acc, nil
}

// Lock account row FOR UPDATE (inside transaction)
func (r *PostgresRepository) GetAccountForUpdateByID(
	ctx context.Context,
	tx *sql.Tx,
	accountID int64,
) (*Account, error) {

	query := `
	SELECT id, balance
	FROM accounts
	WHERE id = $1
	FOR UPDATE
	`

	row := tx.QueryRowContext(ctx, query, accountID)

	var acc Account

	err := row.Scan(
		&acc.ID,
		&acc.Balance,
	)

	if err != nil {
		return nil, err
	}

	return &acc, nil
}

// Update balance (inside transaction)
func (r *PostgresRepository) UpdateBalance(
	ctx context.Context,
	tx *sql.Tx,
	accountID int64,
	newBalance int64,
) error {

	query := `
	UPDATE accounts
	SET balance = $1,
	    updated_at = now()
	WHERE id = $2
	`

	_, err := tx.ExecContext(
		ctx,
		query,
		newBalance,
		accountID,
	)

	return err
}

// Insert transaction record (inside transaction)
func (r *PostgresRepository) CreateTransaction(
	ctx context.Context,
	tx *sql.Tx,
	fromID int64,
	toID int64,
	amount int64,
	status string,
	requestID string,
) error {

	query := `
	INSERT INTO transactions
	(from_account_id, to_account_id, amount, status, request_id)
	VALUES ($1, $2, $3, $4, $5)
	`

	_, err := tx.ExecContext(
		ctx,
		query,
		fromID,
		toID,
		amount,
		status, // dynamic status (completed / failed)
		requestID,
	)

	return err
}

func (r *PostgresRepository) GetAccountByNumberTx(
	ctx context.Context,
	tx *sql.Tx,
	accountNumber string,
) (*Account, error) {

	query := `
	SELECT id, account_number, balance
	FROM accounts
	WHERE account_number = $1
	`

	row := tx.QueryRowContext(ctx, query, accountNumber)

	var acc Account

	err := row.Scan(
		&acc.ID,
		&acc.AccountNumber,
		&acc.Balance,
	)

	if err != nil {
		return nil, err
	}

	return &acc, nil
}

func (r *PostgresRepository) UpdateTransactionStatus(
	ctx context.Context,
	requestID string,
	status string,
) error {

	query := `
	UPDATE transactions
	SET status = $1
	WHERE request_id = $2
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		status,
		requestID,
	)

	return err
}

// CreateAccount inserts a new account and returns the generated ID
func (r *PostgresRepository) CreateAccount(
	ctx context.Context,
	acc *Account,
) error {

	query := `
	INSERT INTO accounts
	(account_number, name, email, phone, dob, balance, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, now(), now())
	RETURNING id
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		acc.AccountNumber,
		acc.Name,
		acc.Email,
		acc.Phone,
		acc.DOB,
		acc.Balance,
	).Scan(&acc.ID)

	return err
}

// UpdateAccount updates an existing account identified by account_number
func (r *PostgresRepository) UpdateAccount(
	ctx context.Context,
	acc *Account,
) error {

	query := `
	UPDATE accounts
	SET name = $1,
		email = $2,
		phone = $3,
		dob = $4,
		balance = $5,
		updated_at = now()
	WHERE account_number = $6
	RETURNING id
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		acc.Name,
		acc.Email,
		acc.Phone,
		acc.DOB,
		acc.Balance,
		acc.AccountNumber,
	).Scan(&acc.ID)

	return err
}

// DeleteAccount deletes account row by account_number
func (r *PostgresRepository) DeleteAccount(
	ctx context.Context,
	accountNumber string,
) error {

	query := `
	DELETE FROM accounts
	WHERE account_number = $1
	`

	_, err := r.db.ExecContext(ctx, query, accountNumber)
	return err
}
