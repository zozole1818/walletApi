package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"zuzanna.com/walletapi/model"
	"zuzanna.com/walletapi/service"
)

var exampleToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.dyt0CoTl4WoVjAHI9Q_CwSKhl6d_9rhM3NrXuJttkao"

type AuthServiceFake struct {
	db map[string]string
}

func (svc AuthServiceFake) Authenticate(login, password string) (string, error) {
	if svc.db[login] == password {
		return exampleToken, nil
	}
	if login == "err" {
		return "", errors.New("some internal error")
	}
	return "", service.ErrUnauthorized
}

func (svc AuthServiceFake) GetUserIDFromToken(echo.Context) (int, error) {
	return 1, nil
}

func prepareLoginRequest(username, password string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	req.Form = url.Values{}
	req.Form.Add("username", username)
	req.Form.Add("password", password)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	return req
}

type testCase struct {
	username              string
	password              string
	httpCode              int
	expectedTokenResponse model.TokenResponse
	expectedErrResponse   model.ErrResponse
}

func TestLogin(t *testing.T) {
	oneMinuteEarlier := time.Now().Add(-time.Minute * 1)
	cases := []testCase{
		// ok
		{username: "ala11", password: "haslo", httpCode: http.StatusCreated, expectedTokenResponse: model.TokenResponse{Token: exampleToken}, expectedErrResponse: model.ErrResponse{}},
		// ko
		{username: "ala99", password: "haslo", httpCode: http.StatusUnauthorized, expectedTokenResponse: model.TokenResponse{}, expectedErrResponse: model.NewErrResponse(http.StatusUnauthorized, ErrWrongLoginMsg)},
		{username: "", password: "", httpCode: http.StatusUnauthorized, expectedTokenResponse: model.TokenResponse{}, expectedErrResponse: model.NewErrResponse(http.StatusUnauthorized, ErrWrongLoginMsg)},
		{username: "err", password: "haslo", httpCode: http.StatusInternalServerError, expectedTokenResponse: model.TokenResponse{}, expectedErrResponse: model.NewErrResponse(http.StatusInternalServerError, ErrInternalServerMsg)},
	}

	for _, testCase := range cases {
		e := echo.New()
		req := prepareLoginRequest(testCase.username, testCase.password)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		controller := LoginController{
			E: e,
			Svc: AuthServiceFake{
				db: map[string]string{"ala11": "haslo"},
			},
		}
		err := controller.Login(c)
		if err != nil {
			t.Errorf("error was not expected while making a login request: %s", err)
		}
		if rec.Code != testCase.httpCode {
			t.Errorf("http code got: %d; want: %d", rec.Code, testCase.httpCode)
		}
		if rec.Result().Header.Get("content-type") != echo.MIMEApplicationJSONCharsetUTF8 {
			t.Errorf("http header content-type got: %s; want: %s", rec.Result().Header.Get("content-type"), echo.MIMEApplicationJSONCharsetUTF8)
		}
		if testCase.httpCode != http.StatusCreated {
			respBody := model.ErrResponse{}
			err = json.Unmarshal(rec.Body.Bytes(), &respBody)
			if err != nil {
				t.Errorf("error was not expected while making a unmarshal body request: %s", err)
			}
			if respBody.Code != testCase.expectedErrResponse.Code {
				t.Errorf("ErrResponse code got: %d; want: %d", respBody.Code, testCase.expectedErrResponse.Code)
			}
			if respBody.Message != http.StatusText(testCase.expectedErrResponse.Code) {
				t.Errorf("ErrResponse message got: %s; want: %s", respBody.Message, http.StatusText(testCase.expectedErrResponse.Code))
			}
			if respBody.Error != testCase.expectedErrResponse.Error {
				t.Errorf("ErrResponse error got: %s; want: %s", respBody.Error, testCase.expectedErrResponse.Error)
			}
			if !respBody.Date.After(oneMinuteEarlier) {
				t.Errorf("ErrResponse date got: %s; want before: %s", respBody.Date, oneMinuteEarlier)
			}
		} else {
			respBody := model.TokenResponse{}
			err = json.Unmarshal(rec.Body.Bytes(), &respBody)
			if err != nil {
				t.Errorf("error was not expected while making a unmarshal body request: %s", err)
			}
			if respBody.Token != exampleToken {
				t.Errorf("JWT token got: %s; want: %s", respBody.Token, exampleToken)
			}
		}
	}
}
