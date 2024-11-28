package producer

import (
	"fmt"
	"github.com/IBM/sarama"
	"github.com/nehachuha1/wbtech-tasks/internal/config"
	"go.uber.org/zap"
)

// Управляющая структура для работы с отправителем сообщений в кафке.
type KafkaProducer struct {
	BrokerURL []string
	Topic     string
	Logger    *zap.SugaredLogger
	Quit      chan bool
}

// Инициализация управляющей структуры
func NewKafkaProducer(kafkaConfig *config.KafkaConfig, logger *zap.SugaredLogger) *KafkaProducer {
	return &KafkaProducer{
		BrokerURL: []string{kafkaConfig.KafkaURL},
		Topic:     kafkaConfig.Topic,
		Logger:    logger,
		Quit:      make(chan bool),
	}
}

// Подключение нового продюсера к брокеру очередей.
func (kp *KafkaProducer) connectProducer() (sarama.SyncProducer, error) {
	producerConfig := sarama.NewConfig()
	producerConfig.Producer.Return.Successes = true
	producerConfig.Producer.RequiredAcks = sarama.WaitForLocal
	producerConfig.Producer.Retry.Max = 5

	return sarama.NewSyncProducer(kp.BrokerURL, producerConfig)
}

// Основной метод структуры для пуша сообщений в очередь. При вызове метода каждый раз инициализируем нового продюсера,
// пушим сообщение в топик и закрываем текущего продюсера
func (kp *KafkaProducer) PushOrderToQueue(data []byte) error {
	producer, err := kp.connectProducer()
	if err != nil {
		kp.Logger.Fatal(fmt.Sprintf("failed on initialize producer to topic %v", kp.Topic))
		return err
	}
	kp.Logger.Info(fmt.Sprintf("initialized producer for topic %v", kp.Topic))
	defer producer.Close()

	msg := &sarama.ProducerMessage{
		Topic: kp.Topic,
		Value: sarama.StringEncoder(data),
	}
	_, _, err = producer.SendMessage(msg)
	if err != nil {
		kp.Logger.Fatal(fmt.Sprintf("failed to send message to topic %v", kp.Topic))
		return err
	}
	kp.Logger.Info(fmt.Sprintf("sent message to topic %v", kp.Topic))
	return nil
}
