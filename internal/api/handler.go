package api

import (
	"encoding/json"
	"net/http"

	"gopherpay/internal/worker"
)

type Handler struct {
	Pool *worker.WorkerPool
}

type TransferRequest struct {
	FromAccount string `json:"from_account"`
	ToAccount   string `json:"to_account"`
	Amount      int64  `json:"amount"`
}

type TransferResponse struct {
	RequestID string `json:"request_id"`
	Status    string `json:"status"`
}

func (h *Handler) Transfer(w http.ResponseWriter, r *http.Request) {

	var req TransferRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// Validate amount is positive integer (cents)
	if req.Amount <= 0 {
		http.Error(w, "amount must be a positive integer (cents)", http.StatusBadRequest)
		return
	}

	// Validate accounts are not empty
	if req.FromAccount == "" || req.ToAccount == "" {
		http.Error(w, "from_account and to_account are required", http.StatusBadRequest)
		return
	}

	requestID := r.Context().Value(RequestIDKey).(string)

	job := worker.TransferJob{
		RequestID:         requestID,
		FromAccountNumber: req.FromAccount,
		ToAccountNumber:   req.ToAccount,
		Amount:            req.Amount,
	}

	// Backpressure: Check if queue is full, return 429 if so
	select {
	case h.Pool.JobQueue <- job:
		// Job successfully queued
		resp := TransferResponse{
			RequestID: requestID,
			Status:    "pending",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)

	default:
		// Queue is full, return 429 Too Many Requests
		http.Error(
			w,
			"transfer queue is full, please retry later",
			http.StatusTooManyRequests,
		)
	}
}
