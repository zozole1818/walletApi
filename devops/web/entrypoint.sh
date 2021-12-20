#!/usr/bin/env bash

cp /app/devops/web/.pgpass ~/.pgpass
chmod 600 ~/.pgpass

/app/devops/web/wait-for-it.sh db:5432 --strict --timeout=60 -- /app/scripts/populate_db.sh

/app/cmd/walletApi
