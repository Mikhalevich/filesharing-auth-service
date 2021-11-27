package db

import "errors"

var (
	ErrNotExist     = errors.New("not exist")
	ErrAlreadyExist = errors.New("already exist")
)
