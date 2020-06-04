package main

import (
	"context"
	"fmt"

	"github.com/Mikhalevich/filesharing-auth-service/db"
	"github.com/Mikhalevich/filesharing-auth-service/proto"
	"github.com/Mikhalevich/filesharing-auth-service/token"
	"golang.org/x/crypto/bcrypt"
)

type storager interface {
	GetByName(name string) (*db.User, error)
	GetByEmail(email string) (*db.User, error)
	Create(u *db.User) error
}

type tokener interface {
	Decode(token string) (*token.CustomClaims, error)
	Encode(user token.User) (string, error)
}

type AuthService struct {
	repo     storager
	tokenSrv tokener
}

func NewAuthService(s storager, tsrv tokener) *AuthService {
	return &AuthService{
		repo:     s,
		tokenSrv: tsrv,
	}
}

func unmarshalUser(u *proto.User) *db.User {
	return &db.User{
		ID:   u.GetId(),
		Name: u.GetName(),
		Pwd:  u.GetPassword(),
	}
}

func (as *AuthService) Create(ctx context.Context, req *proto.User, rsp *proto.Response) error {
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("[Create] hashing password error: %w", err)
	}

	req.Password = string(hashedPass)
	err = as.repo.Create(unmarshalUser(req))
	if err != nil {
		return fmt.Errorf("[Create] creating user error: %w", err)
	}

	token, err := as.tokenSrv.Encode(token.User{
		Name: req.GetName(),
	})

	if err != nil {
		return fmt.Errorf("[Craate] unable to encode token: %w", err)
	}

	rsp.Status = proto.Status_Ok
	rsp.Token = token
	return nil
}

func (as *AuthService) Auth(ctx context.Context, req *proto.User, rsp *proto.Response) error {
	user, err := as.repo.GetByName(req.GetName())
	if err != nil {
		return fmt.Errorf("[Auth] get user error: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Pwd), []byte(req.Password))
	if err != nil {
		return fmt.Errorf("[Auth] password not match: %w", err)
	}

	token, err := as.tokenSrv.Encode(token.User{
		Name: user.Name,
	})

	if err != nil {
		return fmt.Errorf("[Auth] unable to encode token: %w", err)
	}

	rsp.Status = proto.Status_Ok
	rsp.Token = token
	return nil
}
