package cache

import (
	"reflect"
	"testing"
	"time"

	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
	"yalochat.com/salesforce-integration/base/constants"
)

func TestStoreInterconnection(t *testing.T) {
	m, s := CreateRedisServer()
	defer m.Close()
	defer s.Close()
	ttl := time.Second * 2
	opts := &RedisOptions{
		FailOverOptions: &redis.FailoverOptions{
			MasterName:    s.MasterInfo().Name,
			SentinelAddrs: []string{s.Addr()},
		},
		SessionsTTL: ttl,
	}
	rcs, _ := NewRedisCache(opts)

	cache := NewInterconnectionCache(rcs)

	t.Run("Should store a interconnection without errors", func(t *testing.T) {
		interconnection := Interconnection{
			BotID:         "botID",
			BotSlug:       "coppel-bot",
			UserID:        "userID",
			Status:        "status",
			SessionID:     "session",
			SessionKey:    "sessionID",
			AffinityToken: "affinityToken",
			Timestamp:     time.Time{},
			Provider:      "provider",
			Name:          "name",
			Email:         "email",
			PhoneNumber:   "55555555555",
			CaseID:        "caseID",
			ExtraData: map[string]interface{}{
				"data": "data",
			},
		}

		err := cache.StoreInterconnection(interconnection)

		if err != nil {
			t.Fatalf("Error should be nil, but this was retrieved %v", err)
		}
	})
}

func TestRetrieveInterconnection(t *testing.T) {
	m, s := CreateRedisServer()
	defer m.Close()
	defer s.Close()
	ttl := time.Second * 2
	opts := &RedisOptions{
		FailOverOptions: &redis.FailoverOptions{
			MasterName:    s.MasterInfo().Name,
			SentinelAddrs: []string{s.Addr()},
		},
		SessionsTTL: ttl,
	}
	rcs, _ := NewRedisCache(opts)

	cache := NewInterconnectionCache(rcs)

	t.Run("Should retrieve a interconnection without errors", func(t *testing.T) {
		interconnectionExpected := Interconnection{
			BotID:         "botID",
			BotSlug:       "coppel-bot",
			UserID:        "userID",
			Status:        "status",
			SessionID:     "session",
			SessionKey:    "sessionID",
			AffinityToken: "affinityToken",
			Timestamp:     time.Time{},
			Provider:      "provider",
			Name:          "name",
			Email:         "email",
			PhoneNumber:   "55555555555",
			CaseID:        "caseID",
			ExtraData: map[string]interface{}{
				"data": "data",
			},
		}
		cache.StoreInterconnection(interconnectionExpected)

		actual, err := cache.RetrieveInterconnection(Interconnection{
			SessionID: interconnectionExpected.SessionID,
			UserID:    interconnectionExpected.UserID,
		})

		if err != nil {
			t.Fatalf("Error should be nil, but this was retrieved %v", err)
		}
		if !reflect.DeepEqual(&interconnectionExpected, actual) {
			t.Fatalf("This was expected %#v, but this was retrieved %#v", interconnectionExpected, actual)
		}
	})

	t.Run("Should fail to retrieve a not stored interconnection", func(t *testing.T) {
		_, err := cache.RetrieveInterconnection(Interconnection{
			BotSlug: "aCode",
			BotID:   "botID",
		})

		if err.Error() != constants.ErrInterconnectionNotFound.Error() {
			t.Fatalf("Expected this error %v, but this was retrieved %v", constants.ErrInterconnectionNotFound.Error(), err)
		}
	})

	t.Run("Should retrieve interconnection without errors", func(t *testing.T) {
		c := redis.NewClient(&redis.Options{
			Addr: "127.0.0.1:10000",
		})
		rcs := &RedisCache{
			client: c,
		}
		cache := NewInterconnectionCache(rcs)

		arrays, err := cache.RetrieveInterconnection(Interconnection{})

		assert.Nil(t, arrays)
		assert.Error(t, err)
	})
}

