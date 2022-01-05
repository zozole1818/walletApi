package service

import (
	"errors"

	"github.com/labstack/gommon/log"
	"zuzanna.com/walletapi/model"
	"zuzanna.com/walletapi/repository"
)

var ErrBalanceNotFound = errors.New("sender or receiver balances not found")
var ErrBalancesLocked = errors.New("sender or receiver balances are locked, new transaction is not allowed")
var ErrBalanceUnlocked = errors.New("sender or receiver balances are unlocked, inconsistent state")
var ErrInsufficientBalance = errors.New("sender/receiver unlocked or insufficient balance of a sender")
var ErrUnauthorizedTransaction = errors.New("userID from JWT token differ from balance's userID of sender for transaction")

type TransactionService interface {
	Execute(userID int, t model.Transaction) (model.Transaction, error)
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
	transactions, err := svc.repo.GetTransactions(userID)
	if err != nil {
		return nil, err
	}
	return model.ConvertListTransactionDB(transactions), nil
}

// Execute executes transaction that is send specific amount of money from sender balance to receiver balance.
func (svc TransactionServiceImpl) Execute(userID int, t model.Transaction) (model.Transaction, error) {

	err := svc.lockBalances(t.SenderBalanceID, t.ReceiverBalanceID)
	if err != nil {
		return model.Transaction{}, err
	}

	newTransaction, err := svc.makeTransaction(userID, t)
	if err != nil {
		log.Errorf("#Execute(...) error make transaction %+v; error: %v", t, err)

		errFromUpdate := svc.unlockBalances(t.SenderBalanceID, t.ReceiverBalanceID)
		if errFromUpdate != nil {
			return model.Transaction{}, errFromUpdate
		}

		return model.Transaction{}, err
	}
	return model.Transaction(newTransaction), nil
}

func (svc TransactionServiceImpl) makeTransaction(userID int, transaction model.Transaction) (model.TransactionDB, error) {
	return svc.repo.MakeTransaction(model.TransactionDB(transaction), func(t model.TransactionDBFull) (model.TransactionDBFull, error) {
		sender := model.Balance(t.SenderBalance)
		receiver := model.Balance(t.ReceiverBalance)
		transactionFull := model.TransactionFull{
			ID:              t.ID,
			SenderBalance:   &sender,
			ReceiverBalance: &receiver,
			Amount:          t.Amount,
			Currency:        t.Currency,
			Date:            t.Date,
		}
		if userID != t.SenderBalance.UserID {
			log.Warnf("#Execute(...) failed while making transaction, error: %v,", ErrUnauthorizedTransaction)
			return model.TransactionDBFull{}, ErrUnauthorizedTransaction
		}

		if !transactionFull.IsValid() {
			log.Warnf("#Execute(...) failed while making transaction, error: %v", ErrInsufficientBalance)
			return model.TransactionDBFull{}, ErrInsufficientBalance
		}

		transactionFull.Make()

		transactionFull.SenderBalance.Unlock()
		transactionFull.ReceiverBalance.Unlock()

		return model.TransactionDBFull{
			ID:              transactionFull.ID,
			SenderBalance:   model.BalanceDB(*transactionFull.SenderBalance),
			ReceiverBalance: model.BalanceDB(*transactionFull.ReceiverBalance),
			Amount:          transactionFull.Amount,
			Currency:        transactionFull.Currency,
			Date:            transactionFull.Date,
		}, nil
	})
}

func (svc TransactionServiceImpl) lockBalances(senderID, balanceID int) error {
	err := svc.repo.UpdateBalances([]int{senderID, balanceID}, func(bs []model.BalanceDB) ([]model.BalanceDB, error) {
		balances := model.ConvertListBalanceDB(bs)
		if balances[0].IsLocked() || balances[1].IsLocked() {
			return nil, ErrBalancesLocked
		}

		balances[0].Lock()
		balances[1].Lock()

		return model.ConvertListBalance(balances), nil
	})
	if err != nil {
		if err == repository.ErrBalancesNotFound {
			return ErrBalanceNotFound
		}
		log.Errorf("#Execute(...) error cannot acquire lock for sender/receiver balance; error: %v", err)
		return err
	}
	return nil
}

func (svc TransactionServiceImpl) unlockBalances(senderID, balanceID int) error {
	errFromUpdate := svc.repo.UpdateBalances([]int{senderID, balanceID}, func(bs []model.BalanceDB) ([]model.BalanceDB, error) {
		balances := model.ConvertListBalanceDB(bs)
		if !balances[0].IsLocked() || !balances[1].IsLocked() {
			log.Error(ErrBalanceUnlocked.Error())
			return nil, ErrBalanceUnlocked
		}

		balances[0].Unlock()
		balances[1].Unlock()

		return model.ConvertListBalance(balances), nil
	})
	if errFromUpdate != nil {
		log.Errorf("#UpdateBalances(...) error cannot remove lock for sender/receiver balance; error: %v", errFromUpdate)
		return errFromUpdate
	}
	return nil
}
