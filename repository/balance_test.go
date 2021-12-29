package repository

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/pashagolub/pgxmock"
	"zuzanna.com/walletapi/model"
)

func TestGetList(t *testing.T) {
	mockPool, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockPool.Close()

	mockRepo := PostgreBalanceRepo{
		DBConn: dbMockPool{mockPool},
	}

	want := []model.Balance{
		{ID: 1, Currency: "SGD", Balance: 1000, UserID: 1},
		{ID: 2, Currency: "SGD", Balance: 25.25, UserID: 1},
	}

	mockPool.ExpectQuery("SELECT id, currency, balance, user_id FROM balance WHERE user_id=$1").
		WithArgs(1).
		WillReturnRows(pgxmock.NewRows([]string{"id", "currency", "balance", "user_id"}).
			AddRow(want[0].ID, want[0].Currency, want[0].Balance, want[0].UserID).
			AddRow(want[1].ID, want[1].Currency, want[1].Balance, want[1].UserID))

	got, err := mockRepo.GetList(1)
	if err != nil {
		t.Errorf("error was not expected while retrieving balances: %s", err)
	}

	for i := range got {
		if !reflect.DeepEqual(*got[i], want[i]) {
			t.Errorf("error got: %+v want: %+v", *got[i], want[i])
		}
	}

	if err := mockPool.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetTransactions(t *testing.T) {
	mockPool, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockPool.Close()

	mockRepo := PostgreBalanceRepo{
		DBConn: dbMockPool{mockPool},
	}

	foundBalance := model.Balance{
		ID: 1, Currency: "SGD", Balance: 1000, UserID: 1, Locked: false,
	}
	want := []model.Transaction{
		{ID: 1, SenderBalanceID: 1, ReceiverBalanceID: 2, Currency: "SGD", Amount: 3.99, Date: time.Now()},
		{ID: 2, SenderBalanceID: 1, ReceiverBalanceID: 4, Currency: "SGD", Amount: 56.85, Date: time.Now()},
	}

	mockPool.ExpectBeginTx(pgx.TxOptions{})
	mockPool.ExpectQuery("SELECT id, currency, balance, locked, user_id FROM balance WHERE id IN ( $1)").
		WithArgs(1).
		WillReturnRows(pgxmock.NewRows([]string{"id", "currency", "balance", "locked", "user_id"}).
			AddRow(foundBalance.ID, foundBalance.Currency, foundBalance.Balance, foundBalance.Locked, foundBalance.UserID))
	mockPool.ExpectQuery(`select bt.transaction_id, t.sender_id, t.receiver_id, t.currency, t.amount, t."date" 
			from balance_transaction bt
				left join "transaction" t ON bt.transaction_id = t.id
				where bt.balance_id IN ( $1)`).
		WithArgs(1).
		WillReturnRows(pgxmock.NewRows([]string{"bt.transaction_id", "t.sender_id", "t.receiver_id", "t.currency", "t.amount", `t."date"`}).
			AddRow(want[0].ID, want[0].SenderBalanceID, want[0].ReceiverBalanceID, want[0].Currency, want[0].Amount, want[0].Date).
			AddRow(want[1].ID, want[1].SenderBalanceID, want[1].ReceiverBalanceID, want[1].Currency, want[1].Amount, want[1].Date))
	mockPool.ExpectCommit()

	got, err := mockRepo.GetTransactions(1)
	if err != nil {
		t.Errorf("error was not expected while retrieving transactions: %s", err)
	}

	for i := range got {
		if !reflect.DeepEqual(got[i], want[i]) {
			t.Errorf("error got: %+v want: %+v", got[i], want[i])
		}
	}

	if err := mockPool.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestUpdateBalances(t *testing.T) {
	mockPool, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockPool.Close()

	mockRepo := PostgreBalanceRepo{
		DBConn: dbMockPool{mockPool},
	}

	found := []model.Balance{
		{ID: 1, Currency: "SGD", Balance: 1000, UserID: 1, Locked: false},
		{ID: 2, Currency: "SGD", Balance: 25.25, UserID: 1, Locked: false},
	}

	mockPool.ExpectBeginTx(pgx.TxOptions{})
	mockPool.ExpectQuery("SELECT id, currency, balance, locked, user_id FROM balance WHERE id IN ( $1, $2)").
		WithArgs(1, 2).
		WillReturnRows(pgxmock.NewRows([]string{"id", "currency", "balance", "locked", "user_id"}).
			AddRow(found[0].ID, found[0].Currency, found[0].Balance, found[0].Locked, found[0].UserID).
			AddRow(found[1].ID, found[1].Currency, found[1].Balance, found[1].Locked, found[1].UserID))
	mockPool.ExpectExec("UPDATE balance SET balance=$1, locked=$2 WHERE id=$3").
		WithArgs(found[0].Balance, true, found[0].ID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mockPool.ExpectExec("UPDATE balance SET balance=$1, locked=$2 WHERE id=$3").
		WithArgs(found[1].Balance, true, found[1].ID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mockPool.ExpectCommit()

	err = mockRepo.UpdateBalances([]int{1, 2}, func(bs []*model.Balance) ([]*model.Balance, error) {
		bs[0].Lock()
		bs[1].Lock()
		return bs, nil
	})
	if err != nil {
		t.Errorf("error was not expected while updating balances: %s", err)
	}

	if err := mockPool.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestMakeTransaction(t *testing.T) {
	mockPool, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockPool.Close()

	mockRepo := PostgreBalanceRepo{
		DBConn: dbMockPool{mockPool},
	}

	transaction := model.Transaction{
		SenderBalanceID:   1,
		ReceiverBalanceID: 2,
		Currency:          "SGD",
		Amount:            11.49,
	}
	beforeTransaction := transaction.Date

	found := []model.Balance{
		{ID: 1, Currency: "SGD", Balance: 1000, UserID: 1, Locked: true},
		{ID: 2, Currency: "SGD", Balance: 25.25, UserID: 1, Locked: true},
	}

	mockPool.ExpectBeginTx(pgx.TxOptions{})
	mockPool.ExpectQuery("SELECT id, currency, balance, locked, user_id FROM balance WHERE id IN ( $1, $2)").
		WithArgs(1, 2).
		WillReturnRows(pgxmock.NewRows([]string{"id", "currency", "balance", "locked", "user_id"}).
			AddRow(found[0].ID, found[0].Currency, found[0].Balance, found[0].Locked, found[0].UserID).
			AddRow(found[1].ID, found[1].Currency, found[1].Balance, found[1].Locked, found[1].UserID))

	mockPool.ExpectQuery("INSERT INTO transaction (sender_id, receiver_id, currency, amount, date) VALUES ($1, $2, $3, $4, $5) RETURNING id").
		WithArgs(transaction.SenderBalanceID, transaction.ReceiverBalanceID, string(transaction.Currency), transaction.Amount, AnyTime{}).
		WillReturnRows(pgxmock.NewRows([]string{"id"}).
			AddRow(1))
	mockPool.ExpectExec("INSERT INTO balance_transaction (balance_id, transaction_id) VALUES ($1, $2)").
		WithArgs(found[0].ID, 1).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mockPool.ExpectExec("INSERT INTO balance_transaction (balance_id, transaction_id) VALUES ($1, $2)").
		WithArgs(found[1].ID, 1).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	mockPool.ExpectExec("UPDATE balance SET balance=$1, locked=$2 WHERE id=$3").
		WithArgs(found[0].Balance-transaction.Amount, true, found[0].ID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mockPool.ExpectExec("UPDATE balance SET balance=$1, locked=$2 WHERE id=$3").
		WithArgs(found[1].Balance+transaction.Amount, true, found[1].ID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mockPool.ExpectCommit()

	got, err := mockRepo.MakeTransaction(transaction, func(tFull *model.TransactionFull) (*model.TransactionFull, error) {
		if !tFull.IsValid() {
			return nil, fmt.Errorf("transaction should be valid but is not")
		}
		tFull.Make()
		return tFull, nil
	})
	if err != nil {
		t.Errorf("error was not expected while making a transaction: %s", err)
	}

	if ok, msg := validateTransactions(got, transaction, beforeTransaction); !ok {
		t.Errorf(msg)
	}

	if err := mockPool.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

type AnyTime struct{}

// Match satisfies sqlmock.Argument interface
func (a AnyTime) Match(v interface{}) bool {
	_, ok := v.(time.Time)
	return ok
}

func validateTransactions(got, want model.Transaction, beforeTransaction time.Time) (bool, string) {
	if got.ID == 0 {
		return false, fmt.Sprintf("transaction ID got: %d; want != 0", got.ID)
	}
	if got.SenderBalanceID != want.SenderBalanceID {
		return false, fmt.Sprintf("senderBalanceID got: %d; want: %d", got.SenderBalanceID, want.SenderBalanceID)
	}
	if got.ReceiverBalanceID != want.ReceiverBalanceID {
		return false, fmt.Sprintf("ReceiverBalanceID got: %d; want: %d", got.ReceiverBalanceID, want.ReceiverBalanceID)
	}
	if got.Amount != want.Amount {
		return false, fmt.Sprintf("Amount got: %f; want: %f", got.Amount, want.Amount)
	}
	if got.Currency != want.Currency {
		return false, fmt.Sprintf("Currency got: %s; want: %s", got.Currency, want.Currency)
	}
	if !got.Date.After(beforeTransaction) {
		return false, fmt.Sprintf("transaction time got: %s; want before: %s", got.Date, beforeTransaction)
	}
	return true, ""
}
