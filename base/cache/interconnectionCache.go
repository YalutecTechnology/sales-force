package cache

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
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

// InterconnectionCache interface that holds method to retrieve interconnections sessions from redis cache
type InterconnectionCache interface {
	StoreInterconnection(Interconnection) error
	RetrieveInterconnection(Interconnection) (*Interconnection, error)
	DeleteAllInterconnections() error
	DeleteInterconnection(Interconnection) (bool, error)
	RetrieveAllInterconnections(client string) *[]Interconnection
	RetrieveInterconnectionActiveByUserId(userId string) *Interconnection
}

// assembleKey retrive key by template
func assembleKey(interconnection Interconnection) string {
	return fmt.Sprintf(interconnectionKeyTemplate, interconnection.Client, interconnection.UserID)
}

// StoreInterconnection saves a interconnection on Cache
func (rc *RedisCache) StoreInterconnection(interconnection Interconnection) error {
	data, _ := json.Marshal(interconnection)
	return rc.StoreData(assembleKey(interconnection), data, ttl)
}

// RetrieveInterconnection returns a interconnection from the Cache
func (rc *RedisCache) RetrieveInterconnection(interconnection Interconnection) (*Interconnection, error) {

	var redisInterconnection Interconnection
	data, err := rc.RetrieveData(assembleKey(interconnection))
	if err != nil {
		return nil, err
	}
	json.Unmarshal([]byte(data), &redisInterconnection)
	return &redisInterconnection, nil
}

// RetrieveAllInterconnections returns interconnections array from the Cache
func (rc *RedisCache) RetrieveAllInterconnections(client string) *[]Interconnection {
	var redisInterconnectionsArray []Interconnection
	keys, err := rc.GetAllKeysWithScanByMatch(fmt.Sprintf("%s:*:interconnection", client), countScan)
	if err != nil {
		logrus.WithError(err).Error("Redis 'RetrieveAllInterconnections'")
		return nil
	}
	for _, key := range keys {
		var redisInterconnections Interconnection
		data, _ := rc.RetrieveData(key)
		json.Unmarshal([]byte(data), &redisInterconnections)
		redisInterconnectionsArray = append(redisInterconnectionsArray, redisInterconnections)
	}
	return &redisInterconnectionsArray
}

//TODO remove method
// RetrieveInterconnectionActiveByUserId returns interconnection from the Cache with status OnHold or Active
func (rc *RedisCache) RetrieveInterconnectionActiveByUserId(userID string) *Interconnection {
	var redisInterconnection Interconnection
	var err error
	var keys []string
	cursor := uint64(0)

	for {
		keys, cursor, err = rc.ScanKeys(cursor, fmt.Sprintf("*:%s:interconnection", userID), countScan)
		if err != nil {
			logrus.WithError(err).Error("Redis 'RetrieveInterconnectionActiveByUserId'")
			return nil
		}

		var value string
		for i := range keys {
			value, err = rc.RetrieveData(keys[i])
			if err != nil {
				logrus.WithError(err).Error("error RetrieveData")
				return nil
			}
			err = json.Unmarshal([]byte(value), &redisInterconnection)
			if redisInterconnection.UserID == userID {
				if redisInterconnection.Status == "ON_HOLD" || redisInterconnection.Status == "ACTIVE" {
					return &redisInterconnection
				}
				return nil
			}
		}
		if cursor == 0 {
			break
		}
	}
	logrus.Errorf("Interconnection not found in redis with userID : [%s], Description: %s",
		userID, err)
	return nil
}

// DeleteAllInterconnections delete all Interconnections from Cache
func (rc *RedisCache) DeleteAllInterconnections() error {
	return rc.DeleteAll()
}

// DeleteInterconnection delete a Interconnection from Cache
func (rc *RedisCache) DeleteInterconnection(interconnection Interconnection) (bool, error) {
	interconnectionRedisKey := assembleKey(interconnection)
	data := rc.client.Del(interconnectionRedisKey)
	if data.Val() != 1 {
		return false, fmt.Errorf(fmt.Sprintf(deleteRedisError, interconnectionRedisKey))
	}
	return true, nil
}
