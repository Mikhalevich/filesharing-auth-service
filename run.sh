#!/usr/bin/env sh

sleep 2

goose -dir=db/migrations postgres "user=postgres password=123456 dbname=filesharing host=dbpg port=5432 sslmode=disable" up

./filesharing-auth-service