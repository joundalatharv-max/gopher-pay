package worker

import (
	"context"
	"log/slog"

	"gopherpay/internal/wallet"
)

type WorkerPool struct {
	JobQueue chan TransferJob
	service  *wallet.WalletService
}

func NewWorkerPool(
	service *wallet.WalletService,
	bufferSize int,
) *WorkerPool {

	return &WorkerPool{
		JobQueue: make(chan TransferJob, bufferSize),
		service:  service,
	}
}

func (wp *WorkerPool) Start(ctx context.Context, workerCount int) {

	for i := 0; i < workerCount; i++ {

		go func(workerID int) {

			slog.Info("worker started", "worker_id", workerID)

			for job := range wp.JobQueue {

				err := wp.service.Transfer(
					ctx,
					job.FromAccountNumber,
					job.ToAccountNumber,
					job.Amount,
					job.RequestID,
				)

				if err != nil {

					slog.Error(
						"transfer failed",
						"worker_id", workerID,
						"request_id", job.RequestID,
						"error", err,
					)

					wp.service.MarkTransactionFailed(
						ctx,
						job.RequestID,
					)

				} else {

					wp.service.MarkTransactionCompleted(
						ctx,
						job.RequestID,
					)

					slog.Info(
						"transfer completed",
						"request_id", job.RequestID,
					)
				}
			}

		}(i)
	}
}

// GetQueueLoad returns the current number of jobs in the queue and capacity
func (wp *WorkerPool) GetQueueLoad() (current, capacity int) {
	return len(wp.JobQueue), cap(wp.JobQueue)
}
