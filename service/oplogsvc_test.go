package service

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"zuzanna.com/walletapi/model"
)

func TestLogSkipper(t *testing.T) {

	testCases := []struct {
		uri        string
		shouldSkip bool
	}{
		{
			uri:        "/login",
			shouldSkip: false,
		},
		{
			uri:        "/metrics",
			shouldSkip: true,
		},
		{
			uri:        "/swagger",
			shouldSkip: true,
		},
		{
			uri:        "/somefake",
			shouldSkip: false,
		},
		{
			uri:        "/api/v1/balances",
			shouldSkip: false,
		},
	}

	svc := NewJSONOperationalLogService()

	for _, testCase := range testCases {
		req := httptest.NewRequest(http.MethodPost, testCase.uri, nil)
		rec := httptest.NewRecorder()
		e := echo.New()
		c := e.NewContext(req, rec)
		skip := svc.LogSkipper(c)
		if skip != testCase.shouldSkip {
			t.Errorf("skip for request %s got: %t ; want: %t ", testCase.uri, skip, testCase.shouldSkip)
		}
	}

}

func getTime(timeStr string) time.Time {
	t, _ := time.Parse(time.RFC3339Nano, timeStr)
	return t
}

func TestCreateLog(t *testing.T) {
	testCases := []struct {
		contentType       string
		username          string
		password          string
		requestBody       []byte
		responseBody      []byte
		expectedOpLog     model.OperationalLog
		expectedOpLogJSON string
		// set for JSON comparision
		staticTime time.Time
	}{
		{
			contentType:  echo.MIMEApplicationForm,
			username:     "test11",
			password:     "alapass",
			requestBody:  nil,
			responseBody: []byte(`{"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJleHAiOjE2NDE4OTIwMDh9.0zVjL_popqjaFo7MN6R-7CYkWfWMJ5KEgSoMl5prfm8"}`),
			expectedOpLog: newOpLog("example.com", "/login", http.MethodPost, 0, map[string]interface{}{
				"username": []string{"test11"},
				"password": "***",
			}, nil, []byte(`{"token":"***"}`), 201, ""),
			expectedOpLogJSON: `{"type":"OPERATIONAL","time":"2022-01-11T14:09:38.1374053+01:00","level":"INFO","protocol":"http","host":"example.com","path":"/login","method":"POST","userID":0,"request":{"form":{"password":"***","username":["test11"]}},"response":{"body":{"token":"***"},"code":201},"err":""}`,
			staticTime:        getTime("2022-01-11T14:09:38.1374053+01:00"),
		},
		{
			contentType:       echo.MIMEApplicationJSONCharsetUTF8,
			requestBody:       nil,
			responseBody:      []byte(`[{"id":1,"currency":"SGD","balance":898.36}]`),
			expectedOpLog:     newOpLog("example.com", "/api/v1/balances", http.MethodGet, 1, nil, nil, []byte(`[{"id":1,"currency":"SGD","balance":898.36}]`), 200, ""),
			expectedOpLogJSON: `{"type":"OPERATIONAL","time":"2022-01-11T14:09:38.1374053+01:00","level":"INFO","protocol":"http","host":"example.com","path":"/api/v1/balances","method":"GET","userID":1,"request":{},"response":{"body":[{"id":1,"currency":"SGD","balance":898.36}],"code":200},"err":""}`,
			staticTime:        getTime("2022-01-11T14:09:38.1374053+01:00"),
		},
		{
			contentType:       echo.MIMEApplicationJSON,
			requestBody:       nil,
			responseBody:      []byte(`[{"id":1,"currency":"SGD","balance":898.36}]`),
			expectedOpLog:     newOpLog("example.com", "/api/v1/balances", http.MethodGet, 1, nil, nil, []byte(`[{"id":1,"currency":"SGD","balance":898.36}]`), 200, ""),
			expectedOpLogJSON: `{"type":"OPERATIONAL","time":"2022-01-11T14:09:38.1374053+01:00","level":"INFO","protocol":"http","host":"example.com","path":"/api/v1/balances","method":"GET","userID":1,"request":{},"response":{"body":[{"id":1,"currency":"SGD","balance":898.36}],"code":200},"err":""}`,
			staticTime:        getTime("2022-01-11T14:09:38.1374053+01:00"),
		},
		{
			contentType:       echo.MIMEApplicationJSONCharsetUTF8,
			requestBody:       []byte(`{"amount":22,"receiverBalanceId":2,"senderBalanceId":1}`),
			responseBody:      []byte(`{"id":1,"amount":20,"receiverBalanceId":2,"senderBalanceId":1,"currency":"SGD","date":"2022-01-11T09:13:45.8611076+01:00"}`),
			expectedOpLog:     newOpLog("example.com", "/api/v1/transactions", http.MethodPost, 1, nil, []byte(`{"amount":22,"receiverBalanceId":2,"senderBalanceId":1}`), []byte(`{"id":1,"amount":20,"receiverBalanceId":2,"senderBalanceId":1,"currency":"SGD","date":"2022-01-11T09:13:45.8611076+01:00"}`), 201, ""),
			expectedOpLogJSON: `{"type":"OPERATIONAL","time":"2022-01-11T14:09:38.1374053+01:00","level":"INFO","protocol":"http","host":"example.com","path":"/api/v1/transactions","method":"POST","userID":1,"request":{"body":{"amount":22,"receiverBalanceId":2,"senderBalanceId":1}},"response":{"body":{"id":1,"amount":20,"receiverBalanceId":2,"senderBalanceId":1,"currency":"SGD","date":"2022-01-11T09:13:45.8611076+01:00"},"code":201},"err":""}`,
			staticTime:        getTime("2022-01-11T14:09:38.1374053+01:00"),
		},
		{
			contentType:       echo.MIMETextPlain,
			requestBody:       []byte(`{"amount":22,"receiverBalanceId":2,"senderBalanceId":`),
			responseBody:      []byte(`{"code":400,"message":"Bad Request","error":"Could not parse body request. Please doble check the JSON.","date":"2022-01-12T08:50:28.781539+01:00"}`),
			expectedOpLog:     newOpLog("example.com", "/api/v1/transactions", http.MethodPost, 1, nil, nil, []byte(`{"code":400,"message":"Bad Request","error":"Could not parse body request. Please doble check the JSON.","date":"2022-01-12T08:50:28.781539+01:00"}`), 400, "error while marshal body to JSON"),
			expectedOpLogJSON: `{"type":"OPERATIONAL","time":"2022-01-11T14:09:38.1374053+01:00","level":"INFO","protocol":"http","host":"example.com","path":"/api/v1/transactions","method":"POST","userID":1,"request":{},"response":{"body":{"code":400,"message":"Bad Request","error":"Could not parse body request. Please doble check the JSON.","date":"2022-01-12T08:50:28.781539+01:00"},"code":400},"err":"error while marshal body to JSON"}`,
			staticTime:        getTime("2022-01-11T14:09:38.1374053+01:00"),
		},
		{
			contentType:       echo.MIMEApplicationJSONCharsetUTF8,
			requestBody:       []byte(`{"amount":22,"receiverBalanceId":2,"senderBalanceId":`),
			responseBody:      []byte(`{"code":400,"message":"Bad Request","error":"Could not parse body request. Please doble check the JSON.","date":"2022-01-12T08:50:28.781539+01:00"}`),
			expectedOpLog:     newOpLog("example.com", "/api/v1/transactions", http.MethodPost, 1, nil, nil, []byte(`{"code":400,"message":"Bad Request","error":"Could not parse body request. Please doble check the JSON.","date":"2022-01-12T08:50:28.781539+01:00"}`), 400, "error while marshal body to JSON"),
			expectedOpLogJSON: `{"type":"OPERATIONAL","time":"2022-01-11T14:09:38.1374053+01:00","level":"INFO","protocol":"http","host":"example.com","path":"/api/v1/transactions","method":"POST","userID":1,"request":{},"response":{"body":{"code":400,"message":"Bad Request","error":"Could not parse body request. Please doble check the JSON.","date":"2022-01-12T08:50:28.781539+01:00"},"code":400},"err":"error while marshal body to JSON"}`,
			staticTime:        getTime("2022-01-11T14:09:38.1374053+01:00"),
		},
	}

	for _, testCase := range testCases {
		req := httptest.NewRequest(testCase.expectedOpLog.Method, testCase.expectedOpLog.Path, nil)
		req.Header.Set(echo.HeaderContentType, testCase.contentType)
		if testCase.username != "" && testCase.password != "" {
			req.Form = url.Values{}
			req.Form.Add("username", testCase.username)
			req.Form.Add("password", testCase.password)
		}
		rec := httptest.NewRecorder()
		e := echo.New()
		c := e.NewContext(req, rec)
		c.Response().Status = testCase.expectedOpLog.Response.Code
		c.Set("user", &jwt.Token{
			Claims: &JwtCustomClaims{
				UserID: testCase.expectedOpLog.UserID,
			},
		})

		var got model.OperationalLog
		if testCase.contentType == echo.MIMEApplicationForm {
			got = createLog(c, nil, nil)
		}
		got = createLog(c, testCase.requestBody, testCase.responseBody)

		oneMinuteEarlier := time.Now().Add(-1 * time.Minute)
		if !got.Time.After(oneMinuteEarlier) {
			t.Errorf("time for operational log incorrect got: %s ; want after: %s", got.Time, oneMinuteEarlier)
		}
		testCase.expectedOpLog.Time = got.Time

		fmt.Println(string(got.Request.Body))
		fmt.Println(string(testCase.expectedOpLog.Request.Body))
		fmt.Println(string(got.Response.Body))
		fmt.Println(string(testCase.expectedOpLog.Response.Body))
		assert.Equal(t, testCase.expectedOpLog.Type, got.Type, "comparing operational log type")
		assert.Equal(t, testCase.expectedOpLog.Time, got.Time, "comparing operational log time")
		assert.Equal(t, testCase.expectedOpLog.Level, got.Level, "comparing operational log level")
		assert.Equal(t, testCase.expectedOpLog.Protocol, got.Protocol, "comparing operational log protocol")
		assert.Equal(t, testCase.expectedOpLog.Host, got.Host, "comparing operational log host")
		assert.Equal(t, testCase.expectedOpLog.Path, got.Path, "comparing operational log path")
		assert.Equal(t, testCase.expectedOpLog.Method, got.Method, "comparing operational log method")
		assert.Equal(t, testCase.expectedOpLog.UserID, got.UserID, "comparing operational log userID")
		assert.Equal(t, testCase.expectedOpLog.Request.Body, got.Request.Body, "comparing operational log request body")
		assert.Equal(t, testCase.expectedOpLog.Request.Form, got.Request.Form, "comparing operational log request form")
		assert.Equal(t, testCase.expectedOpLog.Response.Body, got.Response.Body, "comparing operational log response body")
		assert.Equal(t, testCase.expectedOpLog.Response.Code, got.Response.Code, "comparing operational log response code")

		got.Time = testCase.staticTime
		assert.Equal(t, testCase.expectedOpLogJSON, NewJSONOperationalLogService().ToJSON(got), "comparing JSON operational log")
		fmt.Println(NewJSONOperationalLogService().ToJSON(got))
	}
}

func newOpLog(host, path, method string, userID int, form map[string]interface{}, reqBody, respBody []byte, respCode int, err string) model.OperationalLog {
	return model.OperationalLog{
		Type:     model.Operaional,
		Time:     time.Now(),
		Level:    model.Info,
		Protocol: model.Http,
		Host:     host,
		Path:     path,
		Method:   method,
		UserID:   userID,
		Request: model.Request{
			Body: reqBody,
			Form: form,
		},
		Response: model.Response{
			Body: respBody,
			Code: respCode,
		},
		Err: err,
	}
}

func TestGetUserIDFromTokenFail(t *testing.T) {
	e := echo.New()
	c := e.NewContext(nil, nil)

	userID, err := getUserIDFromToken(c)
	assert.Equal(t, 0, userID, "comparing userID from context")
	assert.Equal(t, ErrNoUSerInContext, err, "comparing err")
}
