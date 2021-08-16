package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	ddrouter "gopkg.in/DataDog/dd-trace-go.v1/contrib/julienschmidt/httprouter"
	"yalochat.com/salesforce-integration/app/manage"
)

func TestWelcomeAPI(t *testing.T) {
	handler := ddrouter.New(ddrouter.WithServiceName("app.http"))
	API(handler, &manage.ManagerOptions{
		AppName: "AppName",
	}, ApiConfig{})

	t.Run("Should return a http response OK", func(t *testing.T) {
		requestURL := "/v1/welcome"
		req, _ := http.NewRequest("GET", requestURL, nil)
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, req)

		if response.Code != http.StatusOK {
			t.Errorf("Response should be %v, but it answer with %v ", http.StatusOK, response.Code)
		}
	})
}
