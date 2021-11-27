package handler

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"github.com/Mikhalevich/filesharing-auth-service/internal/db"
	"github.com/Mikhalevich/filesharing-auth-service/pkg/token"
	"github.com/Mikhalevich/filesharing/httpcode"
	"github.com/Mikhalevich/filesharing/proto/auth"
	"github.com/Mikhalevich/filesharing/proto/types"
)

type storager interface {
	GetByName(name string) (*db.User, error)
	GetByEmail(email string) (*db.User, error)
	Create(u *db.User) error
	GetPublicUsers() ([]*db.User, error)
}

type handler struct {
	repo    storager
	encoder token.Encoder
}

func New(s storager, te token.Encoder) *handler {
	return &handler{
		repo:    s,
		encoder: te,
	}
}

func unmarshalUser(u *types.User) *db.User {
	user := db.User{
		ID:     u.GetId(),
		Name:   u.GetName(),
		Public: u.GetPublic(),
	}

	user.Pwd.String = u.GetPassword()
	return &user
}

func marshalUser(u *db.User) *types.User {
	return &types.User{
		Id:       u.ID,
		Name:     u.Name,
		Password: u.Pwd.String,
		Public:   u.Public,
	}
}

func (as *handler) Create(ctx context.Context, req *auth.CreateUserRequest, rsp *auth.CreateUserResponse) error {
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(req.GetUser().GetPassword()), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("[Create] hashing password error: %w", err)
	}

	ru := req.GetUser()
	if ru == nil {
		return errors.New("[Create] invalid user")
	}
	ru.Password = string(hashedPass)
	user := unmarshalUser(ru)
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
	rsp.Token = &types.Token{
		Value: token,
	}
	return nil
}

func (as *handler) Auth(ctx context.Context, req *auth.AuthUserRequest, rsp *auth.AuthUserResponse) error {
	user, err := as.repo.GetByName(req.GetUser().GetName())
	if errors.Is(err, db.ErrNotExist) {
		rsp.Status = auth.Status_NotExist
		return nil
	} else if err != nil {
		return fmt.Errorf("[Auth] get user error: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Pwd.String), []byte(req.GetUser().GetPassword()))
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
	rsp.Token = &types.Token{
		Value: token,
	}
	return nil
}

func (as *handler) AuthPublicUser(ctx context.Context, req *auth.AuthPublicUserRequest, rsp *auth.AuthPublicUserResponse) error {
	user, err := as.repo.GetByName(req.GetName())
	if errors.Is(err, db.ErrNotExist) {
		return httpcode.NewNotExistError("user not exist")
	} else if err != nil {
		return httpcode.NewWrapInternalServerError(err, "unable to get user")
	}

	if !user.Public {
		return httpcode.NewBadRequest("user is not public")
	}

	token, err := as.encoder.Encode(token.User{
		ID:     user.ID,
		Name:   user.Name,
		Public: user.Public,
	})

	if err != nil {
		return httpcode.NewWrapInternalServerError(err, "unable to encode token")
	}

	rsp.Token = &types.Token{
		Value: token,
	}
	return nil
}
