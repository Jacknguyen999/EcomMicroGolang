package main

import (
	"github.com/IBM/sarama"
	"github.com/kelseyhightower/envconfig"
	internal "github.com/thomas/EcommerceAPI/order/internal"
	"github.com/tinrab/retry"
	"log"
	"time"
)

type Config struct {
	DatabaseUrl      string `envconfig:"DATABASE_URL"`
	AccountUrl       string `envconfig:"ACCOUNT_SERVICE_URL"`
	ProductUrl       string `envconfig:"PRODUCT_SERVICE_URL"`
	BootstrapServers string `envconfig:"BOOTSTRAP_SERVERS" default:"kafka:9092"`
}

func main() {
	var cfg Config
	var repository internal.Repository
	var producer sarama.AsyncProducer

	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal(err)
	}

	producer, err = sarama.NewAsyncProducer([]string{cfg.BootstrapServers}, nil)
	if err != nil {
		log.Println(err)
	}
	defer producer.Close()

	retry.ForeverSleep(2*time.Second, func(_ int) (err error) {
		repository, err = internal.NewPostgresRepository(cfg.DatabaseUrl)
		if err != nil {
			log.Println(err)
		}
		return
	})
	defer repository.Close()
	log.Println("Listening on port 8080...")
	service := internal.NewOrderService(repository, producer)
	log.Fatal(internal.ListenGRPC(service, cfg.AccountUrl, cfg.ProductUrl, 8080))
}
