package controller

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"zuzanna.com/walletapi/model"
	"zuzanna.com/walletapi/service"
)

type LoginController struct {
	E   *echo.Echo
	Svc service.AuthService
}

func (ctr LoginController) Init() {
	ctr.E.POST(loginEndpoint, ctr.Login)
}

// @Summary Provide your username and password for authentication.
// @Description Login endpoint for getting JWT token.
// @ID Login
// @Tags login
// @Param username formData string true "User's username."
// @Param password formData string true "User's password."
// @Produce  json
// @Success 201 {object} model.TokenResponse
// @Failure 401 {object} model.ErrResponse
// @Failure 500 {object} model.ErrResponse
// @Router /login [post]
// Login returns http response with JWT token required for other endpoints.
func (ctr LoginController) Login(c echo.Context) error {
	log.Infof("POST %s", loginEndpoint)

	username := c.FormValue("username")
	password := c.FormValue("password")

	if username == "" || password == "" {
		return c.JSON(http.StatusUnauthorized, model.NewErrResponse(http.StatusUnauthorized, ErrUnauthorizedMsg))
	}

	token, err := ctr.Svc.Authenticate(username, password)
	if err != nil {
		if errors.Is(err, service.ErrUnauthorized) {
			return c.JSON(http.StatusUnauthorized, model.NewErrResponse(http.StatusUnauthorized, ErrUnauthorizedMsg))
		}
		log.Errorf("error while authenticate user %s; error %w", username, err)
		return c.JSON(http.StatusInternalServerError, model.NewErrResponse(http.StatusInternalServerError, ErrInternalServerMsg))
	}

	return c.JSON(http.StatusCreated, model.TokenResponse{Token: token})
}
