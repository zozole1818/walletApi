package model

import "time"

type Currency string

const (
	SGD Currency = "SGD"
)

type Balance struct {
	ID       int
	Currency Currency
	Balance  float64
	Locked   bool
	UserID   int
}

func (b *Balance) IsLocked() bool {
	return b.Locked
}

func (b *Balance) Lock() {
	b.Locked = true
}

func (b *Balance) Unlock() {
	b.Locked = false
}

func (b *Balance) Increase(amount float64) {
	b.Balance += amount
}

func (b *Balance) Decrease(amount float64) {
	b.Balance -= amount
}

func ConvertListBalanceDB(from []BalanceDB) []Balance {
	arr := []Balance{}
	for _, b := range from {
		arr = append(arr, Balance(b))
	}
	return arr
}

type Transaction struct {
	ID                int
	SenderBalanceID   int
	ReceiverBalanceID int
	Amount            float64
	Currency          Currency
	Date              time.Time
}

type TransactionFull struct {
	ID              int
	SenderBalance   *Balance
	ReceiverBalance *Balance
	Amount          float64
	Currency        Currency
	Date            time.Time
}

func (t *TransactionFull) IsValid() bool {
	return t.SenderBalance.IsLocked() && t.ReceiverBalance.IsLocked() && t.SenderBalance.Balance > t.Amount
}

func (t *TransactionFull) Make() {
	t.SenderBalance.Decrease(t.Amount)
	t.ReceiverBalance.Increase(t.Amount)
	t.Date = time.Now()
}

func ConvertListTransactionDB(from []TransactionDB) []Transaction {
	arr := []Transaction{}
	for _, b := range from {
		arr = append(arr, Transaction(b))
	}
	return arr
}

type Credentials struct {
	ID       int
	Login    string
	Password string
	UserID   int
}
