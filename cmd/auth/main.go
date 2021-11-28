package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/asim/go-micro/v3"

	"github.com/Mikhalevich/filesharing-auth-service/internal/db"
	"github.com/Mikhalevich/filesharing-auth-service/internal/handler"
	"github.com/Mikhalevich/filesharing-auth-service/pkg/token"
	"github.com/Mikhalevich/filesharing/pkg/proto/auth"
	"github.com/Mikhalevich/filesharing/pkg/service"
	"github.com/Mikhalevich/repeater"
)

type params struct {
	ServiceName            string
	DBConnectionString     string
	TokenExpirePeriodInSec int
}

func loadParams() (*params, error) {
	var p params
	p.ServiceName = os.Getenv("FS_SERVICE_NAME")
	if p.ServiceName == "" {
		p.ServiceName = "auth.service"
	}

	p.DBConnectionString = os.Getenv("FS_DB_CONNECTION_STRING")
	if p.DBConnectionString == "" {
		return nil, errors.New("databse connection string is missing, please specify AS_DB_CONNECTION_STRING environment variable")
	}

	p.TokenExpirePeriodInSec = 60 * 60 * 24
	periodString := os.Getenv("FS_TOKEN_EXPIRE_PERIOD_SEC")
	if periodString != "" {
		period, err := strconv.Atoi(periodString)
		if err != nil {
			return nil, fmt.Errorf("unable convert expiration token to integer value %s, error: %w", periodString, err)
		}
		p.TokenExpirePeriodInSec = period
	}

	return &p, nil
}

func main() {
	p, err := loadParams()
	if err != nil {
		fmt.Printf("unable to load params: %v", err)
		return
	}

	srv, err := service.New(p.ServiceName)
	if err != nil {
		fmt.Printf("unable to create service: %v", err)
		return
	}

	var storage *db.Postgres
	if err := repeater.Do(
		func() error {
			storage, err = db.NewPostgres(p.DBConnectionString)
			return err
		},
		repeater.WithTimeout(time.Second*1),
		repeater.WithLogger(srv.Logger()),
		repeater.WithLogMessage("try to connect to database"),
	); err != nil {
		srv.Logger().WithError(err).Error("unable to connect to database")
		return
	}
	defer storage.Close()

	if err := srv.RegisterHandler(func(ms micro.Service, s service.Servicer) error {
		rsaEncoder, err := token.NewRSAEncoder(time.Duration(p.TokenExpirePeriodInSec) * time.Second)
		if err != nil {
			return fmt.Errorf("unable to create auth encoded: %v", err)
		}

		if err := auth.RegisterAuthServiceHandler(ms.Server(), handler.New(storage, rsaEncoder)); err != nil {
			return fmt.Errorf("unable register auth handler: %v", err)
		}
		return nil
	}); err != nil {
		srv.Logger().WithError(err).Error("failed to register handler")
		return
	}

	srv.Logger().WithFields(map[string]interface{}{
		"params": p,
	}).Info("service running")

	srv.Run()
}
