package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger" // echo-swagger middleware
	"zuzanna.com/walletapi/controller"
	_ "zuzanna.com/walletapi/docs"
	"zuzanna.com/walletapi/repository"
	"zuzanna.com/walletapi/service"
)

// @title walletApi by Zuzanna
// @version 1.0
// @description This is POC for walletApi. Is allows you to send money between users already registered in the system.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8000
// @query.collection.format multi
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	// toDo os.Getenv("DATABASE_URL")
	pool, err := pgxpool.Connect(context.Background(), EnvWithDefault("DATABASE_URL", "postgres://postgres:admin@localhost:5432/mobile_wallet"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	e := echo.New()
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	loginSvc := service.AuthServiceImpl{
		CredentialsDB: repository.CredentialsDB{Pool: pool},
	}
	postgreBalanceRepo := repository.NewPostgreBalanceRepo(pool)
	loginController := controller.LoginController{
		E:   e,
		Svc: loginSvc,
	}
	balanceController := controller.BalanceController{
		E:          e,
		BalanceSvc: service.NewBalanceService(postgreBalanceRepo),
		LoginSvc:   loginSvc,
	}
	transactionController := controller.TransactionController{
		E:        e,
		LoginSvc: loginSvc,
		Svc:      service.NewTransactionService(postgreBalanceRepo),
	}

	loginController.Init()
	balanceController.Init()
	transactionController.Init()

	e.Logger.Fatal(e.Start(":8000"))
}

func EnvWithDefault(n string, d string) string {
	env, ok := os.LookupEnv(n)
	if ok {
		return env
	}
	return d
}
