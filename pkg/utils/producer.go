package utils

import (
	"encoding/json"
	"github.com/IBM/sarama"
	"log"
)

var done = make(chan bool)

type Service interface {
	Producer() sarama.AsyncProducer
}

func SendMessageToRecommender(service Service, event any, topic string) error {
	jsonMessage, err := json.Marshal(event)
	if err != nil {
		log.Println("Failed to marshal event:", err)
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(jsonMessage),
	}

	// Send the message asynchronously
	service.Producer().Input() <- msg

	return nil
}

func MsgHandler(service Service) {
	go func() {
		for {
			select {
			case success := <-service.Producer().Successes():
				log.Printf("Message sent to partition %d at offset %d\n", success.Partition, success.Offset)
			case err := <-service.Producer().Errors():
				log.Printf("Failed to send message: %v\n", err)
			case <-done:
				log.Println("Producer closed successfully")
				return
			}
		}
	}()
}

func Close(service Service) {
	if err := service.Producer().Close(); err != nil {
		log.Printf("Failed to close producer: %v\n", err)
	} else {
		done <- true
	}
}
