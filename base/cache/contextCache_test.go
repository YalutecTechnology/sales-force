package cache

import (
	"testing"
	"time"

	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
)

func TestRedisCache_StoreContext(t *testing.T) {
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

	cache := NewContextCache(rcs)

	t.Run("Should store a context without errors", func(t *testing.T) {
		context := Context{
			UserID:    "55555555555",
			Client:    "client",
			Timestamp: 1630073244,
			URL:       "uri",
			MIMEType:  "mime",
			Caption:   "document",
			Text:      "message",
			From:      "user",
		}

		err := cache.StoreContext(context)

		if err != nil {
			t.Fatalf("Error should be nil, but this was retrieved %v", err)
		}
	})
}

func TestRedisCache_RetrieveContext(t *testing.T) {
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

	cache := NewContextCache(rcs)

	defer rcs.client.FlushAll()

	t.Run("Should retrieve line windows without errors", func(t *testing.T) {
		context := Context{
			UserID:    "55555555555",
			Client:    "client",
			Timestamp: 1630073244,
			URL:       "uri",
			MIMEType:  "mime",
			Caption:   "document",
			Text:      "message",
			From:      "user",
		}
		context2 := Context{
			UserID:    "55555555555",
			Timestamp: 1630073242,
			URL:       "uri",
			MIMEType:  "mime",
			Caption:   "document",
			Text:      "message",
			From:      "bot",
		}
		context3 := Context{
			UserID:    "55555555555",
			Timestamp: 1630073243,
			URL:       "uri",
			MIMEType:  "mime",
			Caption:   "document",
			Text:      "message",
			From:      "user",
		}
		cache.StoreContext(context)
		cache.StoreContext(context2)
		cache.StoreContext(context3)

		arrays := cache.RetrieveContext(context.UserID)

		assert.Equal(t, &[]Context{context2, context3, context}, &arrays)
	})

	t.Run("Should retrieve line windows with error", func(t *testing.T) {
		c := redis.NewClient(&redis.Options{
			Addr: "127.0.0.1:10000",
		})
		rcs := &RedisCache{
			client: c,
		}

		cache := NewContextCache(rcs)

		arrays := cache.RetrieveContext("55555555555")

		assert.Empty(t, &arrays)
	})
}

func TestRedisCache_StoreContextToSet(t *testing.T) {
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

	cache := NewContextCache(rcs)

	t.Run("Should store a context to set without errors", func(t *testing.T) {
		context := Context{
			UserID:    "55555555555",
			Client:    "client",
			Timestamp: 1630073244,
			URL:       "uri",
			MIMEType:  "mime",
			Caption:   "document",
			Text:      "message",
			From:      "user",
		}

		context2 := Context{
			UserID:    "55555555555",
			Client:    "client",
			Timestamp: 1630073244,
			URL:       "uri",
			MIMEType:  "mime",
			Caption:   "document",
			Text:      "message",
			From:      "user",
		}

		err := cache.StoreContextToSet(context)

		if err != nil {
			t.Fatalf("Error should be nil, but this was retrieved %v", err)
		}

		err = cache.StoreContextToSet(context2)

		if err != nil {
			t.Fatalf("Error should be nil, but this was retrieved %v", err)
		}
	})
}

func TestRedisCache_RetrieveContextFromSet(t *testing.T) {
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

	cache := NewContextCache(rcs)

	defer rcs.client.FlushAll()

	t.Run("Should retrieve contexts without errors", func(t *testing.T) {
		context := Context{
			UserID:    "55555555555",
			Client:    "client",
			Timestamp: 1630073244,
			URL:       "uri",
			MIMEType:  "mime",
			Caption:   "document",
			Text:      "message",
			From:      "user",
		}
		context2 := Context{
			UserID:    "55555555555",
			Timestamp: 1630073242,
			Client:    "client",
			URL:       "uri",
			MIMEType:  "mime",
			Caption:   "document",
			Text:      "message",
			From:      "bot",
		}
		context3 := Context{
			UserID:    "55555555555",
			Client:    "client",
			Timestamp: 1630073243,
			URL:       "uri",
			MIMEType:  "mime",
			Caption:   "document",
			Text:      "message",
			From:      "user",
		}
		cache.StoreContextToSet(context)
		cache.StoreContextToSet(context2)
		cache.StoreContextToSet(context3)

		arrays := cache.RetrieveContextFromSet(context.Client, context.UserID)

		assert.Equal(t, 3, len(arrays))
		assert.Equal(t, &[]Context{context2, context3, context}, &arrays)
	})

	t.Run("Should retrieve context with error", func(t *testing.T) {
		c := redis.NewClient(&redis.Options{
			Addr: "127.0.0.1:10000",
		})
		rcs := &RedisCache{
			client: c,
		}
		cache := NewContextCache(rcs)

		arrays := cache.RetrieveContextFromSet("client", "55555555555")

		assert.Empty(t, &arrays)
	})
}

func TestRedisCache_CleanContextToDate(t *testing.T) {
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

	cache := NewContextCache(rcs)

	t.Run("Should delete a context set without errors", func(t *testing.T) {
		context := Context{
			UserID:    "55555555555",
			Client:    "client",
			Timestamp: 1630073244,
			URL:       "uri",
			MIMEType:  "mime",
			Caption:   "document",
			Text:      "message",
			From:      "user",
			Ttl:       time.Now().Add(2 * time.Minute * -1),
		}

		context2 := Context{
			UserID:    "55555555555",
			Client:    "client",
			Timestamp: 1630073243,
			URL:       "uri",
			MIMEType:  "mime",
			Caption:   "document",
			Text:      "message",
			From:      "user",
			Ttl:       time.Now().Add(2 * time.Minute * -1),
		}

		cache.StoreContextToSet(context)
		cache.StoreContextToSet(context2)
		err := cache.CleanContextToDate(context.Client, time.Now())
		assert.NoError(t, err)
		elements := cache.RetrieveContextFromSet(context.Client, context.UserID)
		assert.Empty(t, elements)
	})

	t.Run("Should CleanContextToDate with error", func(t *testing.T) {
		c := redis.NewClient(&redis.Options{
			Addr: "127.0.0.1:10000",
		})
		rcs := &RedisCache{
			client: c,
		}

		cache := NewContextCache(rcs)
		err := cache.CleanContextToDate("client", time.Now())
		assert.Error(t, err)
	})
}
