package model

import (
	"encoding/json"
	"time"
)

type LogType string

const (
	Operaional LogType = "OPERATIONAL"
)

type LogLevel string

const (
	Info LogLevel = "INFO"
)

type Protocol string

const (
	Http  Protocol = "http"
	Https Protocol = "https"
)

type OperationalLog struct {
	Type     LogType   `json:"type"`
	Time     time.Time `json:"time"`
	Level    LogLevel  `json:"level"`
	Protocol Protocol  `json:"protocol"`
	Host     string    `json:"host"`
	Path     string    `json:"path"`
	Method   string    `json:"method"`
	UserID   int       `json:"userID"`
	Request  Request   `json:"request"`
	Response Response  `json:"response"`
	Err      string    `json:"err"`
}

type Request struct {
	Body json.RawMessage        `json:"body,omitempty"`
	Form map[string]interface{} `json:"form,omitempty"`
}

type Response struct {
	Body json.RawMessage `json:"body"`
	Code int             `json:"code"`
}
