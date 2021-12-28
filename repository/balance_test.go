package repository

import (
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
		t.Errorf("error was not expected while retrieving balances: %s", err)
	}

	for i := range got {
		if !reflect.DeepEqual(got[i], want[i]) {
			t.Errorf("error got: %+v want: %+v", got[i], want[i])
		}
	}
}
