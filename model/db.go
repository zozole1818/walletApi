package model

import "time"

type BalanceDB struct {
	ID       int
	Currency Currency
	Balance  float64
	Locked   bool
	UserID   int
}

func ConvertListBalance(from []Balance) []BalanceDB {
	arr := []BalanceDB{}
	for _, b := range from {
		arr = append(arr, BalanceDB(b))
	}
	return arr
}

type TransactionDB struct {
	ID                int
	SenderBalanceID   int
	ReceiverBalanceID int
	Amount            float64
	Currency          Currency
	Date              time.Time
}

type TransactionDBFull struct {
	ID              int
	SenderBalance   BalanceDB
	ReceiverBalance BalanceDB
	Amount          float64
	Currency        Currency
	Date            time.Time
}
