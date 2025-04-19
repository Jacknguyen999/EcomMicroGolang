package main

import (
	"log"
	"time"

	"github.com/IBM/sarama"
	"github.com/kelseyhightower/envconfig"
	"github.com/tinrab/retry"

	"github.com/thomas/EcommerceAPI/product/internal"
)

type Config struct {
	DatabaseURL      string `envconfig:"DATABASE_URL"`
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

	// Configure Kafka with retry
	retry.ForeverSleep(2*time.Second, func(_ int) (err error) {
		// Configure Kafka
		config := sarama.NewConfig()
		config.Producer.Return.Successes = false
		config.Producer.Return.Errors = true

		// Create Kafka producer with retry
		producer, err = sarama.NewAsyncProducer([]string{cfg.BootstrapServers}, config)
		if err != nil {
			log.Println("Failed to connect to Kafka:", err)
			return err
		}
		log.Println("Successfully connected to Kafka")
		return nil
	})

	// Handle Kafka producer errors in a goroutine
	if producer != nil {
		go func() {
			for err := range producer.Errors() {
				log.Println("Failed to write message to Kafka:", err)
			}
		}()
		defer producer.Close()
	}

	// Connect to Elasticsearch with retry
	retry.ForeverSleep(2*time.Second, func(_ int) (err error) {
		repository, err = internal.NewElasticRepository(cfg.DatabaseURL)
		if err != nil {
			log.Println("Failed to connect to Elasticsearch:", err)
			return err
		}
		log.Println("Successfully connected to Elasticsearch")
		return nil
	})

	if repository != nil {
		defer repository.Close()
	}

	log.Println("Listening on port 8080...")
	service := internal.NewProductService(repository, producer)
	log.Fatal(internal.ListenGRPC(service, 8080))
}
