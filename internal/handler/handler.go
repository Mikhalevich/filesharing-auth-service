package handler

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"

	"github.com/Mikhalevich/filesharing-auth-service/internal/db"
	"github.com/Mikhalevich/filesharing-auth-service/pkg/token"
	"github.com/Mikhalevich/filesharing/pkg/httperror"
	"github.com/Mikhalevich/filesharing/pkg/proto/auth"
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

func unmarshalUser(u *auth.User) *db.User {
	user := db.User{
		ID:     u.GetId(),
		Name:   u.GetName(),
		Public: u.GetPublic(),
	}

	user.Pwd.String = u.GetPassword()
	return &user
}

// func marshalUser(u *db.User) *auth.User {
// 	return &auth.User{
// 		Id:       u.ID,
// 		Name:     u.Name,
// 		Password: u.Pwd.String,
// 		Public:   u.Public,
// 	}
// }

func (as *handler) Create(ctx context.Context, req *auth.CreateUserRequest, rsp *auth.CreateUserResponse) error {
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(req.GetUser().GetPassword()), bcrypt.DefaultCost)
	if err != nil {
		return httperror.NewInternalError("hash password error").WithError(err)
	}

	ru := req.GetUser()
	if ru == nil {
		return httperror.NewInvalidParams("invalid user")
	}

	ru.Password = string(hashedPass)
	user := unmarshalUser(ru)
	if err := as.repo.Create(user); err != nil {
		if errors.Is(err, db.ErrAlreadyExist) {
			return httperror.NewAlreadyExistError("user already exists")
		}
		return httperror.NewInternalError("create user error").WithError(err)
	}

	token, err := as.encoder.Encode(token.User{
		ID:     user.ID,
		Name:   user.Name,
		Public: user.Public,
	})

	if err != nil {
		return httperror.NewInternalError("encode token").WithError(err)
	}

	rsp.Token = &auth.Token{
		Value: token,
	}
	return nil
}

func (as *handler) Auth(ctx context.Context, req *auth.AuthUserRequest, rsp *auth.AuthUserResponse) error {
	user, err := as.repo.GetByName(req.GetUser().GetName())
	if err != nil {
		if errors.Is(err, db.ErrNotExist) {
			return httperror.NewNotExistError("user not exists")
		}
		return httperror.NewInternalError("get user error").WithError(err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Pwd.String), []byte(req.GetUser().GetPassword())); err != nil {
		return httperror.NewNotMatchError("pwd not match")
	}

	token, err := as.encoder.Encode(token.User{
		ID:     user.ID,
		Name:   user.Name,
		Public: user.Public,
	})

	if err != nil {
		return httperror.NewInternalError("encode token").WithError(err)
	}

	rsp.Token = &auth.Token{
		Value: token,
	}
	return nil
}

func (as *handler) AuthPublicUser(ctx context.Context, req *auth.AuthPublicUserRequest, rsp *auth.AuthPublicUserResponse) error {
	user, err := as.repo.GetByName(req.GetName())
	if err != nil {
		if errors.Is(err, db.ErrNotExist) {
			return httperror.NewNotExistError("user not exist")
		}
		return httperror.NewInternalError("unable to get user").WithError(err)
	}

	if !user.Public {
		return httperror.NewInvalidParams("user is not public")
	}

	token, err := as.encoder.Encode(token.User{
		ID:     user.ID,
		Name:   user.Name,
		Public: user.Public,
	})

	if err != nil {
		return httperror.NewInternalError("unable to encode token").WithError(err)
	}

	rsp.Token = &auth.Token{
		Value: token,
	}
	return nil
}
