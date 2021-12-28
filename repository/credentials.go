package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/labstack/gommon/log"
	"zuzanna.com/walletapi/model"
)

var ErrRecordNotFound = errors.New("record not found")

type CredentialsRepo interface {
	Get(login string) (model.Credentials, error)
}

type PostgreCredentialsRepo struct {
	*pgxpool.Pool
}

func (cred PostgreCredentialsRepo) Get(login string) (model.Credentials, error) {
	credentials := model.Credentials{}
	err := cred.Pool.QueryRow(context.Background(), "SELECT login, password, user_id FROM credentials WHERE login=$1", login).Scan(&credentials.Login, &credentials.Password, &credentials.UserID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return model.Credentials{}, ErrRecordNotFound
		}
		log.Errorf("error while reading credentials for user with login %s; error %v", login, err)
		return model.Credentials{}, err
	}
	return credentials, nil
}
