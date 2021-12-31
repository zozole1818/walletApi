package service

import (
	"errors"
	"fmt"
	"testing"

	"github.com/golang-jwt/jwt"
	"zuzanna.com/walletapi/model"
	"zuzanna.com/walletapi/repository"
)

var errExpected = errors.New("some database error")

type CredentialsRepoFake struct {
	db map[string]model.Credentials
}

func newCredentialsRepoFake() CredentialsRepoFake {
	return CredentialsRepoFake{
		db: map[string]model.Credentials{
			"ala11": {ID: 1, Login: "ala11", Password: "aGFzbG8=", UserID: 1},
			"ola22": {ID: 2, Login: "ola22", Password: "MXFhelhTV0A=", UserID: 2},
		},
	}
}

func (r CredentialsRepoFake) Get(login string) (model.Credentials, error) {
	cred, ok := r.db[login]
	if ok {
		return cred, nil
	} else if login == "err" {
		return model.Credentials{}, errExpected
	}
	return model.Credentials{}, repository.ErrRecordNotFound
}

type testCase struct {
	username    string
	password    string
	expectedErr error
}

func TestAuthenticate(t *testing.T) {

	cases := []testCase{
		{"ala11", "haslo", nil},
		{"ola22", "1qazXSW@", nil},
		{"nofound", "--", ErrUnauthorized},
		{"ala11", "wrongPass", ErrUnauthorized},
		{"err", "haslo", errExpected},
	}

	authSvc := AuthServiceImpl{
		newCredentialsRepoFake(),
	}

	for _, testCase := range cases {
		tokenStr, err := authSvc.Authenticate(testCase.username, testCase.password)
		if testCase.expectedErr == nil {
			token, err := jwt.ParseWithClaims(tokenStr, &JwtCustomClaims{}, func(t *jwt.Token) (interface{}, error) {
				if t.Method.Alg() != "HS256" {
					return nil, fmt.Errorf("unexpected jwt signing method=%v", t.Header["alg"])
				}
				return jwtTokenSign, nil
			})
			if err != nil {
				t.Errorf("error was not expected while parsing JWT token: %s", err)
			}
			if !token.Valid {
				t.Errorf("token invalid got: %s (%+v); want: valid token", tokenStr, token)
			}
		} else {
			if err != testCase.expectedErr {
				t.Errorf("unexpected err got: %s; want: %s", err, testCase.expectedErr)
			}
		}
	}

}
