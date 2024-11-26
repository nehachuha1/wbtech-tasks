package producer

import (
	"fmt"
	"github.com/IBM/sarama"
	"github.com/nehachuha1/wbtech-tasks/internal/config"
	"go.uber.org/zap"
	"time"
)

type KafkaProducer struct {
	BrokerURL []string
	Topic     string
	Logger    *zap.SugaredLogger
	Quit      chan bool
}

func NewKafkaProducer(kafkaConfig *config.KafkaConfig, logger *zap.SugaredLogger) *KafkaProducer {
	return &KafkaProducer{
		BrokerURL: []string{kafkaConfig.KafkaURL},
		Topic:     kafkaConfig.Topic,
		Logger:    logger,
		Quit:      make(chan bool),
	}
}

func (kp *KafkaProducer) connectProducer() (sarama.SyncProducer, error) {
	producerConfig := sarama.NewConfig()
	producerConfig.Producer.Return.Successes = true
	producerConfig.Producer.RequiredAcks = sarama.WaitForLocal
	producerConfig.Producer.Retry.Max = 5

	return sarama.NewSyncProducer(kp.BrokerURL, producerConfig)
}

func (kp *KafkaProducer) PushOrderToQueue(data []byte) error {
	producer, err := kp.connectProducer()
	if err != nil {
		kp.Logger.Fatalw(fmt.Sprintf("failed on initialize producer to topic %v", kp.Topic),
			"source", "service/kafka/consumer", "time", time.Now().String())
		return err
	}
	kp.Logger.Infow(fmt.Sprintf("initialized producer for topic %v", kp.Topic),
		"source", "service/kafka/consumer", "time", time.Now().String())
	defer producer.Close()

	msg := &sarama.ProducerMessage{
		Topic: kp.Topic,
		Value: sarama.StringEncoder(data),
	}
	_, _, err = producer.SendMessage(msg)
	if err != nil {
		kp.Logger.Fatalw(fmt.Sprintf("failed to send message to topic %v", kp.Topic),
			"source", "service/kafka/consumer", "time", time.Now().String())
		return err
	}
	kp.Logger.Infow(fmt.Sprintf("sent message to topic %v", kp.Topic),
		"source", "service/kafka/consumer", "time", time.Now().String())
	return nil
}
