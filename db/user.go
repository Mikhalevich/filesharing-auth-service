package db

type User struct {
	ID     int64
	Name   string
	Pwd    string
	Public bool
	Emails []Email
}
