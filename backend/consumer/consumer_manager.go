package consumer

type ConsumerManager struct {
	consumerService ConsumerServiceInterface
}

func (consumermanager *ConsumerManager) Initialize(rmqConfig RmqConfig) error {
	consumermanager.consumerService = new(ConsumerService)
	return consumermanager.consumerService.Initialize(rmqConfig)
}
