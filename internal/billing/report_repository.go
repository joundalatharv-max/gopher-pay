package billing

import (
	"context"
	"database/sql"
)

type ReportRepository interface {
	GetTransactionsByAccount(ctx context.Context, accountNumber string) (*sql.Rows, error)
	GetAllTransactions(ctx context.Context) (*sql.Rows, error)
}

type PostgresReportRepository struct {
	db *sql.DB
}

func NewPostgresReportRepository(db *sql.DB) *PostgresReportRepository {
	return &PostgresReportRepository{db: db}
}

// func (r *PostgresReportRepository) GetTransactionsByAccount(
// 	ctx context.Context,
// 	accountNumber string,
// ) (*sql.Rows, error) {

// 	query := `
// 		SELECT id, from_account_number, to_account_number, amount, status, request_id, created_at
// 		FROM transactions
// 		WHERE from_account_number = $1 OR to_account_number = $1
// 		ORDER BY created_at ASC
// 	`

// 	return r.db.QueryContext(ctx, query, accountNumber)
// }

func (r *PostgresReportRepository) GetTransactionsByAccount(
	ctx context.Context,
	accountNumber string,
) (*sql.Rows, error) {

	query := `
		SELECT 
			t.id,
			f.account_number AS from_account,
			ta.account_number AS to_account,
			t.amount,
			t.status,
			t.request_id,
			t.created_at
		FROM transactions t
		JOIN accounts f ON t.from_account_id = f.id
		JOIN accounts ta ON t.to_account_id = ta.id
		WHERE f.account_number = $1 
		   OR ta.account_number = $1
		ORDER BY t.created_at ASC
	`

	return r.db.QueryContext(ctx, query, accountNumber)
}

func (r *PostgresReportRepository) GetAllTransactions(ctx context.Context) (*sql.Rows, error) {

	query := `
		SELECT 
			t.id,
			f.account_number AS from_account,
			ta.account_number AS to_account,
			t.amount,
			t.status,
			t.request_id,
			t.created_at
		FROM transactions t
		JOIN accounts f ON t.from_account_id = f.id
		JOIN accounts ta ON t.to_account_id = ta.id
		ORDER BY t.created_at ASC
	`

	return r.db.QueryContext(ctx, query)
}
