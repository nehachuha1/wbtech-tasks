package consumer

import (
	"github.com/nehachuha1/wbtech-tasks/internal/config"
	"github.com/nehachuha1/wbtech-tasks/pkg/log"
	"testing"
)

func TestKafkaManager_ManageMessages(t *testing.T) {
	logger := log.NewLogger("logs.log")
	kafkaConfig := config.NewKafkaConfig()
	kafkaManager := NewKafkaConsumer(kafkaConfig, logger)
	out := make(chan []byte)
	go kafkaManager.ManageMessages(out)
	val := <-out
	t.Logf("%v\n", string(val))
}
