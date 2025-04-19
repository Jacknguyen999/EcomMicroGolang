package main

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/tinrab/retry"
	"log"
	"time"

	"github.com/thomas/EcommerceAPI/account/internal"
	"github.com/thomas/EcommerceAPI/pkg/auth"
)

type Config struct {
	DatabaseURL string `envconfig:"DATABASE_URL" default:"postgres://user:password@localhost/dbname?sslmode=disable"`
	SecretKey   string `envconfig:"SECRET_KEY"`
	Issuer      string `envconfig:"ISSUER"`
}

var (
	cfg        Config
	repository internal.Repository
)

func main() {
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal(err)
	}

	retry.ForeverSleep(2*time.Second, func(_ int) (err error) {
		repository, err = internal.NewPostgresRepository(cfg.DatabaseURL)
		if err != nil {
			log.Println(err)
		}
		return
	})
	defer repository.Close()
	jwtService := auth.NewJwtService(cfg.SecretKey, cfg.Issuer)
	log.Println("Listening on port 8080...")
	service := internal.NewService(repository, jwtService)
	log.Fatal(internal.ListenGRPC(service, 8080))
}
