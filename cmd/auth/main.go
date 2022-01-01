package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/asim/go-micro/v3"

	"github.com/Mikhalevich/filesharing-auth-service/internal/db"
	"github.com/Mikhalevich/filesharing-auth-service/internal/handler"
	"github.com/Mikhalevich/filesharing-auth-service/pkg/token"
	"github.com/Mikhalevich/filesharing/pkg/proto/auth"
	"github.com/Mikhalevich/filesharing/pkg/service"
	"github.com/Mikhalevich/repeater"
)

type config struct {
	service.Config       `yaml:"service"`
	DB                   string `yaml:"db"`
	TokenExpirePeriodSec int    `yaml:"token_expire_period"`
}

func (c *config) Service() service.Config {
	return c.Config
}

func (c *config) Validate() error {
	if c.DB == "" {
		return errors.New("db is required")
	}

	if c.TokenExpirePeriodSec <= 0 {
		return fmt.Errorf("invalid token_expire_period: %d", c.TokenExpirePeriodSec)
	}

	return nil
}

func main() {
	var cfg config
	service.Run(os.Getenv("FS_SERVICE_NAME"), &cfg, func(srv micro.Service, s service.Servicer) error {
		var storage *db.Postgres
		if err := repeater.Do(
			func() error {
				var err error
				storage, err = db.NewPostgres(cfg.DB)
				return err
			},
			repeater.WithTimeout(time.Second*1),
			repeater.WithLogger(s.Logger()),
			repeater.WithLogMessage("try to connect to database"),
		); err != nil {
			return fmt.Errorf("unable to connect to database: %w", err)
		}

		s.AddOption(service.WithPostAction(func() {
			storage.Close()
		}))

		rsaEncoder, err := token.NewRSAEncoder(time.Duration(cfg.TokenExpirePeriodSec) * time.Second)
		if err != nil {
			return fmt.Errorf("unable to create auth encoded: %w", err)
		}

		if err := auth.RegisterAuthServiceHandler(srv.Server(), handler.New(storage, rsaEncoder)); err != nil {
			return fmt.Errorf("unable register auth handler: %v", err)
		}

		return nil
	})
}
