package model

import "testing"

type testCase struct {
	transactionRequest TransactionRequest
	isValid            bool
	errMsg             string
}

func TestIsValid(t *testing.T) {
	cases := []testCase{
		{transactionRequest: TransactionRequest{
			SenderBalanceID:   1,
			ReceiverBalanceID: 2,
			Amount:            10.99,
		},
			isValid: true,
			errMsg:  ""},
		{transactionRequest: TransactionRequest{
			SenderBalanceID:   1,
			ReceiverBalanceID: 2,
			Amount:            0.009,
		},
			isValid: false,
			errMsg:  "amount (rounded down to 2 decimal places) field must be greater then 0"},
		{transactionRequest: TransactionRequest{
			SenderBalanceID:   1,
			ReceiverBalanceID: 2,
			Amount:            -10.99,
		},
			isValid: false,
			errMsg:  "amount (rounded down to 2 decimal places) field must be greater then 0"},
		{transactionRequest: TransactionRequest{
			SenderBalanceID:   0,
			ReceiverBalanceID: -3,
			Amount:            0.99,
		},
			isValid: false,
			errMsg:  "sender or receiver balance not found"},
		{transactionRequest: TransactionRequest{
			SenderBalanceID:   3,
			ReceiverBalanceID: 3,
			Amount:            0.99,
		},
			isValid: false,
			errMsg:  "sender and receiver balances cannot be the same"},
	}
	for _, testCase := range cases {
		ok, err := testCase.transactionRequest.IsValid()
		if ok != testCase.isValid {
			t.Errorf("TransactionRequest.IsValid() got: %t; want: %t", ok, testCase.isValid)
		}
		if testCase.errMsg != "" && err.Error() != testCase.errMsg {
			t.Errorf("TransactionRequest: %+v must be invalid and error must be '%s' instead of '%s'", testCase.transactionRequest, testCase.errMsg, err.Error())
		}
	}
}
