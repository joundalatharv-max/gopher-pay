package billing

import (
	"bufio"
	"context"
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
