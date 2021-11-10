package cache

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	contextKeyTemplate    = "context:user_id:%s:timestamp:%d"
	Ttl                   = 24 * time.Hour
	contextSetKeyTemplate = "%s:%s:context"
)

// Context is a struct that defines the context related to a conversation between agent and client
type Context struct {
	UserID    string    `json:"userId"`
	Client    string    `json:"client"`
	Timestamp int64     `json:"timestamp,omitempty"`
	URL       string    `json:"url,omitempty"`
	MIMEType  string    `json:"mimeType,omitempty"`
	Caption   string    `json:"caption,omitempty"`
	Text      string    `json:"text,omitempty"`
	From      string    `json:"from,omitempty"`
	Ttl       time.Time `json:"ttl,omitempty"`
}

// ContextCache interface that holds method to retrieve context chat from redis cache
type ContextCache interface {
	StoreContext(Context) error
	RetrieveContext(userID string) []Context
	StoreContextToSet(Context) error
	RetrieveContextFromSet(client, userID string) []Context
	CleanContextToDate(client string, dateTime time.Time) error
}

// assembleContextKey retrieve key by template
func assembleContextKey(context Context) string {
	return fmt.Sprintf(contextKeyTemplate, context.UserID, context.Timestamp)
}

// assembleContextSetKey retrieve key by template
func assembleContextSetKey(context Context) string {
	return fmt.Sprintf(contextSetKeyTemplate, context.Client, context.UserID)
}

// StoreContext saves a context on Cache
func (rc *RedisCache) StoreContext(context Context) error {
	data, _ := json.Marshal(context)
	return rc.StoreData(assembleContextKey(context), data, Ttl)
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

// StoreContextToSet saves a context on set Cache
func (rc *RedisCache) StoreContextToSet(context Context) error {
	data, _ := json.Marshal(context)
	return rc.StoreDataToSet(assembleContextSetKey(context), data)
}

// RetrieveContextFromSet returns a context array from the Cache of user
func (rc *RedisCache) RetrieveContextFromSet(client, userID string) []Context {
	var redisContextArray []Context
	dataList, _ := rc.RetrieveDataFromSet(assembleContextSetKey(Context{UserID: userID, Client: client}))
	for _, data := range dataList {
		var redisContext Context
		json.Unmarshal([]byte(data), &redisContext)
		redisContextArray = append(redisContextArray, redisContext)
	}
	return redisContextArray
}

// CleanContextToDate clean context
func (rc *RedisCache) CleanContextToDate(client string, dateTime time.Time) error {
	keys, err := rc.GetAllKeysWithScanByMatch(fmt.Sprintf("%s:*:context", client), countScan)
	if err != nil {
		logrus.WithError(err).Error("Redis 'CleanContext'")
		return err
	}

	for _, key := range keys {
		dataList, _ := rc.RetrieveDataFromSet(key)
		for _, data := range dataList {
			var redisContext Context
			json.Unmarshal([]byte(data), &redisContext)
			if redisContext.Ttl.Before(dateTime) {
				_, err = rc.client.SRem(key, data).Result()
				if err != nil {
					logrus.Errorf("Could not delete member of set : %s", err.Error())
				}
			}
		}
	}
	return nil
}
