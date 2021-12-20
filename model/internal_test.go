package model

import (
	"math"
	"testing"
	"time"
)

func TestBalanceLocking(t *testing.T) {
	b := Balance{
		Locked: true,
	}
	got := b.IsLocked()
	if !got {
		t.Errorf("IsLocked() init = %t; want true", got)
	}

	b.Unlock()
	got = b.IsLocked()
	if got {
		t.Errorf("IsLocked() after unlocking = %t; want false", got)
	}

	b.Lock()
	got = b.IsLocked()
	if !got {
		t.Errorf("IsLocked() after locking = %t; want true", got)
	}
}

func TestBalanceIncrease(t *testing.T) {
	b := Balance{Balance: 1000.00}
	b.Increase(10)
	got := b.Balance
	if got != 1010.00 {
		t.Errorf("Increase(10) = %f; want 1010.00", got)
	}
	b.Increase(999.99)
	got = b.Balance
	if got != 2009.99 {
		t.Errorf("Increase(999.99) = %f; want 2009.99", got)
	}
	b.Increase(0.01)
	got = b.Balance
	if got != 2010.00 {
		t.Errorf("Increase(0.01) = %f; want 2010.00", got)
	}
}

func TestBalanceDecrease(t *testing.T) {
	b := Balance{Balance: 2000.00}
	b.Decrease(10)
	got := b.Balance
	if got != 1990 {
		t.Errorf("Increase(10) = %f; want 1990", got)
	}
	b.Decrease(999.99)
	got = b.Balance
	if got != 990.01 {
		t.Errorf("Increase(999.99) = %f; want 990.01", got)
	}
	b.Decrease(0.01)
	got = b.Balance
	if got != 990.00 {
		t.Errorf("Increase(0.01) = %f; want 990.00", got)
	}
}

func TestTransactionDDDIsValid(t *testing.T) {
	b1 := Balance{Balance: 2000.00, Locked: false}
	b2 := Balance{Balance: 2000.00, Locked: false}

	transaction := TransactionDDD{
		SenderBalance:   &b1,
		ReceiverBalance: &b2,
		Amount:          250.86,
	}

	isValid := transaction.IsValid()
	if !isValid {
		t.Errorf("IsValid() = %t; want true", isValid)
	}
}

func TestTransactionDDDMake(t *testing.T) {
	b1 := Balance{Balance: 2000.00, Locked: false}
	b2 := Balance{Balance: 2000.00, Locked: false}

	transaction := TransactionDDD{
		SenderBalance:   &b1,
		ReceiverBalance: &b2,
		Amount:          250.86,
	}

	beforeTransaction := transaction.Date

	transaction.Make()
	if !withTolerane(transaction.SenderBalance.Balance, 1749.14) {
		t.Errorf("transaction.Make(); SenderBalance.Balance = %f; want 1749.14", transaction.SenderBalance.Balance)
	}
	if !withTolerane(transaction.ReceiverBalance.Balance, 2250.86) {
		t.Errorf("transaction.Make(); ReceiverBalance.Balance = %f; want 2250.86", transaction.ReceiverBalance.Balance)
	}
	if !transaction.Date.After(beforeTransaction) {
		t.Errorf("transaction.Make(); Date = %s; want after %s", transaction.Date.Format(time.RFC3339Nano), beforeTransaction.Format(time.RFC3339Nano))
	}
}

func withTolerane(a, b float64) bool {
	tolerance := 0.01
	if diff := math.Abs(a - b); diff < tolerance {
		return true
	} else {
		return false
	}
}
