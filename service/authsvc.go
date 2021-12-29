package service

import (
	"encoding/base64"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"zuzanna.com/walletapi/model"
	"zuzanna.com/walletapi/repository"
)

var ErrUnauthorized = errors.New("login failed")

type AuthService interface {
	Authenticate(login, password string) (string, error)
	GetUserIDFromToken(echo.Context) (int, error)
}

type AuthServiceImpl struct {
	repository.CredentialsRepo
}

var jwtTokenSign []byte

type JwtCustomClaims struct {
	UserID int `json:"user_id"`
	jwt.StandardClaims
}

func init() {
	env, ok := os.LookupEnv("JWT_SIGN")
	if ok {
		jwtTokenSign = []byte(env)
	} else {
		jwtTokenSign = []byte("mySecret")
	}
}

func GetJwtTokenSign() []byte {
	return jwtTokenSign
}

func JWTErrorHandlerWithContext(err error, c echo.Context) error {
	if err == middleware.ErrJWTMissing {
		return c.JSON(http.StatusBadRequest, model.NewErrResponse(http.StatusBadRequest, "Missing or malformed JWT token."))
	}
	if err == middleware.ErrJWTInvalid {
		return c.JSON(http.StatusUnauthorized, model.NewErrResponse(http.StatusUnauthorized, "Invalid or expired JWT token."))
	}
	return c.JSON(http.StatusUnauthorized, model.NewErrResponse(http.StatusUnauthorized, "Invalid JWT token."))
}

func (svc AuthServiceImpl) Authenticate(login, password string) (string, error) {
	credentials, err := svc.CredentialsRepo.Get(login)
	if err != nil {
		if err == repository.ErrRecordNotFound {
			log.Infof("no record found for login %s", login)
			return "", ErrUnauthorized
		}
		log.Errorf("error while retrieving credentials for login %s; error %v", login, err)
		return "", err
	}

	encodedPass, err := base64.StdEncoding.DecodeString(credentials.Password)
	if err != nil {
		log.Errorf("error while encoding password for login %s; error %v", login, err)
		return "", err
	}
	if login != credentials.Login || password != string(encodedPass) {
		return "", ErrUnauthorized
	}
	claims := &JwtCustomClaims{
		credentials.UserID,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour).Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(jwtTokenSign)
	if err != nil {
		log.Errorf("error while signing token; error %v", err)
		return "", err
	}
	return signedToken, nil
}

func (svc AuthServiceImpl) GetUserIDFromToken(c echo.Context) (int, error) {
	userToken := c.Get("user").(*jwt.Token)
	claims, ok := userToken.Claims.(*JwtCustomClaims)
	if !ok {
		return 0, errors.New("error while reading JWT custom claims")
	}
	return claims.UserID, nil
}
