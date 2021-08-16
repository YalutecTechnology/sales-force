package manage

import (
	"reflect"
	"testing"
	"time"

	"github.com/go-redis/redis"
	"yalochat.com/salesforce-integration/base/cache"
)

func TestCreateManager(t *testing.T) {
	m, s := cache.CreateRedisServer()
	defer m.Close()
	defer s.Close()
	t.Run("Should retrieve a manager instance", func(t *testing.T) {
		expected := &Manager{
			clientName: "salesforce-integration",
		}
		config := &ManagerOptions{
			AppName: "salesforce-integration",
			RedisOptions: cache.RedisOptions{
				FailOverOptions: &redis.FailoverOptions{
					MasterName:    s.MasterInfo().Name,
					SentinelAddrs: []string{s.Addr()},
				},
				SessionsTTL: time.Second,
			},
		}
		actual := CreateManager(config)

		if !reflect.DeepEqual(expected, actual) {
			t.Fatalf("Expected this %#v, but retrieved this %#v", expected, actual)
		}
	})
}
