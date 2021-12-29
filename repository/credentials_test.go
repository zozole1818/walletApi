package repository

import (
	"context"
	"reflect"
	"testing"

	"github.com/jackc/pgx/v4"
	"github.com/pashagolub/pgxmock"
	"zuzanna.com/walletapi/model"
)

type dbMockPool struct {
	pgxmock.PgxPoolIface
}

func (mock dbMockPool) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error) {
	return mock.Begin(ctx)
}

func TestGet(t *testing.T) {
	mockPool, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockPool.Close()

	mockRepo := PostgreCredentialsRepo{
		DBConn: dbMockPool{mockPool},
	}
	want := model.Credentials{
		Login:    "test11",
		Password: "aGFzbG8=",
		UserID:   1,
	}

	rows := pgxmock.NewRows([]string{"login", "password", "user_id"}).
		AddRow(want.Login, want.Password, want.UserID)
	mockPool.ExpectQuery("SELECT login, password, user_id FROM credentials WHERE login=$1").WithArgs(want.Login).WillReturnRows(rows)

	got, err := mockRepo.Get(want.Login)
	if err != nil {
		t.Errorf("error was not expected while retrieving credentials: %s", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("error got: %+v want: %+v", got, want)
	}

	if err := mockPool.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
