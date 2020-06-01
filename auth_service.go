package main

import (
	"context"

	"github.com/Mikhalevich/filesharing-auth-service/proto"
)

type authService struct {
	//todo
}

func (as *authService) GetByToken(ctx context.Context, req *proto.Token, res *proto.User) error {
	return nil
}

func (as *authService) Create(ctx context.Context, req *proto.User, res *proto.Response) error {
	return nil
}

func (as *authService) Auth(ctx context.Context, req *proto.User, res *proto.Token) error {
	return nil
}
