package main

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestMain(t *testing.T) {
	os.Setenv("INTEGRATIONS-ADDON-SFSC_YALO_PASSWORD", "yaloPassword")
	os.Setenv("INTEGRATIONS-ADDON-SFSC_SECRET_KEY", "secret")

	t.Run("Should initialize main without errors", func(t *testing.T) {
		go func() {
			time.Sleep(time.Second * 3)
			httpServer.Shutdown(context.Background())
		}()

		main()
	})
}
