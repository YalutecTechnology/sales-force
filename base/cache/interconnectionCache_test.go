package cache

import (
	"reflect"
	"testing"
	"time"

	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
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

		err := rcs.StoreInterconnection(interconnection)

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
		rcs.StoreInterconnection(interconnectionExpected)

		actual, err := rcs.RetrieveInterconnection(Interconnection{
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
		expectedError := "redis: nil"

		_, err := rcs.RetrieveInterconnection(Interconnection{
			BotSlug: "aCode",
			BotID:   "botID",
		})

		if err.Error() != expectedError {
			t.Fatalf("Expected this error %v, but this was retrieved %v", expectedError, err)
		}
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
		rcs.StoreInterconnection(interconnection)
		rcs.StoreInterconnection(interconnection2)

		arrays := rcs.RetrieveAllInterconnections("client")

		if len(*arrays) != 2 {
			t.Fatalf("This was expected [2], but this was retrieved [%#v]", len(*arrays))
		}
		rcs.DeleteAll()
	})

	t.Run("Should retrieve interconnection without errors", func(t *testing.T) {
		c := redis.NewClient(&redis.Options{
			Addr: "127.0.0.1:10000",
		})
		rcs := &RedisCache{
			client: c,
		}

		arrays := rcs.RetrieveAllInterconnections("client")

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
	rcs.StoreInterconnection(interconnectionExpected)

	t.Run("Should delete all interconnections succesfully", func(t *testing.T) {
		err := rcs.DeleteAllInterconnections()
		if err != nil {
			t.Fatalf("Error should be nil, but this was retrieved %v", err)
		}
	})

	t.Run("Should fail when delete all interconnections", func(t *testing.T) {
		rcs.client.Close()
		err := rcs.DeleteAllInterconnections()
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
	rcs.StoreInterconnection(interconnectionsExpected)

	t.Run("Should delete interconnection succesfully", func(t *testing.T) {
		interconnectionToDelete := Interconnection{
			SessionID: interconnectionsExpected.SessionID,
			UserID:    interconnectionsExpected.UserID,
		}
		_, err := rcs.DeleteInterconnection(interconnectionToDelete)
		if err != nil {
			t.Fatalf("Error should be nil, but this was retrieved %v", err)
		}
	})

	t.Run("Should fail when delete all interconnection", func(t *testing.T) {
		interconnectionToDelete := Interconnection{
			Client: "client",
			UserID: "userId",
		}
		_, err := rcs.DeleteInterconnection(interconnectionToDelete)
		expectedErr := "Could not delete interconnection with key: client:userId:interconnection from Redis"
		if err.Error() != expectedErr {
			t.Fatalf("Error should be <%v>, but this was retrieved <%v>", expectedErr, err)
		}
	})
}

func TestRetrieveInterconnectionActiveByUserId(t *testing.T) {
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

	t.Run("Should retrieve a interconnection without errors", func(t *testing.T) {
		defer rcs.client.FlushAll()
		interconnectionExpected := Interconnection{
			BotID:         "botID",
			Client:        "client",
			BotSlug:       "coppel-bot",
			UserID:        "userID",
			Status:        "ON_HOLD",
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
		rcs.StoreInterconnection(interconnectionExpected)

		actual := rcs.RetrieveInterconnectionActiveByUserId(interconnectionExpected.UserID)

		assert.NotNil(t, actual)
		assert.Equal(t, &interconnectionExpected, actual)
	})

	t.Run("Should fail to retrieve a interconnection", func(t *testing.T) {
		defer rcs.client.FlushAll()
		interconnection := Interconnection{
			BotID:         "botID",
			BotSlug:       "coppel-bot",
			UserID:        "userID",
			Status:        "CLOSED",
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
		rcs.StoreInterconnection(interconnection)

		actual := rcs.RetrieveInterconnectionActiveByUserId(interconnection.UserID)

		assert.Nil(t, actual)
	})

	t.Run("Should retrieve a interconnection not found", func(t *testing.T) {
		defer rcs.client.FlushAll()
		interconnectionExpected := Interconnection{
			BotID:         "botID",
			BotSlug:       "coppel-bot",
			UserID:        "userID",
			Status:        "ON_HOLD",
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

		actual := rcs.RetrieveInterconnectionActiveByUserId(interconnectionExpected.UserID)

		assert.Nil(t, actual)
	})

	t.Run("Should retrieve a interconnection error connection", func(t *testing.T) {

		c := redis.NewClient(&redis.Options{
			Addr: "127.0.0.1:10000",
		})
		rcs := &RedisCache{
			client: c,
		}
		interconnectionExpected := Interconnection{
			BotID:         "botID",
			BotSlug:       "coppel-bot",
			UserID:        "userID",
			Status:        "ON_HOLD",
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

		actual := rcs.RetrieveInterconnectionActiveByUserId(interconnectionExpected.UserID)

		assert.Nil(t, actual)
	})
}
