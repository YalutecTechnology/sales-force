package cache

import (
	"testing"
	"time"

	"github.com/go-redis/redis"
)

func TestNewRedisCache(t *testing.T) {
	m, s := CreateRedisServer()
	defer m.Close()
	defer s.Close()
	ttl := time.Second * 2

	t.Run("Should retrieve a RedisCache without errors", func(t *testing.T) {
		opts := &redis.FailoverOptions{
			MasterName:    s.MasterInfo().Name,
			SentinelAddrs: []string{s.Addr()},
		}
		expectedResult := "FailoverClient"

		rcs, err := NewRedisCache(&RedisOptions{
			FailOverOptions: opts,
			SessionsTTL:     ttl,
		})

		if err != nil {
			t.Fatalf("Expected nil error, but this was retrieved %v", err)
		}
		if rcs.client.Options().Addr != expectedResult {
			t.Fatalf("Expected client Addrs %s, but this is the client created %#v", expectedResult, rcs.client.Options())
		}
	})

	t.Run("Should retrieve a RedisCache with error", func(t *testing.T) {
		opts := &redis.FailoverOptions{
			MasterName:    s.MasterInfo().Name,
			SentinelAddrs: []string{},
		}
		expectedError := "all sentinels are unreachable"

		_, err := NewRedisCache(&RedisOptions{
			FailOverOptions: opts,
			SessionsTTL:     ttl,
		})

		if err == nil {
			t.Fatalf("Expected this error <%s>, but nil error was retrieved", expectedError)
		}
	})

	t.Run("Should fail getting redis client", func(t *testing.T) {
		_, err := NewRedisCache(&RedisOptions{
			SessionsTTL: ttl,
			Options:     nil,
		})
		if ErrGettingRedisClient != err {
			t.Fatalf("Expected this error %v, but nil error was retrieved %v", ErrGettingRedisClient, err)
		}
	})
}

func TestPing(t *testing.T) {
	m, s := CreateRedisServer()
	defer m.Close()
	defer s.Close()
	rcs := RedisCache{}

	t.Run("Should retrieve no error", func(t *testing.T) {
		c := redis.NewClient(&redis.Options{
			Addr: m.Addr(),
		})
		rcs.client = c
		err := rcs.ping()

		if err != nil {
			t.Fatalf("Expected nil error, but this error was retrieved %v", err)
		}
	})

	t.Run("Should retrieve a RedisCache with error", func(t *testing.T) {
		c := redis.NewClient(&redis.Options{
			Addr: "localhost:10000",
		})
		rcs.client = c
		expectedError := "dial tcp [::1]:6379: connect: connection refused"

		err := rcs.ping()

		if err == nil {
			t.Fatalf("Expected this error <%s>, but this error was retrieved %v", expectedError, err)
		}
	})
}

func TestStoreData(t *testing.T) {
	m, s := CreateRedisServer()
	defer m.Close()
	defer s.Close()

	t.Run("Should failed store data", func(t *testing.T) {
		c := redis.NewClient(&redis.Options{
			Addr: "127.0.0.1:10000",
		})
		rc := &RedisCache{
			client: c,
		}
		expectedErr := "dial tcp 127.0.0.1:10000: connect: connection refused"
		err := rc.StoreData("key-test", nil, time.Second)
		if err.Error() != expectedErr {
			t.Fatalf("Error should be %v, but this was retrieved %v", expectedErr, err)
		}
	})
}

func TestRetrieveData(t *testing.T) {
	m, s := CreateRedisServer()
	defer m.Close()
	defer s.Close()

	t.Run("Should failed getting data", func(t *testing.T) {
		c := redis.NewClient(&redis.Options{
			Addr: "127.0.0.1:10000",
		})
		rc := &RedisCache{
			client: c,
		}
		expectedErr := "dial tcp 127.0.0.1:10000: connect: connection refused"
		_, err := rc.RetrieveData("key-test")
		if err.Error() != expectedErr {
			t.Fatalf("Error should be %v, but this was retrieved %v", expectedErr, err)
		}
	})
}
