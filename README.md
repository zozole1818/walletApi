# walletAPI

## Introduction
walletAPI provides a fund transfer solution for users.
This is REST API written in Go backed by PostreSQL.

Project is POC and the **key features** are:
* transfer of money between users
* information about user's balance
* information about user's transactions

### Starting point
* user registration functionality skipped - provisioning of users/credentials via script - `/scripts/populate_db.sh` - executed from within `/devops/web/entrypoint.sh`
* bash script `/devops/web/wait-for-it.sh` taken from [wait-for-it](https://github.com/vishnubob/wait-for-it)

## Assumptions/Limitations
* user can only have zero or positive balance (no debet)
* only one type of currency is supported - that is SGD, all transactions and balance are in SGD
* amount of money send in TransferRequest is rounded down to 2 decimal places
* docker-compose that starts walletApi and postgres db **DOES NOT** mount any files - that's why if you kill the docker-compose's dockers and start again, fresh installation will be available
* server is using http

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
All available endpoints can be access via swagger UI on address `http://localhost:8000/swagger/index.html`.
You should be able to see:
![swagger-api](/pictures/swagger-api.PNG)

**NOTE:** Please first access /login endpoint to get JWT token. To Access other endpoints Click "Authorize" button (right upper corner) and paste there "Bearer \<your-token\>".
**Do not forget add _Bearer_ prefix!!!** (swagger 2.0 used here do not support jwt token auth)

Users provisioned during startup:
| username    | password |
| ----------- | -------- |
| test11      | haslo    |
| zazu18      | haslo    |
| johndoe11   | haslo    |
| jimsmith44  | haslo    |

**NOTE:** When you set Bearer token then you're able to see balances and transactions for **the user that was authorized**. To see balances and transactions of different user you must login with different credentials.

#### Prometheus metric endpoint
Prometheus metric endpoint is not visible on swagger UI. To see metrics please go to `http://localhost:8000/metrics`.

### Future enhancement?
This is only POC created really fast. Many things can be done in a different way or added, e.g.:
* credentials for users could be in some LDAP? for sure could have better coding in db
* depends on how the api will be used endpoints could change to include userID in endpoints - for now UserID is retrieved from JWT token
* authentication should be stronger not just JWT token based on username/password
* all structs/objects can and should be more detailed, e.g. transaction request should have title, description...
* retriving transaction for now is only by userID - some more restrictions and pagination would be good
* more tests added
* must have https
* fully functional solution should have logging capture, e.g. Elasticsearch -> Kibana
* when comes to logging audit logging would be nice
* endpoint for Prometheus metrics available - Prometheus scraping the endpoint would be a good idea :) + Grafana for nice graphs if needed
