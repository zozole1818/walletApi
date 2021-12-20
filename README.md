# walletAPI

## Introduction
walletAPI provides a fund transfer solution for users.
That is REST API written in Go backed by PostreSQL.

Project is POC and the **key features** are:
* transfer of money between users
* information about user balance
* information about user transactions

### Starting point
* user registration and login functionality skipped - provision of users/credentials via :!script:!

## Assumptions/Limitations
* user can only has zero or positive balance (no debet)
* only one type of currency is supported - that is SGD, all transactions and balance is in SGD

## Get started
### Instalation
```bash
$ git clone https://github.com/zozole1818/walletApi.git
$ cd walletApi/devops
$ docker-compose up # or sudo docker-compose up
```

If you can see in the terminal:
```bash
web_1  |    ____    __
web_1  |   / __/___/ /  ___
web_1  |  / _// __/ _ \/ _ \
web_1  | /___/\__/_//_/\___/ v4.0.0
web_1  | High performance, minimalist Go web framework
web_1  | https://echo.labstack.com
web_1  | ____________________________________O/_______
web_1  |                                     O\
web_1  | ⇨ http server started on [::]:8000

```

... then you're ready to go.

### Endpoints
All available endpoints are can be access via swagger UI on address `http://localhost:8000/swagger/index.html`.
