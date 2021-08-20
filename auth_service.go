package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/Mikhalevich/filesharing-auth-service/db"
	"github.com/Mikhalevich/filesharing-auth-service/token"
	"github.com/Mikhalevich/filesharing/proto/auth"
	"golang.org/x/crypto/bcrypt"
)

type storager interface {
	GetByName(name string) (*db.User, error)
	GetByEmail(email string) (*db.User, error)
	Create(u *db.User) error
}

type AuthService struct {
	repo    storager
	encoder token.Encoder
}

func NewAuthService(s storager, te token.Encoder) *AuthService {
	return &AuthService{
		repo:    s,
		encoder: te,
	}
}

func unmarshalUser(u *auth.User) *db.User {
	return &db.User{
		ID:   u.GetId(),
		Name: u.GetName(),
		Pwd:  u.GetPassword(),
	}
}

func (as *AuthService) Create(ctx context.Context, req *auth.User, rsp *auth.Response) error {
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("[Create] hashing password error: %w", err)
	}

	req.Password = string(hashedPass)
	user := unmarshalUser(req)
	err = as.repo.Create(user)
	if errors.Is(err, db.ErrAlreadyExist) {
		rsp.Status = auth.Status_AlreadyExist
		return nil
	} else if err != nil {
		return fmt.Errorf("[Create] creating user error: %w", err)
	}

	token, err := as.encoder.Encode(token.User{
		ID:     user.ID,
		Name:   user.Name,
		Public: user.Public,
	})

	if err != nil {
		return fmt.Errorf("[Create] unable to encode token: %w", err)
	}

	rsp.Status = auth.Status_Ok
	rsp.Token = token
	return nil
}

func (as *AuthService) Auth(ctx context.Context, req *auth.User, rsp *auth.Response) error {
	user, err := as.repo.GetByName(req.GetName())
	if errors.Is(err, db.ErrNotExist) {
		rsp.Status = auth.Status_NotExist
		return nil
	} else if err != nil {
		return fmt.Errorf("[Auth] get user error: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Pwd), []byte(req.Password))
	if err != nil {
		rsp.Status = auth.Status_PwdNotMatch
		return nil
	}

	token, err := as.encoder.Encode(token.User{
		ID:     user.ID,
		Name:   user.Name,
		Public: user.Public,
	})

	if err != nil {
		return fmt.Errorf("[Auth] unable to encode token: %w", err)
	}

	rsp.Status = auth.Status_Ok
	rsp.Token = token
	return nil
}
