package main

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestMain(t *testing.T) {
	os.Setenv("SALESFORCE-INTEGRATION_YALO_PASSWORD", "yaloPassword")
	os.Setenv("SALESFORCE-INTEGRATION_SECRET_KEY", "secret")

	t.Run("Should initialize main without errors", func(t *testing.T) {
		go func() {
			time.Sleep(time.Second * 3)
			httpServer.Shutdown(context.Background())
		}()

		main()
	})
}
