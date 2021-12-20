package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/labstack/gommon/log"
	"zuzanna.com/walletapi/model"
)

var ErrBalancesNotFound = errors.New("missing required balances")

type BalanceRepo interface {
	GetList(userID int) ([]*model.Balance, error)
	UpdateBalances(balanceIDs []int, updateFn func(b []*model.Balance) ([]*model.Balance, error)) error

	// maybe change model.Transaction to model.TransactionRequest?
	MakeTransaction(t model.Transaction, fn func(t *model.TransactionDDD) (*model.TransactionDDD, error)) (model.Transaction, error)

	GetTransactions(userID int) ([]model.Transaction, error)
}

type PostgreBalanceRepo struct {
	*pgxpool.Pool
}

func NewPostgreBalanceRepo(pool *pgxpool.Pool) *PostgreBalanceRepo {
	return &PostgreBalanceRepo{Pool: pool}
}

// Get retrieves all balances assigned to particular user.
func (r PostgreBalanceRepo) GetList(userID int) ([]*model.Balance, error) {
	balances := []*model.Balance{}
	rows, err := r.Pool.Query(context.Background(), "SELECT id, currency, balance, user_id FROM balance WHERE user_id=$1", userID)
	if err != nil {
		log.Errorf("error while retrieving balances for user with ID %d; error %w", userID, err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		tmp := model.Balance{}
		err = rows.Scan(&tmp.ID, &tmp.Currency, &tmp.Balance, &tmp.UserID)
		if err != nil {
			log.Errorf("error while reading balances for user with ID %d; error %w", userID, err)
			return nil, err
		}
		balances = append(balances, &tmp)
	}
	return balances, nil
}

func (r PostgreBalanceRepo) GetTransactions(userID int) ([]model.Transaction, error) {
	tx, err := r.Pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		log.Errorf("#GetTransactions(...) failed, error: %w", err)
		return nil, fmt.Errorf("unable to start a transaction; error: %w", err)
	}
	defer func() {
		_ = r.finishTx(err, tx)
	}()

	existingBalances, err := r.getBalances(tx, userID)
	if err != nil {
		return nil, err
	}
	balanceIDs := []int{}
	for _, b := range existingBalances {
		balanceIDs = append(balanceIDs, b.ID)
	}

	transactions, err := r.getTransactionsByBalanceIDs(tx, balanceIDs)
	if err != nil {
		return nil, err
	}
	return transactions, nil

}

func (r PostgreBalanceRepo) getTransactionsByBalanceIDs(tx pgx.Tx, balanceIDs []int) ([]model.Transaction, error) {
	transactions := []model.Transaction{}
	query :=
		`select bt.transaction_id, t.sender_id, t.receiver_id, t.currency, t.amount, t."date" 
		from balance_transaction bt
			left join "transaction" t ON bt.transaction_id = t.id
			where bt.balance_id IN (`
	for i := range balanceIDs {
		query += " $" + strconv.Itoa(i+1)
		if i < len(balanceIDs)-1 {
			query += ","
		}
	}
	query += ")"
	rows, err := tx.Query(context.Background(), query, toArgs(balanceIDs)...)

	if err != nil {
		if err == pgx.ErrNoRows {
			return make([]model.Transaction, 0), nil
		}
		log.Errorf("#getTransactionsByBalanceIDs(...) error while retrieving transactions for balances with IDs %s; error %w", balanceIDs, err)
		return nil, err
	}

	for rows.Next() {
		tmp := model.Transaction{}
		err = rows.Scan(&tmp.ID, &tmp.SenderBalanceID, &tmp.ReceiverBalanceID, &tmp.Currency, &tmp.Amount, &tmp.Date)
		if err != nil {
			log.Errorf("#getTransactionsByBalanceIDs(...) error while scanning transactions for balances with IDs %s; error %w", balanceIDs, err)
			return nil, err
		}
		transactions = append(transactions, tmp)
	}
	return transactions, nil
}

// UpdateBalance update couple of balances by applying updateFn. All actions than happen here are included in one transaction.
func (r PostgreBalanceRepo) UpdateBalances(IDs []int, updateFn func(bs []*model.Balance) ([]*model.Balance, error)) error {
	tx, err := r.Pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		log.Errorf("#UpdateBalances(...) failed, error: %w", err)
		return fmt.Errorf("unable to start a transaction; error: %w", err)
	}
	defer func() {
		_ = r.finishTx(err, tx)
	}()

	existingBalances, err := r.getBalances(tx, IDs...)
	if err != nil {
		return err
	}
	if len(existingBalances) != len(IDs) {
		log.Errorf("#UpdateBalances(...) failed, found %d balance(s) instead of %d", len(existingBalances), len(IDs))
		return ErrBalancesNotFound
	}

	updatedBalance, err := updateFn(existingBalances)
	if err != nil {
		return err
	}

	err = r.saveBalances(tx, updatedBalance)
	if err != nil {
		return err
	}

	return nil
}

