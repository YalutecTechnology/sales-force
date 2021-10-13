package cache

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	contextKeyTemplate = "context:user_id:%s:timestamp:%d"
	ttl                = 24 * time.Hour
)

// Context is a struct that defines the context related to a conversation between agent and client
type Context struct {
	UserID    string `json:"userId"`
	Timestamp int64  `json:"timestamp,omitempty"`
	URL       string `json:"url,omitempty"`
	MIMEType  string `json:"mimeType,omitempty"`
	Caption   string `json:"caption,omitempty"`
	Text      string `json:"text,omitempty"`
	From      string `json:"from,omitempty"`
}

// ContextCache interface that holds method to retrieve context chat from redis cache
type ContextCache interface {
	StoreContext(Context) error
	RetrieveContext(userID string) []Context
}

// assembleContextKey retrive key by template
func assembleContextKey(context Context) string {
	return fmt.Sprintf(contextKeyTemplate, context.UserID, context.Timestamp)
}

// StoreContext saves a context on Cache
func (rc *RedisCache) StoreContext(context Context) error {
	data, _ := json.Marshal(context)
	return rc.StoreData(assembleContextKey(context), data, ttl)
}

// RetrieveContext returns a context from the Cache
func (rc *RedisCache) RetrieveContext(userID string) []Context {
	var redisContextArray []Context
	keys, err := rc.GetAllKeysWithScanByMatch(fmt.Sprintf("context:user_id:%s:timestamp:*", userID), countScan)
	if err != nil {
		logrus.WithError(err).Error("Redis 'RetrieveAllInterconnections'")
		return nil
	}
	for _, key := range keys {
		var redisContext Context
		data, _ := rc.RetrieveData(key)
		json.Unmarshal([]byte(data), &redisContext)
		redisContextArray = append(redisContextArray, redisContext)
	}

	return redisContextArray
}
