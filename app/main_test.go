package main

import (
	"context"
	"os"
	"testing"
	"time"

	"yalochat.com/salesforce-integration/base/cache"
)

func TestMain(t *testing.T) {
	os.Setenv("SALESFORCE-INTEGRATION_YALO_PASSWORD", "yaloPassword")
	os.Setenv("SALESFORCE-INTEGRATION_SALESFORCE_PASSWORD", "salesforcePassword")
	os.Setenv("SALESFORCE-INTEGRATION_SECRET_KEY", "secret")

	m, s := cache.CreateRedisServer()
	defer m.Close()
	defer s.Close()

	t.Run("Should initialize main without errors and Redis", func(t *testing.T) {
		go func() {
			time.Sleep(time.Second * 3)
			httpServer.Shutdown(context.Background())
		}()
		os.Setenv("SALESFORCE-INTEGRATION_REDIS_ADDRESS", m.Addr())

		main()
	})

	t.Run("Should initialize main without errors and Sentinel REDIS", func(t *testing.T) {
		go func() {
			time.Sleep(time.Second * 3)
			httpServer.Shutdown(context.Background())
		}()
		os.Setenv("SALESFORCE-INTEGRATION_REDIS_MASTER", m.Addr())
		os.Setenv("SALESFORCE-INTEGRATION_REDIS_SENTINEL", s.Addr())

		main()
	})
}
