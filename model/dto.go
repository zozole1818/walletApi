package model

import (
	"net/http"
	"time"
)

type TransactionRequest struct {
	SenderBalanceID   int     `json:"senderBalanceId,omitempty" example:"1"`
	ReceiverBalanceID int     `json:"receiverBalanceId,omitempty" example:"2"`
	Amount            float64 `json:"amount,omitempty" example:"20"`
}

type TransactionResponse struct {
	ID                int       `json:"id,omitempty"`
	SenderBalanceID   int       `json:"senderBalanceId,omitempty"`
	ReceiverBalanceID int       `json:"receiverBalanceId,omitempty"`
	Amount            float64   `json:"amount,omitempty"`
	Currency          string    `json:"currency,omitempty"`
	Date              time.Time `json:"date,omitempty"`
}

func NewTransactionResponse(t Transaction) TransactionResponse {
	return TransactionResponse{ID: t.ID, SenderBalanceID: t.SenderBalanceID, ReceiverBalanceID: t.ReceiverBalanceID, Amount: t.Amount, Currency: string(t.Currency), Date: t.Date}
}

func NewTransactionResponses(ts []Transaction) []TransactionResponse {
	ret := make([]TransactionResponse, len(ts))
	for i, t := range ts {
		ret[i] = NewTransactionResponse(t)
	}
	return ret
}

type TokenResponse struct {
	Token string `json:"token,omitempty" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLC....jUxOTd9.1vqDegq6YpbXuI5qrfKDG_-AloRajTBuE1eZCMhU1no"`
}

type ErrResponse struct {
	Code    int       `json:"code,omitempty" example:"401"`
	Message string    `json:"message,omitempty" example:"Unauthorized"`
	Error   string    `json:"error,omitempty" example:"Login failed. Please double check username and password."`
	Date    time.Time `json:"date,omitempty" example:"2021-12-19T15:25:58.907966Z"`
}

func NewErrResponse(code int, message string) ErrResponse {
	return ErrResponse{
		Code:    code,
		Message: http.StatusText(code),
		Error:   message,
		Date:    time.Now(),
	}
}

type BalanceResponse struct {
	ID       int     `json:"id,omitempty" example:"1"`
	Currency string  `json:"currency,omitempty" example:"SGD"`
	Balance  float64 `json:"balance" example:"10000.00"`
}

func NewBalanceResponse(b *Balance) BalanceResponse {
	return BalanceResponse{
		ID:       b.ID,
		Currency: string(b.Currency),
		Balance:  b.Balance,
	}
}

func NewBalanceResponses(bs []*Balance) []BalanceResponse {
	balances := []BalanceResponse{}
	for _, b := range bs {
		balances = append(balances, NewBalanceResponse(b))
	}
	return balances
}
