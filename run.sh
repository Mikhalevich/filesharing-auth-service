#!/usr/bin/env sh

repeater goose -dir=db/migrations postgres $FS_DB_CONNECTION_STRING up

./filesharing-auth-service