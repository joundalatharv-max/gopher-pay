package worker

type TransferJob struct {
	RequestID         string
	FromAccountNumber string
	ToAccountNumber   string
	Amount            int64
}
