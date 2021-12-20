package service

import (
	"encoding/base64"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/gommon/log"
	"zuzanna.com/walletapi/repository"
)

var ErrUnauthorized = errors.New("login failed")
var ErrTokenInvalid = errors.New("invalid JWT token")
var ErrMissingAuthHeader = errors.New("missing Authorization header")

type AuthService interface {
	// Authenticate authenticate the user using login and password. Returns JWT token and error whether occurs.
	Authenticate(login, password string) (string, error)
	// GetUserID reads userID from JWT token. Returns userID and error whether occurs. When token invalid ErrTokenInvalid returnd.
	GetUserID(header string) (int, error)
	// IsAdmin(jwt string) (bool, error)
}

type AuthServiceImpl struct {
	repository.CredentialsDB
}

var jwtTokenSign []byte

type jwtCustomClaims struct {
	UserID int `json:"user_id"`
	// Admin bool   `json:"admin"`
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

func (svc AuthServiceImpl) Authenticate(login, password string) (string, error) {
	credentials, err := svc.CredentialsDB.Get(login)
	if err != nil {
		if err == repository.ErrRecordNotFound {
			log.Infof("no record found for login %s", login)
			return "", ErrUnauthorized
		}
		log.Errorf("error while retrieving credentials for login %s; error %w", login, err)
		return "", err
	}
	// todo better password hashing perhaps?
	encodedPass, err := base64.StdEncoding.DecodeString(credentials.Password)
	if err != nil {
		log.Errorf("error while encoding password for login %s; error %w", login, err)
		return "", err
	}
	if login != credentials.Login || password != string(encodedPass) {
		return "", ErrUnauthorized
	}
	claims := &jwtCustomClaims{
		credentials.UserID,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour).Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(jwtTokenSign)
	if err != nil {
		log.Errorf("error while signing token; error %w", err)
		return "", err
	}
	return signedToken, nil
}

func (svc AuthServiceImpl) GetUserID(header string) (int, error) {
	if header == "" {
		return 0, ErrMissingAuthHeader
	}
	cleanJWT := strings.Replace(header, "Bearer ", "", 1)
	token, err := jwt.ParseWithClaims(cleanJWT, &jwtCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtTokenSign, nil
	})
	if err != nil {
		log.Errorf("error while parsing JWT token; error %w", err)
		return 0, ErrTokenInvalid
	}
	customClaims, ok := token.Claims.(*jwtCustomClaims)
	if !ok || !token.Valid {
		log.Errorf("error while casting JWT custom claims; error %w", err)
		return 0, ErrTokenInvalid
	}
	return customClaims.UserID, nil
}
