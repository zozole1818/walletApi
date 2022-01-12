package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"zuzanna.com/walletapi/model"
)

var ErrNoUSerInContext = errors.New("could not retrieve userID from context")

type OperationalLogService interface {
	CreateLog(c echo.Context, reqBody, resBody []byte)
	LogSkipper(c echo.Context) bool
}

type JSONOperationalLogService struct {
}

func NewJSONOperationalLogService() JSONOperationalLogService {
	return JSONOperationalLogService{}
}

func (svc JSONOperationalLogService) ToJSON(opLog model.OperationalLog) string {
	b, err := json.Marshal(opLog)
	if err != nil {
		log.Errorf("could not marshal operational log (%+v) to JSON, err: %s", opLog, err)
		return ""
	}
	return string(b)
}

func (svc JSONOperationalLogService) LogSkipper(c echo.Context) bool {
	uri := c.Request().RequestURI
	if strings.HasPrefix(uri, "/metrics") || strings.HasPrefix(uri, "/swagger") {
		return true
	}
	return false
}

func (svc JSONOperationalLogService) CreateLog(c echo.Context, reqBody, resBody []byte) {
	fmt.Println(svc.ToJSON(createLog(c, reqBody, resBody)))
}

func createLog(c echo.Context, reqBody, resBody []byte) model.OperationalLog {
	opLog := initOperationalLog()
	req := c.Request()
	resp := c.Response()
	opLog.Host = req.Host
	opLog.Path = req.RequestURI
	opLog.Method = req.Method
	if req.Header.Get("content-type") == echo.MIMEApplicationForm {
		opLog.Request = model.Request{Form: byteFormToMap(req)}
	} else {
		body, err := createRawMessageWithCheck(reqBody)
		opLog.Request = model.Request{Body: body}
		opLog.Err = wrapErr(opLog.Err, err)
	}
	if req.RequestURI == "/login" {
		body, err := hideSensitiveData(resBody)
		opLog.Response = model.Response{Body: body, Code: resp.Status}
		opLog.Err = wrapErr(opLog.Err, err)
	} else {
		opLog.Response = model.Response{Body: createRawMessage(resBody), Code: resp.Status}
		id, err := getUserIDFromToken(c)
		opLog.UserID = id
		opLog.Err = wrapErr(opLog.Err, err)
	}
	return opLog
}

func wrapErr(initial string, wrapped error) string {
	if wrapped == nil {
		return initial
	}
	if initial == "" {
		return wrapped.Error()
	}
	return initial + "; " + wrapped.Error()
}

func initOperationalLog() model.OperationalLog {
	return model.OperationalLog{Type: model.Operaional, Level: model.Info, Time: time.Now(), Protocol: model.Http}
}

func getUserIDFromToken(c echo.Context) (int, error) {
	contextUser := c.Get("user")
	if contextUser == nil {
		log.Infof("could not retrieve user from context for operational log")
		return 0, ErrNoUSerInContext
	}
	userToken := contextUser.(*jwt.Token)
	claims, ok := userToken.Claims.(*JwtCustomClaims)
	if !ok {
		log.Error("could not retrieve userID from context for operational log")
		return 0, ErrNoUSerInContext
	}
	return claims.UserID, nil
}

func createRawMessage(body []byte) json.RawMessage {
	if len(body) == 0 {
		return nil
	}
	return json.RawMessage(strings.TrimSpace(string(body)))
}

func createRawMessageWithCheck(body []byte) (json.RawMessage, error) {
	if len(body) == 0 {
		return nil, nil
	}
	foundJSON := json.RawMessage(strings.TrimSpace(string(body)))
	b, err := json.Marshal(foundJSON)
	if err != nil {
		log.Errorf("could not marshal body (%s) to map, err: %s", string(body), err)
		return nil, errors.New("error while marshal body to JSON")
	}
	return b, nil
}

func byteFormToMap(req *http.Request) map[string]interface{} {
	m := make(map[string]interface{})
	req.ParseForm()
	values := req.Form
	for k, v := range values {
		// hide sensitive data
		if k == "password" {
			m[k] = "***"
			continue
		}
		m[k] = v
	}
	return m
}

func hideSensitiveData(body []byte) ([]byte, error) {
	if len(body) == 0 {
		return nil, nil
	}
	m := make(map[string]interface{})
	err := json.Unmarshal(body, &m)
	if err != nil {
		log.Errorf("could not marshal body (%s) to map, err: %s", string(body), err)
		return nil, errors.New("error while unmarshal body to map")
	}
	// hide sensitive data
	if _, ok := m["token"]; ok {
		m["token"] = "***"
	}
	b, err := json.Marshal(m)
	if err != nil {
		log.Errorf("could not marshal data (%+v) to JSON, err: %s", m, err)
		return nil, errors.New("error while marshal body to JSON")
	}
	return b, nil
}
