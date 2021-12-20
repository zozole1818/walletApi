package model

import "time"

type BalanceDB struct {
	ID       int
	Currency Currency
	Balance  float64
	Locked   bool
	UserID   int
}

type TransactionDB struct {
	ID                int
	SenderBalanceID   int
	ReceiverBalanceID int
	Amount            float64
	Currency          Currency
	Date              time.Time
}
