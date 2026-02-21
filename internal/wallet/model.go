package wallet

import "time"

type Account struct {
	ID            int64
	AccountNumber string
	Name          string
	Email         string
	Phone         string
	DOB           time.Time
	Balance       int64
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Transaction struct {
	ID            int64
	FromAccountID int64
	ToAccountID   int64
	Amount        int64
	RequestID     string
	CreatedAt     time.Time
}
