package main

import (
	"context"
	"os"
	"testing"
	"time"

	"yalochat.com/salesforce-integration/base/cache"
)

func TestMain_Main(t *testing.T) {
	os.Setenv("SALESFORCE-INTEGRATION_YALO_PASSWORD", "yaloPassword")
	os.Setenv("SALESFORCE-INTEGRATION_SALESFORCE_PASSWORD", "salesforcePassword")
	os.Setenv("SALESFORCE-INTEGRATION_SFC_SOURCE_FLOW_BOT", "default={\"subject\":\"Asunto por defecto\",\"providers\":{\"whatsapp\":{\"button_id\":\"buttonId\",\"owner_id\":\"oownerId\"},\"facebook\":{\"button_id\":\"buttonId\",\"owner_id\":\"ownerID\"}}}")
	os.Setenv("SALESFORCE-INTEGRATION_SECRET_KEY", "secret")
	os.Setenv("SALESFORCE-INTEGRATION_MESSAGES", `{"waitAgent":"Esperando un agente","WelcomeTemplate":"Hola soy %s y necesito ayuda","Context":"Contexto","DescriptionCase":"Caso levantado por el Bot","UploadImageError":"Imagen no enviada","UploadImageSuccess":"**El usuario adjunto una imagen al caso**","queuePosition":"Posición en la cola","waitTime":"Tiempo de espera","botLabel":"Bot","clientLabel":"Cliente"}`)

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
