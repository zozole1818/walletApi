#!/bash

#$ cat ~/.pgpass
## hostname:port:database:username:password
#localhost:5432:*:postgres:admin

psql -h localhost -U postgres -c 'DROP DATABASE wallets;'
