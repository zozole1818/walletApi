package model

import "testing"

func TestTransactionRequestIsValid(t *testing.T) {
	tr := TransactionRequest{
		SenderBalanceID:   1,
		ReceiverBalanceID: 2,
		Amount:            10.99,
	}
	ok, err := tr.IsValid()
	if !ok || err != nil {
		t.Errorf("TransactionRequest: %+v must be valid", tr)
	}
}

func TestTransactionRequestInvalidAmount(t *testing.T) {
	tr := TransactionRequest{
		SenderBalanceID:   1,
		ReceiverBalanceID: 2,
		Amount:            -10.99,
	}
	ok, err := tr.IsValid()
	if ok {
		t.Errorf("TransactionRequest: %+v must be invalid", tr)
	}
	errMsg := "amount field must be greater then 0"
	if err == nil || err.Error() != errMsg {
		t.Errorf("TransactionRequest: %+v must be invalid and error must be '%s' instead of '%s'", tr, errMsg, err.Error())
	}
}

func TestTransactionRequestInvalidSenderOrReceiver(t *testing.T) {
	tr := TransactionRequest{
		SenderBalanceID:   0,
		ReceiverBalanceID: -3,
		Amount:            0.99,
	}
	ok, err := tr.IsValid()
	if ok {
		t.Errorf("TransactionRequest: %+v must be invalid", tr)
	}
	errMsg := "sender or receiver balance not found"
	if err == nil || err.Error() != errMsg {
		t.Errorf("TransactionRequest: %+v must be invalid and error must be '%s' instead of '%s'", tr, errMsg, err.Error())
	}
}

func TestTransactionRequestInvalidSenderTheSameAsReceiver(t *testing.T) {
	tr := TransactionRequest{
		SenderBalanceID:   3,
		ReceiverBalanceID: 3,
		Amount:            0.99,
	}
	ok, err := tr.IsValid()
	if ok {
		t.Errorf("TransactionRequest: %+v must be invalid", tr)
	}
	errMsg := "sender and receiver balances cannot be the same"
	if err == nil || err.Error() != errMsg {
		t.Errorf("TransactionRequest: %+v must be invalid and error must be '%s' instead of '%s'", tr, errMsg, err.Error())
	}
}
