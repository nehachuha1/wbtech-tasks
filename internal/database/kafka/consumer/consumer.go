package consumer

import (
	"fmt"
	"github.com/IBM/sarama"
	"github.com/nehachuha1/wbtech-tasks/internal/config"
	"go.uber.org/zap"
	"time"
)

type KafkaConsumer struct {
	BrokerURL []string
	Topic     string
	Logger    *zap.SugaredLogger
	Quit      chan bool
}

func NewKafkaConsumer(kafkaConfig *config.KafkaConfig, logger *zap.SugaredLogger) *KafkaConsumer {
	kafkaManager := &KafkaConsumer{
		BrokerURL: []string{kafkaConfig.KafkaURL},
		Topic:     kafkaConfig.Topic,
		Logger:    logger,
	}

	return kafkaManager
}

func (km *KafkaConsumer) connectConsumer() (sarama.Consumer, error) {
	cfg := sarama.NewConfig()
	cfg.Consumer.Return.Errors = true

	return sarama.NewConsumer(km.BrokerURL, cfg)
}

func (km *KafkaConsumer) InitializeConsumer(quit chan bool) (sarama.PartitionConsumer, error) {
	worker, err := km.connectConsumer()
	if err != nil {
		return nil, err
	}
	consumer, err := worker.ConsumePartition(km.Topic, 0, sarama.OffsetNewest)
	if err != nil {
		return nil, err
	}
	km.Logger.Infow(fmt.Sprintf("consumer for topic %v started",
		km.Topic), "source", "service/kafka/consumer", "time", time.Now().String())

	go func() {
		for {
			select {
			case <-quit:
				if err = worker.Close(); err != nil {
					// TODO: add exception proceed
					return
				}
				// TODO: add exception proceed
				return
			}
		}
	}()

	return consumer, nil
}
