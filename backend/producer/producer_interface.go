package producer

type ProducerServiceInterface interface {
	Initialize(rmqConfig RmqConfig) error
	Publish(message []byte, taskName string, queueName string) error
}
