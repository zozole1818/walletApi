package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"zuzanna.com/walletapi/model"
	"zuzanna.com/walletapi/service"
)

var ErrCannotParseBodyMsg = "Could not parse body request. Please doble check the JSON."
var ErrUnauthorizedTransactionMsg = "User has no privilages to make requested transaction."
var ErrErrInsufficientBalanceMsg = "There is not enough money on sender's balance to make requested transaction."
var ErrBalancesLockedMsg = "Sender or receiver balance is locked - no money transfer allowed right now."
var ErrBalancesNotFoundMsg = "Sender or receiver balance not found."

type TransactionController struct {
	G        *echo.Group
	Svc      service.TransactionService
	LoginSvc service.AuthService
}

func (ctr *TransactionController) Init() {
	ctr.G.GET(transactionsEndpoint, ctr.RetriveTransactions)
	ctr.G.POST(transactionsEndpoint, ctr.ExecuteTransaction)
}

// @Summary Executes transaction between two balances.
// @Description Triggers transfer of money from sender balance to receiver balance.
// @Security ApiKeyAuth
// @ID ExecuteTransaction
// @Tags transactions
// @Param user body model.TransactionRequest true "Transaction definifion."
// @Accept  json
// @Produce  json
// @Success 201 {object} model.TransactionResponse
// @Failure 400 {object} model.ErrResponse
// @Failure 401 {object} model.ErrResponse
// @Failure 500 {object} model.ErrResponse
// @Router /api/v1/transactions [post]
func (ctr *TransactionController) ExecuteTransaction(c echo.Context) error {
	log.Infof("POST %s", transactionsEndpoint)

	userID, err := ctr.LoginSvc.GetUserIDFromToken(c)
	if err != nil {
		log.Errorf("error while reading user ID from token; error: ", err)
		return c.JSON(http.StatusUnauthorized, model.NewErrResponse(http.StatusUnauthorized, ErrInvalidTokenMsg))
	}

	t := new(model.TransactionRequest)
	err = c.Bind(t)

	if err != nil {
		log.Errorf("cannot bind TransactionRequest struct with the Request body; error: %w", err)
		return c.JSON(http.StatusBadRequest, model.NewErrResponse(http.StatusBadRequest, ErrCannotParseBodyMsg))
	}

	if ok, err := t.IsValid(); !ok {
		return c.JSON(http.StatusBadRequest, model.NewErrResponse(http.StatusBadRequest, err.Error()))
	}

	transaction := model.Transaction{
		SenderBalanceID:   t.SenderBalanceID,
		ReceiverBalanceID: t.ReceiverBalanceID,
		Amount:            t.Amount,
		Currency:          model.SGD,
	}

	transaction, err = ctr.Svc.Execute(userID, transaction)
	if err != nil {
		log.Errorf("cannot execute transaction; error: %v", err)
		if err == service.ErrBalanceNotFound {
			return c.JSON(http.StatusBadRequest, model.NewErrResponse(http.StatusBadRequest, ErrBalancesNotFoundMsg))
		}
		if err == service.ErrBalancesLocked {
			return c.JSON(http.StatusConflict, model.NewErrResponse(http.StatusConflict, ErrBalancesLockedMsg))
		}
		if err == service.ErrInsufficientBalance {
			return c.JSON(http.StatusBadRequest, model.NewErrResponse(http.StatusBadRequest, ErrErrInsufficientBalanceMsg))
		}
		if err == service.ErrUnauthorizedTransaction {
			return c.JSON(http.StatusUnauthorized, model.NewErrResponse(http.StatusUnauthorized, ErrUnauthorizedTransactionMsg))
		}
		return c.JSON(http.StatusInternalServerError, model.NewErrResponse(http.StatusInternalServerError, ErrInternalServerMsg))
	}

	return c.JSON(http.StatusCreated, model.NewTransactionResponse(transaction))
}

// @Summary Retrives list of transactions.
// @Description Retrives list of transactions for the authenticated user.
// @Security ApiKeyAuth
// @ID RetriveTransactions
// @Tags transactions
// @Produce  json
// @Success 201 {array} model.TransactionResponse
// @Failure 401 {object} model.ErrResponse
// @Failure 500 {object} model.ErrResponse
// @Router /api/v1/transactions [get]
func (ctr *TransactionController) RetriveTransactions(c echo.Context) error {
	log.Infof("GET %s", transactionsEndpoint)

	userID, err := ctr.LoginSvc.GetUserIDFromToken(c)
	if err != nil {
		log.Errorf("error while reading user ID from token; error: ", err)
		return c.JSON(http.StatusUnauthorized, model.NewErrResponse(http.StatusUnauthorized, ErrInvalidTokenMsg))
	}

	transactions, err := ctr.Svc.Retrieve(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.NewErrResponse(http.StatusInternalServerError, ErrInternalServerMsg))
	}

	return c.JSON(http.StatusCreated, model.NewTransactionResponses(transactions))
}
