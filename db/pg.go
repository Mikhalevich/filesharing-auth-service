package db

import (
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

type Postgres struct {
	db *sql.DB
}

func NewPostgres(connectionStr string) (*Postgres, error) {
	pgDB, err := sql.Open("postgres", connectionStr)
	if err != nil {
		return nil, err
	}

	return &Postgres{
		db: pgDB,
	}, nil
}

func (p *Postgres) Close() error {
	return p.db.Close()
}

func (p *Postgres) emailsByUserID(userID int64) ([]Email, error) {
	rows, err := p.db.Query("SELECT email, prim, verified, verification_code FROM Emails WHERE userID = $1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	emails := []Email{}
	for rows.Next() {
		e := Email{}
		if err := rows.Scan(&e.Email, &e.Primary, &e.Verified, &e.VerificationCode); err != nil {
			return nil, err
		}
		emails = append(emails, e)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return emails, nil
}

func (p *Postgres) userByQuery(query string, args ...interface{}) (*User, error) {
	row := p.db.QueryRow(query, args...)

	user := User{}
	err := row.Scan(&user.ID, &user.Name, &user.Pwd)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotExist
	} else if err != nil {
		return nil, err
	}

	emails, err := p.emailsByUserID(user.ID)
	if err != nil {
		return nil, err
	}
	user.Emails = emails

	return &user, nil
}

func (p *Postgres) GetByName(name string) (*User, error) {
	return p.userByQuery("SELECT * FROM Users WHERE name = $1", name)
}

func (p *Postgres) GetByEmail(email string) (*User, error) {
	return p.userByQuery("SELECT * FROM Users WHERE Users.id = (SELECT userID FROM Emails WHERE email = $1)", email)
}

func (p *Postgres) addEmailTx(userID int64, e Email, tx Transaction) error {
	_, err := tx.Exec("INSERT INTO Emails(userID, email, prim, verified, verification_code) VALUES($1, $2, $3, $4, $5)"+
		"ON CONFLICT(email) DO UPDATE SET email = excluded.email, prim = excluded.prim, verified = excluded.verified, verification_code = excluded.verification_code",
		userID, e.Email, e.Primary, e.Verified, e.VerificationCode)
	return err
}

func (p *Postgres) AddEmail(userID int64, e Email) error {
	return p.addEmailTx(userID, e, p.db)
}

func isUniqueViolationError(err error) bool {
	var pgErr *pq.Error
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" {
			return true
		}
	}
	return false
}

func (p *Postgres) Create(u *User) error {
	return WithTransaction(p.db, func(tx Transaction) error {
		var id int64
		err := tx.QueryRow("INSERT INTO Users(name, password) VALUES($1, $2) RETURNING id", u.Name, u.Pwd).Scan(&id)
		if err != nil {
			if isUniqueViolationError(err) {
				return ErrAlreadyExist
			}
			return err
		}

		u.ID = id

		for _, e := range u.Emails {
			err := p.addEmailTx(id, e, tx)
			if err != nil {
				return err
			}
		}

		return nil
	})
}
