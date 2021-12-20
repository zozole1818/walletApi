package service

import (
	"errors"

	"zuzanna.com/walletapi/model"
	"zuzanna.com/walletapi/repository"
)

// different file?
var ErrNoUserID = errors.New("user ID required")

type BalanceService interface {
	GetByUserID(userID int) ([]*model.Balance, error)
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

// Get return User with coresponding balance.
func (svc BalanceServiceImpl) GetByUserID(userID int) ([]*model.Balance, error) {
	return svc.repo.GetList(userID)
}
