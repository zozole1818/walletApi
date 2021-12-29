#!/usr/bin/env bash

#$ cat ~/.pgpass
## hostname:port:database:username:password
#db:5432:*:postgres:admin

psql -h db -U postgres -c 'CREATE DATABASE wallets;'
psql -h db -U postgres -d wallets -c 'CREATE TABLE "user"(ID SERIAL PRIMARY KEY NOT NULL, first_name VARCHAR(10) NOT NULL, last_name VARCHAR(10) NOT NULL, age INT NOT NULL);'

psql -h db -U postgres -d wallets -c 'CREATE TABLE "credentials"(ID SERIAL PRIMARY KEY NOT NULL, login VARCHAR(20) NOT NULL UNIQUE, password VARCHAR(30) NOT NULL, user_ID INT references "user"(ID) NOT NULL);'
psql -h db -U postgres -d wallets -c 'CREATE TABLE "balance"(ID SERIAL PRIMARY KEY NOT NULL, currency VARCHAR(3) NOT NULL, balance NUMERIC(12, 2) NOT NULL, locked BOOLEAN DEFAULT false, user_ID INT references "user"(ID) NOT NULL);'
psql -h db -U postgres -d wallets -c 'CREATE TABLE "transaction"(ID SERIAL PRIMARY KEY NOT NULL, sender_ID INT NOT NULL, receiver_ID INT NOT NULL, currency VARCHAR(3) NOT NULL, amount NUMERIC(12, 2) NOT NULL, date TIMESTAMP NOT NULL);'
psql -h db -U postgres -d wallets -c 'CREATE TABLE "balance_transaction"(balance_ID INT references "balance"(ID), transaction_ID INT references "transaction"(ID));'

psql -h db -U postgres -d wallets -c 'INSERT INTO "user"(first_name, last_name, age) VALUES('"'"'Alice'"'"', '"'"'Cruz'"'"', 25);'
psql -h db -U postgres -d wallets -c 'INSERT INTO "credentials"(login, password, user_ID) VALUES('"'"'test11'"'"', '"'"'aGFzbG8='"'"', 1);'
psql -h db -U postgres -d wallets -c 'INSERT INTO "balance"(currency, balance, user_ID) VALUES('"'"'SGD'"'"', 1000, 1);'

psql -h db -U postgres -d wallets -c 'INSERT INTO "user"(first_name, last_name, age) VALUES('"'"'Zuzanna'"'"', '"'"'Zazu'"'"', 18);'
psql -h db -U postgres -d wallets -c 'INSERT INTO "credentials"(login, password, user_ID) VALUES('"'"'zazu18'"'"', '"'"'aGFzbG8='"'"', 2);'
psql -h db -U postgres -d wallets -c 'INSERT INTO "balance"(currency, balance, user_ID) VALUES('"'"'SGD'"'"', 100, 2);'

psql -h db -U postgres -d wallets -c 'INSERT INTO "user"(first_name, last_name, age) VALUES('"'"'John'"'"', '"'"'Doe'"'"', 40);'
psql -h db -U postgres -d wallets -c 'INSERT INTO "credentials"(login, password, user_ID) VALUES('"'"'johndoe11'"'"', '"'"'aGFzbG8='"'"', 3);'
psql -h db -U postgres -d wallets -c 'INSERT INTO "balance"(currency, balance, user_ID) VALUES('"'"'SGD'"'"', 20000, 3);'

psql -h db -U postgres -d wallets -c 'INSERT INTO "user"(first_name, last_name, age) VALUES('"'"'Jim'"'"', '"'"'Smith'"'"', 50);'
psql -h db -U postgres -d wallets -c 'INSERT INTO "credentials"(login, password, user_ID) VALUES('"'"'jimsmith44'"'"', '"'"'aGFzbG8='"'"', 4);'
psql -h db -U postgres -d wallets -c 'INSERT INTO "balance"(currency, balance, user_ID) VALUES('"'"'SGD'"'"', 10, 4);'

psql -h db -U postgres -d wallets -c 'INSERT INTO "transaction"(sender_ID, receiver_ID, currency, amount, date) VALUES(1, 2, '"'"'SGD'"'"', 100, now());'
psql -h db -U postgres -d wallets -c 'INSERT INTO "balance_transaction"(balance_ID, transaction_ID) VALUES(1, 1);'
psql -h db -U postgres -d wallets -c 'INSERT INTO "balance_transaction"(balance_ID, transaction_ID) VALUES(2, 1);'
