package consumer

import (
	"log"
)

type ConsumerService struct {
	rmqServe *RmqServices
}

// init consumer service
func (consumer *ConsumerService) Initialize(rmqConfig RmqConfig) error {
	var err error
	consumer.rmqServe = new(RmqServices)
	err = consumer.rmqServe.initRmqServices(rmqConfig)
	if err != nil {
		log.Fatal(err)
	}
	consumer.rmqServe.rmqSubscribe(rmqConfig)

	return err
}
