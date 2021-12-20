package controller

import (
	"errors"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"zuzanna.com/walletapi/model"
	"zuzanna.com/walletapi/service"
)

// new file?
var ErrBadUserIDMsg = "Proper user ID required."
var ErrInternalServerMsg = "Server error, please try again."
var ErrUnauthorizedMsg = "Login failed. Please double check username and password."
var ErrInvalidTokenMsg = "Invalid token."
var ErrMissingAuthHeaderMsg = "Authorization header required."

type BalanceController struct {
	E          *echo.Echo
	BalanceSvc service.BalanceService
	LoginSvc   service.AuthService
}

func (ctr BalanceController) Init() {
	ctr.E.GET(balancesEndpoint, ctr.GetBalances)
}

// @Summary Retrieves list of balances for authenticated user.
// @Description Retrieves list of balances for authenticated user.
// @Security ApiKeyAuth
// @ID GetBalances
// @Tags balances
// @Produce  json
// @Success 200 {array} model.BalanceResponse
// @Failure 401 {object} model.ErrResponse
// @Failure 500 {object} model.ErrResponse
// @Router /api/v1/balances [get]
func (ctr BalanceController) GetBalances(c echo.Context) error {
	log.Infof("GET %s", replaceID(balancesEndpoint, ""))

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

	balances, err := ctr.BalanceSvc.GetByUserID(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.NewErrResponse(http.StatusInternalServerError, ErrInternalServerMsg))
	}
	return c.JSON(http.StatusOK, model.NewBalanceResponses(balances))
}

func replaceID(s string, id string) string {
	return strings.Replace(s, ":id", id, 1)
}
