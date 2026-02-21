package wallet

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type WalletService struct {
	db   *sql.DB
	repo WalletRepository
}

func NewWalletService(db *sql.DB, repo WalletRepository) *WalletService {
	return &WalletService{
		db:   db,
		repo: repo,
	}
}

// Transfer processes the transfer using requestID provided by API middleware
func (s *WalletService) Transfer(
	ctx context.Context,
	fromAccountNumber string,
	toAccountNumber string,
	amount int64,
	requestID string,
) error {

	if amount <= 0 {
		return errors.New("amount must be positive")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Fetch accounts inside transaction
	fromAccount, err := s.repo.GetAccountByNumberTx(ctx, tx, fromAccountNumber)
	if err != nil {
		// Cannot create FK-safe transaction record if sender doesn't exist
		return fmt.Errorf("from account not found")
	}

	toAccount, err := s.repo.GetAccountByNumberTx(ctx, tx, toAccountNumber)
	if err != nil {
		// mark failed transaction and commit it
		_ = s.repo.CreateTransaction(
			ctx,
			tx,
			fromAccount.ID,
			fromAccount.ID,
			amount,
			"failed",
			requestID,
		)
		commitErr := tx.Commit()
		if commitErr != nil {
			return fmt.Errorf("to account not found, commit error: %w", commitErr)
		}
		err = nil // Clear error so defer doesn't try to rollback
		return fmt.Errorf("to account not found")
	}

	// Lock rows FOR UPDATE
	fromLocked, err := s.repo.GetAccountForUpdateByID(ctx, tx, fromAccount.ID)
	if err != nil {
		return err
	}

	toLocked, err := s.repo.GetAccountForUpdateByID(ctx, tx, toAccount.ID)
	if err != nil {
		return err
	}

	// Check balance
	if fromLocked.Balance < amount {
		// Mark transaction as failed and commit
		_ = s.repo.CreateTransaction(
			ctx,
			tx,
			fromLocked.ID,
			toLocked.ID,
			amount,
			"failed",
			requestID,
		)
		commitErr := tx.Commit()
		if commitErr != nil {
			return fmt.Errorf("insufficient funds, commit error: %w", commitErr)
		}
		err = nil // Clear error so defer doesn't try to rollback
		return errors.New("insufficient funds")
	}

	// Update balances
	err = s.repo.UpdateBalance(
		ctx,
		tx,
		fromLocked.ID,
		fromLocked.Balance-amount,
	)
	if err != nil {
		return err
	}

	err = s.repo.UpdateBalance(
		ctx,
		tx,
		toLocked.ID,
		toLocked.Balance+amount,
	)
	if err != nil {
		return err
	}

	// Mark completed
	err = s.repo.CreateTransaction(
		ctx,
		tx,
		fromLocked.ID,
		toLocked.ID,
		amount,
		"completed",
		requestID,
	)
	if err != nil {
		return err
	}

	// Commit on success
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}
	err = nil // Clear error so defer doesn't try to rollback

	return nil
}

//
// Worker status update helpers
//

func (s *WalletService) MarkTransactionFailed(
	ctx context.Context,
	requestID string,
) error {

	return s.repo.UpdateTransactionStatus(
		ctx,
		requestID,
		"failed",
	)
}

func (s *WalletService) MarkTransactionCompleted(
	ctx context.Context,
	requestID string,
) error {

	return s.repo.UpdateTransactionStatus(
		ctx,
		requestID,
		"completed",
	)
}
