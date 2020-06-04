package main

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/Mikhalevich/filesharing-auth-service/db"
	"github.com/Mikhalevich/filesharing-auth-service/proto"
	"github.com/Mikhalevich/filesharing-auth-service/token"
	"github.com/micro/go-micro/v2"
	"github.com/micro/go-micro/v2/server"
	"github.com/sirupsen/logrus"
)

type params struct {
	ServiceName        string
	DBConnectionString string
}

func loadParams() (*params, error) {
	var p params
	p.ServiceName = os.Getenv("AS_SERVICE_NAME")
	if p.ServiceName == "" {
		p.ServiceName = "auth.service"
	}

	p.DBConnectionString = os.Getenv("AS_DB_CONNECTION_STRING")
	if p.DBConnectionString == "" {
		return nil, errors.New("databse connection string is missing, please specify AS_DB_CONNECTION_STRING environment variable")
	}

	return &p, nil
}

func makeLoggerWrapper(logger *logrus.Logger) server.HandlerWrapper {
	return func(fn server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) error {
			logger.Infof("processing %s", req.Method())
			start := time.Now()
			defer logger.Infof("end processing %s, time = %v", req.Method(), time.Now().Sub(start))
			err := fn(ctx, req, rsp)
			if err != nil {
				logger.Errorln(err)
			}
			return err
		}
	}
}

func main() {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	p, err := loadParams()
	if err != nil {
		logger.Errorln(err)
		return
	}

	logger.Infof("running auth service with params: service_name = \"%s\"", p.ServiceName)

	service := micro.NewService(
		micro.Name(p.ServiceName),
		micro.WrapHandler(makeLoggerWrapper(logger)),
	)

	service.Init()

	var storage *db.Postgres
	for i := 0; i < 3; i++ {
		storage, err = db.NewPostgres(p.DBConnectionString)
		if err == nil {
			break
		}

		logger.Infof("attemp connect to database: %d  error: %v", i, err)
	}

	if err != nil {
		logger.Errorln(err)
		return
	}
	defer storage.Close()

	proto.RegisterAuthServiceHandler(service.Server(), NewAuthService(storage, token.NewTokenService(time.Hour*72)))

	err = service.Run()
	if err != nil {
		logger.Errorln(err)
		return
	}
}
