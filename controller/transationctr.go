package controller

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"zuzanna.com/walletapi/model"
	"zuzanna.com/walletapi/service"
)

var ErrBadRequestMsg = "Could not parse body request. Please doble check the JSON."

type TransactionController struct {
	E        *echo.Echo
	Svc      service.TransactionService
	LoginSvc service.AuthService
}

func (ctr *TransactionController) Init() {
	ctr.E.GET(transactionsEndpoint, ctr.RetriveTransactions)
	ctr.E.POST(transactionsEndpoint, ctr.ExecuteTransaction)
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

	h := c.Request().Header.Get("Authorization")
	userID, err := ctr.LoginSvc.GetUserID(h)
	if err != nil {
		if errors.Is(err, service.ErrMissingAuthHeader) {
			log.Errorf("error: %w", err)
			return c.JSON(http.StatusUnauthorized, model.NewErrResponse(http.StatusUnauthorized, ErrMissingAuthHeaderMsg))
		}
		if errors.Is(err, service.ErrTokenInvalid) {
			log.Errorf("invalid token; err %w", err)
			return c.JSON(http.StatusUnauthorized, model.NewErrResponse(http.StatusUnauthorized, ErrInvalidTokenMsg))
		}
	}

	t := new(model.TransactionRequest)
	err = c.Bind(t)
	log.Infof("body from requets: %+v", t)
	if err != nil {
		log.Errorf("cannot bind TransactionRequest struct with the Request body; error: %w", err)
		return c.JSON(http.StatusBadRequest, model.NewErrResponse(http.StatusBadRequest, ErrBadRequestMsg))
	}

	transaction := model.Transaction{
		SenderBalanceID:   t.SenderBalanceID,
		ReceiverBalanceID: t.ReceiverBalanceID,
		Amount:            t.Amount,
		Currency:          model.SGD,
	}

	transaction, err = ctr.Svc.Execute(userID, transaction)
	if err != nil {
		log.Errorf("cannot execute transaction; error: %w", err)
		// ToDo differenciate errors and return meaningful responses
		return c.JSON(http.StatusInternalServerError, model.NewErrResponse(http.StatusInternalServerError, err.Error()))
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

	h := c.Request().Header.Get("Authorization")
	userID, err := ctr.LoginSvc.GetUserID(h)
	if err != nil {
		if errors.Is(err, service.ErrMissingAuthHeader) {
			log.Errorf("error: %w", err)
			return c.JSON(http.StatusUnauthorized, model.NewErrResponse(http.StatusUnauthorized, ErrMissingAuthHeaderMsg))
		}
		if errors.Is(err, service.ErrTokenInvalid) {
			log.Errorf("invalid token; err %w", err)
			return c.JSON(http.StatusUnauthorized, model.NewErrResponse(http.StatusUnauthorized, ErrInvalidTokenMsg))
		}
	}

	transactions, err := ctr.Svc.Retrieve(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.NewErrResponse(http.StatusInternalServerError, ErrInternalServerMsg))
	}

	return c.JSON(http.StatusCreated, model.NewTransactionResponses(transactions))
}
