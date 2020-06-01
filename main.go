package main

import (
	"context"
	"os"
	"time"

	"github.com/Mikhalevich/filesharing-auth-service/proto"
	"github.com/micro/go-micro/v2"
	"github.com/micro/go-micro/v2/server"
	"github.com/sirupsen/logrus"
)

type params struct {
	ServiceName string
}

func loadParams() (*params, error) {
	var p params
	p.ServiceName = os.Getenv("AS_SERVICE_NAME")
	if p.ServiceName == "" {
		p.ServiceName = "auth.service"
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

	proto.RegisterAuthServiceHandler(service.Server(), &authService{})

	err = service.Run()
	if err != nil {
		logger.Errorln(err)
		return
	}
}
