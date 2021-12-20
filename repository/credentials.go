package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/labstack/gommon/log"
)

var ErrRecordNotFound = errors.New("record not found")

type Credentials struct {
	ID       int
	Login    string
	Password string
	UserID   int
}

type CredentialsDB struct {
	*pgxpool.Pool
}

func (cred CredentialsDB) Get(login string) (Credentials, error) {
	credentials := Credentials{}
	err := cred.Pool.QueryRow(context.Background(), "SELECT login, password, user_id FROM credentials WHERE login=$1", login).Scan(&credentials.Login, &credentials.Password, &credentials.UserID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return Credentials{}, ErrRecordNotFound
		}
		log.Errorf("error while reading credentials for user with login %s; error %w", login, err)
		return Credentials{}, err
	}
	return credentials, nil
}
