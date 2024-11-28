package consumer

import (
	"fmt"
	"github.com/IBM/sarama"
	"github.com/nehachuha1/wbtech-tasks/internal/config"
	"go.uber.org/zap"
)

// Управляющая структура для работы с получателем сообщений в кафке.
type KafkaConsumer struct {
	BrokerURL []string
	Topic     string
	Logger    *zap.SugaredLogger
	Quit      chan bool
}

// Инициализация управляющей структуры
func NewKafkaConsumer(kafkaConfig *config.KafkaConfig, logger *zap.SugaredLogger) *KafkaConsumer {
	kafkaManager := &KafkaConsumer{
		BrokerURL: []string{kafkaConfig.KafkaURL},
		Topic:     kafkaConfig.Topic,
		Logger:    logger,
	}

	return kafkaManager
}

// Подключение нового получателя к брокеру очередей.
func (km *KafkaConsumer) connectConsumer() (sarama.Consumer, error) {
	cfg := sarama.NewConfig()
	cfg.Consumer.Return.Errors = true

	return sarama.NewConsumer(km.BrokerURL, cfg)
}

// Инициализация получателя для партиции топика. Далее в горутине "слушаем" сигнал на канал выхода quit, если ловим его,
// то закрываем подключение
func (km *KafkaConsumer) InitializeConsumer(quit chan bool) (sarama.PartitionConsumer, error) {
	worker, err := km.connectConsumer()
	if err != nil {
		km.Logger.Warn(fmt.Sprintf("failed on consumer connection: %v", err))
		return nil, err
	}
	consumer, err := worker.ConsumePartition(km.Topic, 0, sarama.OffsetNewest)
	if err != nil {
		km.Logger.Warn(fmt.Sprintf("failed on consume partition in topic %v: %v", km.Topic, err))
		return nil, err
	}

	km.Logger.Info(fmt.Sprintf("consumer for topic %v started", km.Topic))
	go func() {
		for {
			select {
			case <-quit:
				if err = worker.Close(); err != nil {
					km.Logger.Warn(fmt.Sprintf("failed on closing consumer connection: %v", err))
					return
				}
				km.Logger.Info(fmt.Sprintf("successfully closed consumer of topic %v", km.Topic))
				return
			}
		}
	}()

	return consumer, nil
}
