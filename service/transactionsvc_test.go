package service

import (
	"testing"

	"zuzanna.com/walletapi/model"
)

type transactionTestCase struct {
	userID          int
	senderBalance   model.BalanceDB
	receiverBalance model.BalanceDB
	transaction     model.Transaction
	expectedErr     error
}

type lockBalanceTestCase struct {
	b1          model.BalanceDB
	b2          model.BalanceDB
	expectedErr error
}

var balances = []model.BalanceDB{
	{
		ID:       1,
		Currency: model.SGD,
		Balance:  100.77,
		Locked:   true,
		UserID:   1,
	},
	{
		ID:       2,
		Currency: model.SGD,
		Balance:  100.77,
		Locked:   true,
		UserID:   11,
	},
	{
		ID:       22,
		Currency: model.SGD,
		Balance:  100.77,
		Locked:   true,
		UserID:   2,
	},
	{
		ID:       3,
		Currency: model.SGD,
		Balance:  59.45,
		Locked:   true,
		UserID:   5,
	},
	{
		ID:       23,
		Currency: model.SGD,
		Balance:  100.77,
		Locked:   true,
		UserID:   2,
	},
	{
		ID:       45,
		Currency: model.SGD,
		Balance:  55.87,
		Locked:   true,
		UserID:   5,
	},
}

// ID of transaction holds the index in slice - for BalanceRepoFake logic
var transactionTestCases = []transactionTestCase{
	{userID: 1,
		senderBalance:   balances[0],
		receiverBalance: balances[1],
		transaction: model.Transaction{
			ID:                0, // index 0
			SenderBalanceID:   balances[0].ID,
			ReceiverBalanceID: balances[1].ID,
			Amount:            10.34,
			Currency:          model.SGD,
		},
		expectedErr: nil},
	{userID: 1,
		senderBalance:   balances[2],
		receiverBalance: balances[3],
		transaction: model.Transaction{
			ID:                1, // index 1
			SenderBalanceID:   balances[2].ID,
			ReceiverBalanceID: balances[3].ID,
			Amount:            10.34,
			Currency:          model.SGD,
		},
		expectedErr: ErrUnauthorizedTransaction},
	{userID: 2,
		senderBalance:   balances[4],
		receiverBalance: balances[5],
		transaction: model.Transaction{
			ID:                2, // index 2
			SenderBalanceID:   balances[4].ID,
			ReceiverBalanceID: balances[5].ID,
			Amount:            1000.85,
			Currency:          model.SGD,
		},
		expectedErr: ErrInsufficientBalance},
}

func TestMakeTransaction(t *testing.T) {
	svc := TransactionServiceImpl{newBalanceRepoFake()}

	for _, test := range transactionTestCases {
		newTransaction, err := svc.makeTransaction(test.userID, test.transaction)
		if err != test.expectedErr {
			t.Errorf("error got: %s; want: %v", err, test.expectedErr)
		}
		if err != nil && (model.TransactionDB{}) != newTransaction {
			t.Errorf("new transaction wrong, got: %+v; want: empty struct", newTransaction)
		}
		if err == nil {
			if newTransaction.ID < 100 {
				t.Errorf("new transaction ID not set got: %d; want: > 100", newTransaction.ID)
			}
			if newTransaction.SenderBalanceID != test.transaction.SenderBalanceID {
				t.Errorf("new transaction SenderBalanceID wrong, got: %d; want: %d", newTransaction.ID, test.transaction.SenderBalanceID)
			}
			if newTransaction.ReceiverBalanceID != test.transaction.ReceiverBalanceID {
				t.Errorf("new transaction ReceiverBalanceID wrong, got: %d; want: %d", newTransaction.ReceiverBalanceID, test.transaction.ReceiverBalanceID)
			}
			if newTransaction.Amount != test.transaction.Amount {
				t.Errorf("new transaction Amount wrong, got: %f; want: %f", newTransaction.Amount, test.transaction.Amount)
			}
			if newTransaction.Currency != test.transaction.Currency {
				t.Errorf("new transaction Currency wrong, got: %s; want: %s", newTransaction.Currency, test.transaction.Currency)
			}
		}
	}
}

func TestLockBalances(t *testing.T) {
	b1 := &balances[0]
	b1.Locked = false
	b2 := &balances[1]
	b2.Locked = false

	b3 := &balances[2]
	b3.Locked = true
	b4 := &balances[3]
	b4.Locked = false

	b5 := &balances[2]
	b5.Locked = false
	b6 := &balances[3]
	b6.Locked = true
	var lockBalanceTestCases = []lockBalanceTestCase{
		{b1: *b1, b2: *b2, expectedErr: nil},
		{b1: *b3, b2: *b4, expectedErr: ErrBalancesLocked},
		{b1: *b5, b2: *b6, expectedErr: ErrBalancesLocked},
	}

	svc := TransactionServiceImpl{newBalanceRepoFake()}

	for _, test := range lockBalanceTestCases {
		err := svc.lockBalances(test.b1.ID, test.b2.ID)
		if err != test.expectedErr {
			t.Errorf("error got: %s; want: %v", err, test.expectedErr)
		}
	}

	err := svc.lockBalances(-1, 0)
	if err != ErrBalanceNotFound {
		t.Errorf("error got: %s; want: %v", err, ErrBalanceNotFound)
	}
}

func TestUnlockBalances(t *testing.T) {
	b1 := &balances[0]
	b1.Locked = true
	b2 := &balances[1]
	b2.Locked = true

	b3 := &balances[2]
	b3.Locked = true
	b4 := &balances[3]
	b4.Locked = false

	b5 := &balances[2]
	b5.Locked = false
	b6 := &balances[3]
	b6.Locked = true
	var lockBalanceTestCases = []lockBalanceTestCase{
		{b1: *b1, b2: *b2, expectedErr: nil},
		{b1: *b3, b2: *b4, expectedErr: ErrBalanceUnlocked},
		{b1: *b5, b2: *b6, expectedErr: ErrBalanceUnlocked},
	}

	svc := TransactionServiceImpl{newBalanceRepoFake()}

	for _, test := range lockBalanceTestCases {
		err := svc.unlockBalances(test.b1.ID, test.b2.ID)
		if err != test.expectedErr {
			t.Errorf("error got: %s; want: %v", err, test.expectedErr)
		}
	}
}
