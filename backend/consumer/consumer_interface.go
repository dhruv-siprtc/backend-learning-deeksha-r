package consumer

type ConsumerServiceInterface interface {
	Initialize(rmqConfig RmqConfig) error
}
