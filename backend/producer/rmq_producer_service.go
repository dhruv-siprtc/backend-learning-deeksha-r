package producer

type RmqConfig struct {
	QueueName          string
	ExchangeName       string
	BindingKey         string
	PrefetchCount      int
	ConnectionPoolSize int
	DelayedQueue       string
	RmQURL             string
	FailedQueue        string
	TimeoutQueue       string
}

type RmqProducerService struct {
	rmqService RmqServices
	rmqConfig  RmqConfig
}

func (rebitmq *RmqProducerService) Initialize(rmqConfig RmqConfig) error {
	rebitmq.rmqService = RmqServices{}
	err := rebitmq.rmqService.initRmqServices(rmqConfig)
	rebitmq.rmqConfig = rmqConfig
	return err
}

func (rebitmq *RmqProducerService) Publish(message []byte, taskName string, queueName string) error {
	return rebitmq.rmqService.rmqPublish(message, taskName, queueName)
}
