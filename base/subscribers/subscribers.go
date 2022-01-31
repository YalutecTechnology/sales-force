package subscribers

import "context"

type Producer interface {
	SendMessage(msg interface{}) error
}

type Consumer interface {
	Start()
}

type Service interface {
	Process(ctx context.Context, message []byte) error
}
