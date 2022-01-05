package service

import (
	"zuzanna.com/walletapi/model"
	"zuzanna.com/walletapi/repository"
)

type BalanceService interface {
	GetByUserID(userID int) ([]model.Balance, error)
}

type BalanceServiceImpl struct {
	repo repository.BalanceRepo
}

func NewBalanceService(r repository.BalanceRepo) BalanceServiceImpl {
	if r == nil {
		panic("repo cannot be nil!")
	}
	return BalanceServiceImpl{repo: r}
}

func (svc BalanceServiceImpl) GetByUserID(userID int) ([]model.Balance, error) {
	balances, err := svc.repo.GetList(userID)
	if err != nil {
		return nil, err
	}
	return model.ConvertListBalanceDB(balances), nil
}
