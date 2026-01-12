package producer

import (
	"encoding/json"

	"go-backend-learning/backend/config"
	"go-backend-learning/backend/models"
)

type ProducerServiceManager struct {
	produceService ProducerServiceInterface
}

var UserProducer ProducerServiceManager

func (producerServiceManager *ProducerServiceManager) Initialize(rmqConfig RmqConfig) error {
	producerServiceManager.produceService = new(RmqProducerService)
	return producerServiceManager.produceService.Initialize(rmqConfig)
}

func (producerServiceManager *ProducerServiceManager) ProcesswithVerifyUser(userDetails models.User, queueName string) error {
	userData, err := json.Marshal(userDetails)
	if err != nil {
		return err
	}
	err = producerServiceManager.produceService.Publish(userData, config.Config.UserTaskProducer.QueueTaskName, queueName)
	return err
}
