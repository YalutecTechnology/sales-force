package kafka

import (
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/sirupsen/logrus"
	"yalochat.com/salesforce-integration/base/subscribers"
)

type producer interface {
	ProduceChannel() chan *kafka.Message
	Events() chan kafka.Event
}

type publisher struct {
	producer
}

var myPublisher *publisher

func NewProducer(settings KafkaSettings) subscribers.Producer {

	if myPublisher == nil {
		prefix, producer := newProducer(settings.Host, settings.Port, settings.User, settings.Password)
		myPublisher = &publisher{producer}
		go myPublisher.readEvents(prefix)
	}

	return myPublisher
}

func newProducer(host, port, user, password string) (string, producer) {
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": fmt.Sprintf(`%s:%s`, host, port),
		"sasl.mechanisms":   "PLAIN",
		"security.protocol": "SASL_SSL",
		"sasl.username":     user,
		"sasl.password":     password,
	})
	if err != nil {
		logrus.Errorf(`[producer_new:%s]`, err.Error())
	}

	return "", producer
}

func (p *publisher) SendMessage(prms interface{}) error {
	params, ok := prms.(KafkaMessageParams)
	if !ok {
		return fmt.Errorf("the params are not valid. Please, review the documentation about KafkaMessageParams")
	}

	message := kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &params.Topic,
			Partition: kafka.PartitionAny,
		},
		Value: params.Msg,
	}

	if len(params.Key) > 0 {
		message.Key = params.Key
	}

	p.producer.ProduceChannel() <- &(message)

	return nil
}

func (p *publisher) readEvents(prefix string) {

	for e := range p.producer.Events() {
		switch ev := e.(type) {
		case *kafka.Message:
			m := ev
			if m.TopicPartition.Error != nil {
				logrus.Errorf("[kafka_producer][Delivery_failed:%v]", m.TopicPartition.Error.Error())
			}

			continue
		}
	}
}
