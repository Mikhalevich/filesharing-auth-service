package db

type Email struct {
	Email            string
	Verified         bool
	Primary          bool
	VerificationCode string
}
