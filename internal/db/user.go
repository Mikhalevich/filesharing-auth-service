package db

import (
	"database/sql"
)

type User struct {
	ID     int64
	Name   string
	Pwd    sql.NullString
	Public bool
	Emails []Email
}
