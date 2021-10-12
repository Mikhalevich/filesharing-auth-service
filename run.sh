#!/usr/bin/env sh

set -xe

export GOOSE_DRIVER=postgres
export GOOSE_DBSTRING=$FS_DB_CONNECTION_STRING

repeater -c=5 goose -dir=db/migrations up

./filesharing-auth-service