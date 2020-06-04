package db

type User struct {
	ID     int64
	Name   string
	Pwd    string
	Emails []Email
}
