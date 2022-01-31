package kafka

type KafkaSettings struct {
	Host     string
	Port     string
	User     string
	Password string
}

type KafkaMessageParams struct {
	Msg   []byte
	Key   []byte
	Topic string
}
