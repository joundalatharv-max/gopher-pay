package wallet

import (
	"context"
	"database/sql"
)

type WalletRepository interface {

	// Get account by account_number
	GetAccountByNumber(
		ctx context.Context,
		accountNumber string,
	) (*Account, error)

	// Lock account row FOR UPDATE inside transaction
	GetAccountForUpdateByID(
		ctx context.Context,
		tx *sql.Tx,
		accountID int64,
	) (*Account, error)

	// Update account balance
	UpdateBalance(
		ctx context.Context,
		tx *sql.Tx,
		accountID int64,
		newBalance int64,
	) error

	// Get account by account_number (inside transaction)
	GetAccountByNumberTx(
		ctx context.Context,
		tx *sql.Tx,
		accountNumber string,
	) (*Account, error)

	CreateTransaction(
		ctx context.Context,
		tx *sql.Tx,
		fromID int64,
		toID int64,
		amount int64,
		status string,
		requestID string,
	) error

	UpdateTransactionStatus(
		ctx context.Context,
		requestID string,
		status string,
	) error
}

// Create transaction record
// CreateTransaction(
// 	ctx context.Context,
// 	tx *sql.Tx,
// 	fromAccountID int64,
// 	toAccountID int64,
// 	amount int64,
// 	requestID string,
// ) error
