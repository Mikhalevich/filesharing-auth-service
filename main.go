package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/Mikhalevich/filesharing-auth-service/db"
	"github.com/Mikhalevich/filesharing-auth-service/proto"
	"github.com/Mikhalevich/filesharing-auth-service/token"
	"github.com/micro/go-micro/v2"
	"github.com/micro/go-micro/v2/server"
	"github.com/sirupsen/logrus"
)

type params struct {
	ServiceName            string
	DBConnectionString     string
	PrivateCert            string
	TokenExpirePeriodInSec int
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

	p.PrivateCert = os.Getenv("AS_PRIVATE_CERT")
	if p.PrivateCert == "" {
		return nil, errors.New("private certificate is missing, please specify AS_PRIVATE_CERT environment variable")
	}

	p.TokenExpirePeriodInSec = 60 * 60 * 24
	periodString := os.Getenv("AS_TOKEN_EXPIRE_PERIOD_SEC")
	if periodString != "" {
		period, err := strconv.Atoi(periodString)
		if err != nil {
			return nil, fmt.Errorf("unable convert expiration token to integer value %s, error: %w", periodString, err)
		}
		p.TokenExpirePeriodInSec = period
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
		logger.Errorln(fmt.Errorf("unable to load params: %W", err))
		return
	}

	logger.Infof("running auth service with params: %v\n", p)

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

		time.Sleep(time.Second * 1)
		logger.Infof("try to connect to database: %d  error: %v\n", i, err)
	}

	if err != nil {
		logger.Errorln(fmt.Errorf("unable to connect to database: %w", err))
		return
	}
	defer storage.Close()

	privateKey, err := token.LoadCertFromFile(p.PrivateCert)
	if err != nil {
		logger.Errorln(fmt.Errorf("unable to load private certificate: %w", err))
		return
	}

	rsaEncoder, err := token.NewRSAEncoder(privateKey, time.Duration(p.TokenExpirePeriodInSec)*time.Second)
	proto.RegisterAuthServiceHandler(service.Server(), NewAuthService(storage, rsaEncoder))

	err = service.Run()
	if err != nil {
		logger.Errorln(err)
		return
	}
}
