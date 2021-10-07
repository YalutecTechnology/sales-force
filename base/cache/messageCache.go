package cache

import (
	"fmt"
	"time"

	"yalochat.com/salesforce-integration/base/constants"
)

const (
	ttlMessage = time.Second * 6
)

type MessageCache struct {
	cache ICache
}

func NewMessageCache(cache ICache) *MessageCache {
	return &MessageCache{cache: cache}
}

type IMessageCache interface {
	IsRepeatedMessage(string) bool
}

func (mc *MessageCache) IsRepeatedMessage(key string) bool {
	key = assembleMessageKey(key)
	_, ok := mc.cache.Get(key)
	if !ok {
		mc.cache.Set(key, struct{}{}, ttlMessage)
		return false
	}

	return true
}

func assembleMessageKey(key string) string {
	return fmt.Sprintln(constants.MessageKey, key)
}
