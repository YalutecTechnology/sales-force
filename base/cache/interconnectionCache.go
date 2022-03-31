package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
	"yalochat.com/salesforce-integration/base/constants"
)

// InterconnectionStatus contains interconnection status to match with InterconnectionStatus
type InterconnectionStatus int

const (
	deleteRedisError           = "Could not delete interconnection with key: %s from Redis"
	interconnectionKeyTemplate = "%s:%s:interconnection"
)

// Interconnection is a struct that defines the interconnection session related to a conversation between agent in Salesforce and user in the bot
type Interconnection struct {
	Client        string                 `json:"client"`
	UserID        string                 `json:"userID"`
	SessionID     string                 `json:"sessionID"`
	SessionKey    string                 `json:"sessionKey"`
	AffinityToken string                 `json:"affinityToken"`
	Status        string                 `json:"status"`
	Timestamp     time.Time              `json:"timestamp"`
	Provider      string                 `json:"provider"`
	BotSlug       string                 `json:"botSlug"`
	BotID         string                 `json:"botID"`
	Name          string                 `json:"name"`
	Email         string                 `json:"email"`
	PhoneNumber   string                 `json:"phoneNumber"`
	CaseID        string                 `json:"caseID"`
	ExtraData     map[string]interface{} `json:"extraData"`
}

type InterconnectionCache struct {
	cache *RedisCache
}

func NewInterconnectionCache(cache *RedisCache) *InterconnectionCache {
	return &InterconnectionCache{cache: cache}
}

// IInterconnectionCache interface that holds method to retrieve interconnections sessions from redis cache
type IInterconnectionCache interface {
	StoreInterconnection(Interconnection) error
	RetrieveInterconnection(Interconnection) (*Interconnection, error)
	DeleteAllInterconnections() error
	DeleteInterconnection(Interconnection) (bool, error)
	RetrieveAllInterconnections(client string) *[]Interconnection
}

// assembleKey retrive key by template
func assembleKey(interconnection Interconnection) string {
	return fmt.Sprintf(interconnectionKeyTemplate, interconnection.Client, interconnection.UserID)
}

// StoreInterconnection saves a interconnection on Cache
func (rc *InterconnectionCache) StoreInterconnection(interconnection Interconnection) error {
	data, _ := json.Marshal(interconnection)
	return rc.cache.StoreData(assembleKey(interconnection), data, Ttl)
}

// RetrieveInterconnection returns a interconnection from the Cache
func (rc *InterconnectionCache) RetrieveInterconnection(interconnection Interconnection) (*Interconnection, error) {

	var redisInterconnection Interconnection
	data, err := rc.cache.RetrieveData(assembleKey(interconnection))
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, constants.ErrInterconnectionNotFound
		}
		return nil, err
	}
	json.Unmarshal([]byte(data), &redisInterconnection)
	return &redisInterconnection, nil
}

// RetrieveAllInterconnections returns interconnections array from the Cache
func (rc *InterconnectionCache) RetrieveAllInterconnections(client string) *[]Interconnection {
	var redisInterconnectionsArray []Interconnection
	keys, err := rc.cache.GetAllKeysWithScanByMatch(fmt.Sprintf("%s:*:interconnection", client), countScan)
	if err != nil {
		logrus.WithError(err).Error("Redis 'RetrieveAllInterconnections'")
		return nil
	}
	for _, key := range keys {
		var redisInterconnections Interconnection
		data, _ := rc.cache.RetrieveData(key)
		json.Unmarshal([]byte(data), &redisInterconnections)
		redisInterconnectionsArray = append(redisInterconnectionsArray, redisInterconnections)
	}
	return &redisInterconnectionsArray
}

// DeleteAllInterconnections delete all Interconnections from Cache
func (rc *InterconnectionCache) DeleteAllInterconnections() error {
	return rc.cache.DeleteAll()
}

// DeleteInterconnection delete a Interconnection from Cache
func (rc *InterconnectionCache) DeleteInterconnection(interconnection Interconnection) (bool, error) {
	interconnectionRedisKey := assembleKey(interconnection)
	data := rc.cache.client.Del(interconnectionRedisKey)
	if data.Val() != 1 {
		return false, fmt.Errorf(fmt.Sprintf(deleteRedisError, interconnectionRedisKey))
	}
	return true, nil
}
