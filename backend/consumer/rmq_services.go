package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"go-backend-learning/backend/config"
	"go-backend-learning/backend/manager"
	"go-backend-learning/backend/models"

	"github.com/sirupsen/logrus"
	paotaconfig "github.com/surendratiwari3/paota/config"
	"github.com/surendratiwari3/paota/schema"
	"github.com/surendratiwari3/paota/workerpool"
)

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

type RmqServices struct {
	WorkerPool *workerpool.Pool
	userMgr    *manager.UserManager
}

func (rmqService *RmqServices) initRmqServices(rmqConfig RmqConfig) error {
	rebitmq := paotaconfig.Config{
		Broker:        "amqp",
		TaskQueueName: rmqConfig.QueueName,
		AMQP: &paotaconfig.AMQPConfig{
			Url:                rmqConfig.RmQURL,
			Exchange:           rmqConfig.ExchangeName,
			ExchangeType:       "direct",
			BindingKey:         rmqConfig.BindingKey,
			PrefetchCount:      rmqConfig.PrefetchCount,
			ConnectionPoolSize: rmqConfig.ConnectionPoolSize,
			DelayedQueue:       rmqConfig.DelayedQueue,
			TimeoutQueue:       rmqConfig.TimeoutQueue,
			FailedQueue:        rmqConfig.FailedQueue,
		},
	}
	newWorkerPool, err := workerpool.NewWorkerPoolWithConfig(context.Background(), 10, rmqConfig.QueueName, rebitmq)
	if err != nil {
		return fmt.Errorf("failed to initialize worker pool: %w", err)
	}
	if newWorkerPool == nil {
		return fmt.Errorf("worker pool is nil")
	}

	rmqService.userMgr = new(manager.UserManager)
	rmqService.WorkerPool = &newWorkerPool
	return nil
}

func (rmqService *RmqServices) rmqSubscribe(rmqConfig RmqConfig) error {
	logrusLog := logrus.StandardLogger()
	logrusLog.SetFormatter(&logrus.JSONFormatter{})
	logrusLog.SetReportCaller(true)

	rebitmq := paotaconfig.Config{
		Broker:        "amqp",
		TaskQueueName: rmqConfig.QueueName,
		AMQP: &paotaconfig.AMQPConfig{
			Url:                rmqConfig.RmQURL,
			Exchange:           rmqConfig.ExchangeName,
			ExchangeType:       "direct",
			BindingKey:         rmqConfig.BindingKey,
			PrefetchCount:      rmqConfig.PrefetchCount,
			ConnectionPoolSize: rmqConfig.ConnectionPoolSize,
			DelayedQueue:       rmqConfig.DelayedQueue,
			TimeoutQueue:       rmqConfig.TimeoutQueue,
			FailedQueue:        rmqConfig.FailedQueue,
		},
	}
	err := paotaconfig.GetConfigProvider().SetApplicationConfig(rebitmq)
	if err != nil {
		return err
	}

	newWorkerPool, err := workerpool.NewWorkerPool(context.Background(), 10, rmqConfig.QueueName)
	if err != nil {
		os.Exit(0)
	} else if newWorkerPool == nil {
		os.Exit(0)
	}
	log.Println("newWorkerPool created successfully")

	regTasks := map[string]interface{}{}

	if rmqConfig.QueueName == config.Config.UserTaskProducer.QueueName {
		regTasks[config.Config.UserTaskProducer.QueueTaskName] = rmqService.HandleVerifyUser
	}

	err = newWorkerPool.RegisterTasks(regTasks)
	if err != nil {
		return err
	}

	err = newWorkerPool.Start()
	if err != nil {
		return err
	}

	return nil
}

func (rmqService *RmqServices) HandleVerifyUser(arg *schema.Signature) error {
	if value, ok := arg.Args[0].Value.(string); ok {
		userDetails := models.User{}
		err := json.Unmarshal([]byte(value), &userDetails)
		if err != nil {
			log.Printf("HandleVerifyUser: Logs parsing error: %v", err)
			return nil
		}
		log.Printf("start Verify user Logs: %+v", userDetails)
		fmt.Println("verifyUser", userDetails.ID)
		rmqService.userMgr.VerifyUser(userDetails.ID)
		return nil
	}
	return nil
}
