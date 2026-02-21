package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"gopherpay/internal/billing"
	"gopherpay/internal/wallet"
	"gopherpay/internal/worker"
)

type Handler struct {
	Pool   *worker.WorkerPool
	Wallet *wallet.WalletService
	Report *billing.ReportService
}

// AdminTransactions returns all transactions or those for a specific account
// GET /admin/transactions?account_number=ACC1234
func (h *Handler) AdminTransactions(w http.ResponseWriter, r *http.Request) {
	acctNum := r.URL.Query().Get("account_number")

	txs, err := h.Report.FetchTransactions(r.Context(), acctNum)
	if err != nil {
		http.Error(w, "failed to fetch transactions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(txs)
}

type TransferRequest struct {
	FromAccount string `json:"from_account"`
	ToAccount   string `json:"to_account"`
	Amount      int64  `json:"amount"`
}

type TransferResponse struct {
	RequestID string `json:"request_id"`
	Status    string `json:"status"`
	Message   string `json:"message,omitempty"`
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

	// Prevent self-transfer at API level so caller gets immediate feedback
	if req.FromAccount == req.ToAccount {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "cannot transfer to the same account"})
		return
	}

	requestID := r.Context().Value(RequestIDKey).(string)

	// If caller requests synchronous processing (e.g. ?sync=1), run transfer inline
	if r.URL.Query().Get("sync") == "1" {
		err := h.Wallet.Transfer(r.Context(), req.FromAccount, req.ToAccount, req.Amount, requestID)
		if err != nil {
			// Map common errors to HTTP status codes
			status := http.StatusInternalServerError
			msg := err.Error()
			if msg == "insufficient funds" {
				status = http.StatusBadRequest
			} else if msg == "from account not found" || msg == "to account not found" {
				status = http.StatusNotFound
			} else if msg == "cannot transfer to the same account" {
				status = http.StatusBadRequest
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(status)
			json.NewEncoder(w).Encode(TransferResponse{
				RequestID: requestID,
				Status:    "failed",
				Message:   msg,
			})
			return
		}

		// success
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TransferResponse{
			RequestID: requestID,
			Status:    "completed",
			Message:   "transfer completed",
		})
		return
	}

	// Asynchronous: queue job on worker pool
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

type CreateAccountRequest struct {
	AccountNumber string `json:"account_number"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	Phone         string `json:"phone"`
	DOB           string `json:"dob"` // date-only YYYY-MM-DD
	Balance       int64  `json:"balance"`
}

type CreateAccountResponse struct {
	ID            int64  `json:"id"`
	AccountNumber string `json:"account_number"`
	Message       string `json:"message,omitempty"`
}

func (h *Handler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	// Support multiple HTTP methods on /accounts
	switch r.Method {
	case http.MethodPost:
		var req CreateAccountRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		if req.AccountNumber == "" || req.Name == "" {
			http.Error(w, "account_number and name are required", http.StatusBadRequest)
			return
		}

		var dob time.Time
		if req.DOB != "" {
			t, err := time.Parse("2006-01-02", req.DOB)
			if err != nil {
				http.Error(w, "dob must be YYYY-MM-DD", http.StatusBadRequest)
				return
			}
			dob = t
		}

		acc := &wallet.Account{
			AccountNumber: req.AccountNumber,
			Name:          req.Name,
			Email:         req.Email,
			Phone:         req.Phone,
			DOB:           dob,
			Balance:       req.Balance,
		}

		if err := h.Wallet.CreateAccount(r.Context(), acc); err != nil {
			slog.Error("create account failed", "error", err, "account_number", acc.AccountNumber)
			http.Error(w, "failed to create account", http.StatusInternalServerError)
			return
		}

		resp := CreateAccountResponse{
			ID:            acc.ID,
			AccountNumber: acc.AccountNumber,
			Message:       "account created",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)

	case http.MethodGet:
		// GET /accounts?account_number=ACC123
		acctNum := r.URL.Query().Get("account_number")
		if acctNum == "" {
			http.Error(w, "account_number query param required", http.StatusBadRequest)
			return
		}

		acc, err := h.Wallet.GetAccountByNumber(r.Context(), acctNum)
		if err != nil {
			http.Error(w, "account not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(acc)

	case http.MethodPut:
		// Update account via body (must include account_number)
		var req CreateAccountRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		if req.AccountNumber == "" {
			http.Error(w, "account_number is required", http.StatusBadRequest)
			return
		}

		var dob time.Time
		if req.DOB != "" {
			t, err := time.Parse("2006-01-02", req.DOB)
			if err != nil {
				http.Error(w, "dob must be YYYY-MM-DD", http.StatusBadRequest)
				return
			}
			dob = t
		}

		acc := &wallet.Account{
			AccountNumber: req.AccountNumber,
			Name:          req.Name,
			Email:         req.Email,
			Phone:         req.Phone,
			DOB:           dob,
			Balance:       req.Balance,
		}

		if err := h.Wallet.UpdateAccount(r.Context(), acc); err != nil {
			slog.Error("update account failed", "error", err, "account_number", acc.AccountNumber)
			http.Error(w, "failed to update account", http.StatusInternalServerError)
			return
		}

		// Return success message and account id
		resp := CreateAccountResponse{

			AccountNumber: acc.AccountNumber,
			Message:       "account updated",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)

	case http.MethodDelete:
		acctNum := r.URL.Query().Get("account_number")
		if acctNum == "" {
			http.Error(w, "account_number query param required", http.StatusBadRequest)
			return
		}

		if err := h.Wallet.DeleteAccount(r.Context(), acctNum); err != nil {
			slog.Error("delete account failed", "error", err, "account_number", acctNum)
			http.Error(w, "failed to delete account", http.StatusInternalServerError)
			return
		}

		resp := CreateAccountResponse{
			AccountNumber: acctNum,
			Message:       "account deleted",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
