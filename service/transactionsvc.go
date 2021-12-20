package service

import (
	"errors"

	"github.com/labstack/gommon/log"
	"zuzanna.com/walletapi/model"
	"zuzanna.com/walletapi/repository"
)

var ErrTransactionConflict = errors.New("sender or receiver balances are locked, new transaction is not allowed")
var ErrInsufficientBalance = errors.New("sender/receiver locked or insufficient balance of a sender")
var ErrUnauthorizedTransaction = errors.New("userId from JWT token differ from balance sender ID from transaction")

type TransactionService interface {
	Execute(userID int, t model.Transaction) (model.Transaction, error)
	// change to query object?
	Retrieve(userID int) ([]model.Transaction, error)
}

type TransactionServiceImpl struct {
	repo repository.BalanceRepo
}

func NewTransactionService(r repository.BalanceRepo) TransactionService {
	if r == nil {
		panic("repo cannot be nil!")
	}
	return TransactionServiceImpl{repo: r}
}

func (svc TransactionServiceImpl) Retrieve(userID int) ([]model.Transaction, error) {
	return svc.repo.GetTransactions(userID)
}

// Execute executes transaction that is send specific amount of money from sender balance to receiver balance.
func (svc TransactionServiceImpl) Execute(userID int, t model.Transaction) (model.Transaction, error) {

	// acquire locks on balances
	err := svc.repo.UpdateBalances([]int{t.SenderBalanceID, t.ReceiverBalanceID}, func(bs []*model.Balance) ([]*model.Balance, error) {
		if bs[0].IsLocked() || bs[1].IsLocked() {
			return nil, ErrTransactionConflict
		}

		bs[0].Lock()
		bs[1].Lock()

		return bs, nil
	})
	if err != nil {
		log.Errorf("#Execute(...) error cannot acquire lock for sender/receiver balance; error: %w", err)
		return model.Transaction{}, err
	}

	// make transaction
	newTransaction, err := svc.repo.MakeTransaction(t, func(t *model.TransactionDDD) (*model.TransactionDDD, error) {
		if userID != t.SenderBalance.ID {
			log.Warnf("#Execute(...) failed while making transaction, error: %w,", ErrUnauthorizedTransaction)
			return nil, ErrUnauthorizedTransaction
		}

		if !t.IsValid() {
			log.Warnf("#Execute(...) failed while making transaction, error: %w", ErrInsufficientBalance)
			return nil, ErrInsufficientBalance
		}

		t.Make()

		t.SenderBalance.Unlock()
		t.ReceiverBalance.Unlock()
		return t, nil
	})
	if err != nil {
		log.Errorf("#Execute(...) error make transaction %+v; error: %w", t, err)
		// release locks on balances
		err := svc.repo.UpdateBalances([]int{t.SenderBalanceID, t.ReceiverBalanceID}, func(bs []*model.Balance) ([]*model.Balance, error) {
			if bs[0].IsLocked() && bs[1].IsLocked() {
				return nil, ErrTransactionConflict
			}

			bs[0].Unlock()
			bs[1].Unlock()

			return bs, nil
		})
		if err != nil {
			log.Errorf("#Execute(...) error cannot acquire lock for sender/receiver balance; error: %w", err)
			return model.Transaction{}, err
		}
		return model.Transaction{}, err
	}
	return newTransaction, nil
}