func TestRetrieveInterconnections(t *testing.T) {
	m, s := CreateRedisServer()
	defer m.Close()
	defer s.Close()
	ttl := time.Second * 2
	opts := &RedisOptions{
		FailOverOptions: &redis.FailoverOptions{
			MasterName:    s.MasterInfo().Name,
			SentinelAddrs: []string{s.Addr()},
		},
		SessionsTTL: ttl,
	}
	rcs, _ := NewRedisCache(opts)

	cache := NewInterconnectionCache(rcs)

	t.Run("Should retrieve interconnection without errors", func(t *testing.T) {
		interconnection := Interconnection{
			BotID:         "botID",
			BotSlug:       "coppel-bot",
			Client:        "client",
			UserID:        "userID",
			Status:        "status",
			SessionID:     "session",
			SessionKey:    "sessionID",
			AffinityToken: "affinityToken",
			Timestamp:     time.Time{},
			Provider:      "provider",
			Name:          "name",
			Email:         "email",
			PhoneNumber:   "55555555555",
			CaseID:        "caseID",
			ExtraData: map[string]interface{}{
				"data": "data",
			},
		}
		interconnection2 := Interconnection{
			BotID:         "botID2",
			BotSlug:       "m2-seller-2",
			Client:        "client",
			UserID:        "userID2",
			Status:        "status",
			SessionID:     "session",
			SessionKey:    "sessionID",
			AffinityToken: "affinityToken",
			Timestamp:     time.Time{},
			Provider:      "provider",
			Name:          "name",
			Email:         "email",
			PhoneNumber:   "55555555555",
			CaseID:        "caseID",
			ExtraData: map[string]interface{}{
				"data": "data",
			},
		}
		cache.StoreInterconnection(interconnection)
		cache.StoreInterconnection(interconnection2)

		arrays := cache.RetrieveAllInterconnections("client")

		if len(*arrays) != 2 {
			t.Fatalf("This was expected [2], but this was retrieved [%#v]", len(*arrays))
		}
		cache.cache.DeleteAll()
	})

	t.Run("Should retrieve interconnection without errors", func(t *testing.T) {
		c := redis.NewClient(&redis.Options{
			Addr: "127.0.0.1:10000",
		})
		rcs := &RedisCache{
			client: c,
		}
		cache := NewInterconnectionCache(rcs)

		arrays := cache.RetrieveAllInterconnections("client")

		assert.Nil(t, arrays)
	})
}

func TestDeleteAllInterconnections(t *testing.T) {
	m, s := CreateRedisServer()
	defer m.Close()
	defer s.Close()
	ttl := time.Second * 2
	opts := &RedisOptions{
		FailOverOptions: &redis.FailoverOptions{
			MasterName:    s.MasterInfo().Name,
			SentinelAddrs: []string{s.Addr()},
		},
		SessionsTTL: ttl,
	}
	rcs, _ := NewRedisCache(opts)

	cache := NewInterconnectionCache(rcs)

	interconnectionExpected := Interconnection{
		BotID:         "botID2",
		BotSlug:       "m2-seller-2",
		UserID:        "userID2",
		Status:        "status",
		SessionID:     "sessionID",
		SessionKey:    "sessionKey",
		AffinityToken: "affinityToken",
		Timestamp:     time.Time{},
		Provider:      "provider",
		Name:          "name",
		Email:         "email",
		PhoneNumber:   "55555555555",
		CaseID:        "caseID",
		ExtraData: map[string]interface{}{
			"data": "data",
		},
	}
	cache.StoreInterconnection(interconnectionExpected)

	t.Run("Should delete all interconnections succesfully", func(t *testing.T) {
		err := cache.DeleteAllInterconnections()
		if err != nil {
			t.Fatalf("Error should be nil, but this was retrieved %v", err)
		}
	})

	t.Run("Should fail when delete all interconnections", func(t *testing.T) {
		rcs.client.Close()
		err := cache.DeleteAllInterconnections()
		expectedErr := "redis: client is closed"
		if err.Error() != expectedErr {
			t.Fatalf("Error should be <%v>, but this was retrieved <%v>", expectedErr, err)
		}
	})
}

func TestDeleteInterconnection(t *testing.T) {
	m, s := CreateRedisServer()
	defer m.Close()
	defer s.Close()
	ttl := time.Second * 2
	opts := &RedisOptions{
		FailOverOptions: &redis.FailoverOptions{
			MasterName:    s.MasterInfo().Name,
			SentinelAddrs: []string{s.Addr()},
		},
		SessionsTTL: ttl,
	}
	rcs, _ := NewRedisCache(opts)

	cache := NewInterconnectionCache(rcs)

	interconnectionsExpected := Interconnection{
		BotID:         "botID2",
		BotSlug:       "m2-seller-2",
		UserID:        "userID2",
		Status:        "status",
		SessionID:     "sessionID",
		SessionKey:    "sessionKey",
		AffinityToken: "affinityToken",
		Timestamp:     time.Time{},
		Provider:      "provider",
		Name:          "name",
		Email:         "email",
		PhoneNumber:   "55555555555",
		CaseID:        "caseID",
		ExtraData: map[string]interface{}{
			"data": "data",
		},
	}
	cache.StoreInterconnection(interconnectionsExpected)

	t.Run("Should delete interconnection succesfully", func(t *testing.T) {
		interconnectionToDelete := Interconnection{
			SessionID: interconnectionsExpected.SessionID,
			UserID:    interconnectionsExpected.UserID,
		}
		_, err := cache.DeleteInterconnection(interconnectionToDelete)
		if err != nil {
			t.Fatalf("Error should be nil, but this was retrieved %v", err)
		}
	})

	t.Run("Should fail when delete all interconnection", func(t *testing.T) {
		interconnectionToDelete := Interconnection{
			Client: "client",
			UserID: "userId",
		}
		_, err := cache.DeleteInterconnection(interconnectionToDelete)
		expectedErr := "Could not delete interconnection with key: client:userId:interconnection from Redis"
		if err.Error() != expectedErr {
			t.Fatalf("Error should be <%v>, but this was retrieved <%v>", expectedErr, err)
		}
	})
}