func (r PostgreBalanceRepo) MakeTransaction(t model.Transaction, fn func(t *model.TransactionDDD) (*model.TransactionDDD, error)) (model.Transaction, error) {
	tx, err := r.Pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		log.Errorf("#MakeTransaction(...) failed, error: %w", err)
		return model.Transaction{}, fmt.Errorf("unable to start a transaction; error: %w", err)
	}
	defer func() {
		_ = r.finishTx(err, tx)
	}()

	existingBalances, err := r.getBalances(tx, t.SenderBalanceID, t.ReceiverBalanceID)
	if err != nil {
		return model.Transaction{}, err
	}

	var transaction *model.TransactionDDD
	if existingBalances[0].ID == t.SenderBalanceID {
		transaction = &model.TransactionDDD{
			SenderBalance:   existingBalances[0],
			ReceiverBalance: existingBalances[1],
			Amount:          t.Amount,
			Currency:        t.Currency,
		}
	} else {
		transaction = &model.TransactionDDD{
			SenderBalance:   existingBalances[1],
			ReceiverBalance: existingBalances[0],
			Amount:          t.Amount,
			Currency:        t.Currency,
		}
	}

	transaction, err = fn(transaction)

	newTransaction, err := r.createTransaction(tx, *transaction)
	if err != nil {
		return model.Transaction{}, err
	}

	err = r.saveBalances(tx, []*model.Balance{transaction.SenderBalance, transaction.ReceiverBalance})
	if err != nil {
		return model.Transaction{}, err
	}

	return newTransaction, nil
}

func (r PostgreBalanceRepo) createTransaction(tx pgx.Tx, t model.TransactionDDD) (model.Transaction, error) {
	var tID int
	err := tx.QueryRow(context.Background(),
		"INSERT INTO transaction (sender_id, receiver_id, currency, amount, date) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		t.SenderBalance.ID, t.ReceiverBalance.ID, string(t.Currency), t.Amount, t.Date).Scan(&tID)
	if err != nil {
		log.Errorf("#createTransaction(...) error while inserting into transaction table: %w", err)
		return model.Transaction{}, err
	}
	transaction := model.Transaction{
		ID:                tID,
		SenderBalanceID:   t.SenderBalance.ID,
		ReceiverBalanceID: t.ReceiverBalance.ID,
		Amount:            t.Amount,
		Currency:          t.Currency,
		Date:              t.Date,
	}

	_, err = tx.Exec(context.Background(),
		"INSERT INTO balance_transaction (balance_id, transaction_id) VALUES ($1, $2)",
		t.SenderBalance.ID, tID)
	if err != nil {
		log.Errorf("#createTransaction(...) error while inserting into balance_transaction table for balance_i %s: %w", t.SenderBalance.ID, err)
		return model.Transaction{}, err
	}
	_, err = tx.Exec(context.Background(),
		"INSERT INTO balance_transaction (balance_id, transaction_id) VALUES ($1, $2)",
		t.ReceiverBalance.ID, tID)
	if err != nil {
		log.Errorf("#createTransaction(...) error while inserting into balance_transaction table for balance_i %s: %w", t.ReceiverBalance.ID, err)
		return model.Transaction{}, err
	}

	return transaction, err
}

func toArgs(IDs []int) []interface{} {
	ret := make([]interface{}, len(IDs))
	for i, ID := range IDs {
		ret[i] = ID
	}
	return ret
}

func (r PostgreBalanceRepo) getBalances(tx pgx.Tx, IDs ...int) ([]*model.Balance, error) {
	balances := []*model.Balance{}
	query := "SELECT id, currency, balance, locked, user_id FROM balance WHERE id IN ("
	for i := range IDs {
		query += " $" + strconv.Itoa(i+1)
		if i < len(IDs)-1 {
			query += ","
		}
	}
	query += ")"

	log.Warnf("query is: ", query)

	rows, err := tx.Query(context.Background(), query, toArgs(IDs)...)
	if err != nil {
		if err == pgx.ErrNoRows {
			log.Infof("no records for balances with IDs %s", IDs)
			return make([]*model.Balance, 0), nil
		}
		log.Errorf("#getBalances(...) error while retrieving balances with IDs %s; error %w", IDs, err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		tmp := model.Balance{}
		err = rows.Scan(&tmp.ID, &tmp.Currency, &tmp.Balance, &tmp.Locked, &tmp.UserID)
		if err != nil {
			log.Errorf("#getBalances(...) error while scanning balances with IDs %s; error %w", IDs, err)
			return nil, err
		}
		balances = append(balances, &tmp)
	}
	return balances, nil
}

func (r PostgreBalanceRepo) saveBalance(tx pgx.Tx, balance *model.Balance) error {
	_, err := tx.Exec(context.Background(), "UPDATE balance SET balance=$1 WHERE id=$2", balance.Balance, balance.ID)
	if err != nil {
		log.Errorf("#saveBalance(...) error: %w", err)
		return err
	}
	return nil
}

func (r PostgreBalanceRepo) saveBalances(tx pgx.Tx, balance []*model.Balance) error {
	for _, b := range balance {
		err := r.saveBalance(tx, b)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r PostgreBalanceRepo) finishTx(err error, tx pgx.Tx) error {
	if err != nil {
		log.Errorf("#finishTransaction(...) error: %w", err)
		if rollbackErr := tx.Rollback(context.Background()); rollbackErr != nil {
			log.Errorf("#finishTransaction(...) failed when rollback, error: %w", err)
			return fmt.Errorf("unable to rollback a transaction; error: %w", err)
		}

		return err
	}
	if commitErr := tx.Commit(context.Background()); commitErr != nil {
		log.Errorf("finishTransaction failed when commiting, error: %w", err)
		return fmt.Errorf("unable to commit a transaction; error: %w", err)
	}
	return nil
}
