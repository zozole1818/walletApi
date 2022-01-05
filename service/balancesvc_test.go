package service

import (
	"errors"
	"math/rand"
	"reflect"
	"testing"

	"zuzanna.com/walletapi/model"
	"zuzanna.com/walletapi/repository"
)

type BalanceRepoFake struct {
	db map[int][]model.BalanceDB
}

func newBalanceRepoFake() BalanceRepoFake {
	return BalanceRepoFake{
		db: map[int][]model.BalanceDB{
			1: {{ID: 1, Currency: model.SGD, Balance: 1000, Locked: false, UserID: 1}},
			2: {{ID: 2, Currency: model.SGD, Balance: 25.65, Locked: false, UserID: 2}, {ID: 3, Currency: model.SGD, Balance: 12759.77, Locked: false, UserID: 2}},
			3: {},
		},
	}
}

func (r BalanceRepoFake) GetList(userID int) ([]model.BalanceDB, error) {
	balances, ok := r.db[userID]
	if ok {
		return balances, nil
	} else if userID == 4 {
		return nil, errExpected
	}
	return nil, errors.New("user not found")
}

func (r BalanceRepoFake) UpdateBalances(balanceIDs []int, updateFn func(b []model.BalanceDB) ([]model.BalanceDB, error)) error {
	if balanceIDs[0] == -1 {
		return repository.ErrBalancesNotFound
	}
	var b1 model.BalanceDB
	var b2 model.BalanceDB
	for _, b := range balances {
		if balanceIDs[0] == b.ID {
			b1 = b
		}
		if balanceIDs[1] == b.ID {
			b2 = b
		}
	}
	_, err := updateFn([]model.BalanceDB{b1, b2})
	return err
}

func (r BalanceRepoFake) MakeTransaction(t model.TransactionDB, fn func(t model.TransactionDBFull) (model.TransactionDBFull, error)) (model.TransactionDB, error) {
	transactionFullDB := makeTransactionDBFull(t)
	transactionFullDB, err := fn(transactionFullDB)
	if err != nil {
		return model.TransactionDB{}, err
	}
	t.ID = rand.Intn(9) + 101
	t.Date = transactionFullDB.Date
	return t, nil
}

func (r BalanceRepoFake) GetTransactions(userID int) ([]model.TransactionDB, error) {
	return nil, nil
}

func makeTransactionDBFull(t model.TransactionDB) model.TransactionDBFull {
	return model.TransactionDBFull{
		SenderBalance:   transactionTestCases[t.ID].senderBalance,
		ReceiverBalance: transactionTestCases[t.ID].receiverBalance,
		Amount:          t.Amount,
		Currency:        model.SGD,
	}
}

type balanceTestCase struct {
	userID           int
	expectedBalances []model.Balance
	expectedErr      error
}

func TestGetByUserID(t *testing.T) {
	svc := BalanceServiceImpl{repo: newBalanceRepoFake()}

	testCases := []balanceTestCase{
		{userID: 1,
			expectedBalances: []model.Balance{{ID: 1, Currency: model.SGD, Balance: 1000, Locked: false, UserID: 1}},
			expectedErr:      nil},
		{userID: 2,
			expectedBalances: []model.Balance{{ID: 2, Currency: model.SGD, Balance: 25.65, Locked: false, UserID: 2}, {ID: 3, Currency: model.SGD, Balance: 12759.77, Locked: false, UserID: 2}},
			expectedErr:      nil},
		{userID: 3,
			expectedBalances: []model.Balance{},
			expectedErr:      nil},
		{userID: 4,
			expectedBalances: nil,
			expectedErr:      errExpected},
	}

	for _, testCase := range testCases {
		balanses, err := svc.GetByUserID(testCase.userID)
		if testCase.expectedErr != err {
			t.Errorf("error got: %s; want: %s", err, errExpected)
		}
		for i, b := range testCase.expectedBalances {
			if !reflect.DeepEqual(balanses[i], b) {
				t.Errorf("wrong balance got: %+v; want: %+v", balanses[i], b)
			}
		}
	}
}
