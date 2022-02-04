package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/sirupsen/logrus"
	"yalochat.com/salesforce-integration/base/subscribers"
)

type reader interface {
	ReadMessage(time.Duration) (*kafka.Message, error)
	Close() error
}

type consumer struct {
	host    string
	port    string
	rTopic  string
	reader  reader
	service subscribers.Service
}

func NewConsumer(s subscribers.Service, topicreader, serviceName, offset string, settings KafkaSettings) subscribers.Consumer {
	host := settings.Host
	port := settings.Port
	user := settings.User
	password := settings.Password
	return &consumer{
		host:    host,
		port:    port,
		rTopic:  topicreader,
		reader:  newReader(topicreader, host, port, serviceName, offset, user, password),
		service: s,
	}
}

func newReader(topic, host, port, serviceName, offset, user, password string) reader {

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": fmt.Sprintf("%v:%v", host, port),
		"group.id":          fmt.Sprintf("%s-%s", serviceName, topic),
		"auto.offset.reset": offset,
		"sasl.mechanisms":   "PLAIN",
		"security.protocol": "SASL_SSL",
		"sasl.username":     user,
		"sasl.password":     password,
	})

	if err != nil {
		panic(err)
	}

	c.Subscribe(topic, nil)
	return c
}

func (s consumer) Start() {
	defer s.reader.Close()
	for {
		s.readMessage()
	}
}

func (s consumer) readMessage() {
	m, err := s.reader.ReadMessage(-1)
	if err != nil {
		logrus.Errorf("[reading_err:%s][topic:%s]", err, s.rTopic)
		return
	}

	err = s.service.Process(context.Background(), m.Value)
	if err != nil {
		logrus.Errorf("[process_error:%s][topic:%s]", err, s.rTopic)
	}
}
