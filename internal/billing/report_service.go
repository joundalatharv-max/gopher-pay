package billing

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/csv"
	"os"
	"strconv"
)

type ReportService struct {
	repo ReportRepository
}

func NewReportService(repo ReportRepository) *ReportService {
	return &ReportService{repo: repo}
}

func (s *ReportService) GenerateReport(
	ctx context.Context,
	accountNumber string,
	filename string,
) error {

	rows, err := s.repo.GetTransactionsByAccount(ctx, accountNumber)
	if err != nil {
		return err
	}
	defer rows.Close()

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(bufio.NewWriter(file))

	// Write header
	writer.Write([]string{
		"id",
		"from_account",
		"to_account",
		"amount",
		"status",
		"request_id",
		"created_at",
	})

	for rows.Next() {
		var (
			id        int64
			fromAcc   string
			toAcc     string
			amount    int64
			status    string
			requestID string
			createdAt string
		)

		if err := rows.Scan(
			&id,
			&fromAcc,
			&toAcc,
			&amount,
			&status,
			&requestID,
			&createdAt,
		); err != nil {
			return err
		}

		writer.Write([]string{
			strconv.FormatInt(id, 10),
			fromAcc,
			toAcc,
			strconv.FormatInt(amount, 10),
			status,
			requestID,
			createdAt,
		})
	}

	writer.Flush()

	return writer.Error()
}

type TransactionView struct {
	ID        int64  `json:"id"`
	From      string `json:"from_account"`
	To        string `json:"to_account"`
	Amount    int64  `json:"amount"`
	Status    string `json:"status"`
	RequestID string `json:"request_id"`
	CreatedAt string `json:"created_at"`
}

// FetchTransactions returns transactions for a specific account or all when accountNumber is empty
func (s *ReportService) FetchTransactions(ctx context.Context, accountNumber string) ([]TransactionView, error) {
	var rows *sql.Rows
	var err error

	if accountNumber == "" {
		rows, err = s.repo.GetAllTransactions(ctx)
	} else {
		rows, err = s.repo.GetTransactionsByAccount(ctx, accountNumber)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []TransactionView

	for rows.Next() {
		var tv TransactionView
		if err := rows.Scan(
			&tv.ID,
			&tv.From,
			&tv.To,
			&tv.Amount,
			&tv.Status,
			&tv.RequestID,
			&tv.CreatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, tv)
	}

	return result, nil
}
