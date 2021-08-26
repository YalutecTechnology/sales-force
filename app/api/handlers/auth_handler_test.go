package handlers

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dgrijalva/jwt-go"
	ddrouter "gopkg.in/DataDog/dd-trace-go.v1/contrib/julienschmidt/httprouter"
	"yalochat.com/salesforce-integration/app/manage"
	"yalochat.com/salesforce-integration/base/helpers"
)

const (
	yaloUserTest     = "yaloUser"
	yaloPasswordTest = "yaloPassword"
	yaloTokenTest    = "eyJhbGciOiJIUzM4NCIsInR5cCI6IkpXVCJ9.eyJwYXNzd29yZCI6IjEyYjgxNmM2YjhhM2M0NWVkYmUwZjg1ZDhiYTQ0YjNkNWYxNDhhYjkiLCJyb2xlIjoiWUFMT19ST0xFIiwidXNlcm5hbWUiOiJ5YWxvVXNlciJ9.2HsCTU6IPwlFQV_Vu8w_IjxtLJIQs-_W-8jobwWtGmkzc9SYaibc7QN-5caZgVrC"
	secretTest       = "secret"
)

var apiConfig = ApiConfig{
	YaloUsername: yaloUserTest,
	YaloPassword: yaloPasswordTest,
	SecretKey:    secretTest,
}

func TestAuthenticate(t *testing.T) {
	handler := ddrouter.New(ddrouter.WithServiceName("appName.http"))
	managerOptions := &manage.ManagerOptions{AppName: "appName"}
	API(handler, managerOptions, apiConfig)

	t.Run("Should return an yalo token", func(t *testing.T) {
		requestURL := "/v1/authenticate"
		req, _ := http.NewRequest("POST", requestURL, bytes.NewBuffer([]byte("{\"username\":\"yaloUser\",\"password\":\"yaloPassword\"}")))
		response := httptest.NewRecorder()
		expected := yaloTokenTest

		handler.ServeHTTP(response, req)
		if response.Code != http.StatusOK {
			t.Errorf("Response should be %v, but it answer with %v ", http.StatusOK, response.Code)
		}
		var token tokenResult
		err := json.NewDecoder(response.Body).Decode(&token)
		if err != nil {
			t.Fatalf("Response should produce an token result, but found this: %s", response.Body)
		}

		if token.Token != expected {
			t.Fatalf("Response should produce an token %s, but found this: %s", expected, token.Token)
		}

	})

	t.Run("Should fail when parsing body of post", func(t *testing.T) {
		requestURL := "/v1/authenticate"
		req, _ := http.NewRequest("POST", requestURL, bytes.NewBuffer([]byte("Invalid body")))
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, req)

		if response.Code != http.StatusBadRequest {
			t.Errorf("Response should be %v, but it answer with %v ", http.StatusBadRequest, response.Code)
		}

		var failResponse helpers.FailedResponse
		if err := json.Unmarshal(response.Body.Bytes(), &failResponse); err != nil {
			t.Fatal("Error parsing response")
		}
		if !strings.Contains(failResponse.ErrorDescription, helpers.InvalidPayload) {
			t.Fatalf("Message should be contains <%s>, but this was found <%s>", helpers.InvalidPayload, failResponse.ErrorDescription)
		}
	})

	t.Run("Should fail to send invalid username", func(t *testing.T) {
		requestURL := "/v1/authenticate"
		req, _ := http.NewRequest("POST", requestURL, bytes.NewBuffer([]byte("{\"username\":\"otherUser\",\"password\":\"password\"}")))
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, req)

		if response.Code != http.StatusUnauthorized {
			t.Errorf("Response should be %v, but it answer with %v ", http.StatusUnauthorized, response.Code)
		}

		var failResponse helpers.FailedResponse
		if err := json.Unmarshal(response.Body.Bytes(), &failResponse); err != nil {
			t.Fatal("Error parsing response")
		}
		if !strings.Contains(failResponse.ErrorDescription, "username Invalid") {
			t.Fatalf("Message should be contains <%s>, but this was found <%s>", "username Invalid", failResponse.ErrorDescription)
		}
	})

	t.Run("Should fail to send invalid password", func(t *testing.T) {
		requestURL := "/v1/authenticate"
		req, _ := http.NewRequest("POST", requestURL, bytes.NewBuffer([]byte("{\"username\":\"yaloUser\",\"password\":\"otherPassword\"}")))
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, req)

		if response.Code != http.StatusUnauthorized {
			t.Errorf("Response should be %v, but it answer with %v ", http.StatusUnauthorized, response.Code)
		}

		var failResponse helpers.FailedResponse
		if err := json.Unmarshal(response.Body.Bytes(), &failResponse); err != nil {
			t.Fatal("Error parsing response")
		}
		if !strings.Contains(failResponse.ErrorDescription, "Invalid credentials.") {
			t.Fatalf("Message should be contains <%s>, but this was found <%s>", "Invalid credentials.", failResponse.ErrorDescription)
		}
	})

	t.Run("Should fail to sign a token", func(t *testing.T) {
		singMethod = &jwt.SigningMethodECDSA{}
		handler = ddrouter.New(ddrouter.WithServiceName("appName.http"))
		managerOptions := &manage.ManagerOptions{AppName: "appName"}
		API(handler, managerOptions, apiConfig)
		requestURL := "/v1/authenticate"
		req, _ := http.NewRequest("POST", requestURL, bytes.NewBuffer([]byte("{\"username\":\"yaloUser\",\"password\":\"yaloPassword\"}")))
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, req)
		if response.Code != http.StatusInternalServerError {
			t.Errorf("Response should be %v, but it answer with %v ", http.StatusInternalServerError, response.Code)
		}

		var failResponse helpers.FailedResponse
		if err := json.Unmarshal(response.Body.Bytes(), &failResponse); err != nil {
			t.Fatal("Error parsing response")
		}
		if !strings.Contains(failResponse.ErrorDescription, "key is of invalid type") {
			t.Fatalf("Message should be contains <%s>, but this was found <%s>", "key is of invalid type", failResponse.ErrorDescription)
		}

		singMethod = jwt.SigningMethodHS384
	})

}

func TestProtected(t *testing.T) {
	handler := ddrouter.New(ddrouter.WithServiceName("appName.http"))
	managerOptions := &manage.ManagerOptions{AppName: "appName"}
	API(handler, managerOptions, apiConfig)

	t.Run("Should return a user ", func(t *testing.T) {
		requestURL := "/v1/tokens/check"
		req, _ := http.NewRequest("GET", requestURL, nil)
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", yaloTokenTest))
		response := httptest.NewRecorder()
		expected := AuthUser{Username: yaloUserTest, Role: Yalo}

		handler.ServeHTTP(response, req)
		if response.Code != http.StatusOK {
			t.Errorf("Response should be %v, but it answer with %v ", http.StatusOK, response.Code)
		}
		var user AuthUser
		err := json.NewDecoder(response.Body).Decode(&user)
		if err != nil {
			t.Fatalf("Response should produce an auth user, but found this: %s", response.Body)
		}

		if user != expected {
			t.Fatalf("Response should produce an auth user %s, but found this: %s", expected, user)
		}

	})
}

func TestAuthorizeMiddlewareByHeader(t *testing.T) {
	handler := ddrouter.New(ddrouter.WithServiceName("appName.http"))
	managerOptions := &manage.ManagerOptions{AppName: "appName"}
	API(handler, managerOptions, apiConfig)

	t.Run("Should respond successfully with Yalo token test", func(t *testing.T) {
		requestURL := "/v1/tokens/check"
		req, _ := http.NewRequest("GET", requestURL, nil)
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", yaloTokenTest))
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, req)
		if response.Code != http.StatusOK {
			t.Errorf("Response should be %v, but it answer with %v ", http.StatusOK, response.Code)
		}

	})

	t.Run("Should fail for missing word Bearer", func(t *testing.T) {
		requestURL := "/v1/tokens/check"
		req, _ := http.NewRequest("GET", requestURL, nil)
		req.Header.Add("Authorization", yaloTokenTest)
		response := httptest.NewRecorder()
		expectedErrorMessage := "Invalid Authorization token, header invalid."

		handler.ServeHTTP(response, req)
		if response.Code != http.StatusUnauthorized {
			t.Errorf("Response should be %v, but it answer with %v ", http.StatusUnauthorized, response.Code)
		}
		var result helpers.FailedResponse
		err := json.NewDecoder(response.Body).Decode(&result)
		if err != nil {
			t.Fatalf("Response should produce a failed response, but found this: %s", response.Body)
		}

		if result.ErrorDescription != expectedErrorMessage {
			t.Fatalf("Response should produce a message %s, but found this: %s", expectedErrorMessage, result.ErrorDescription)
		}

	})

	t.Run("Should fail for not sending header", func(t *testing.T) {
		requestURL := "/v1/tokens/check"
		req, _ := http.NewRequest("GET", requestURL, nil)
		response := httptest.NewRecorder()
		expectedErrorMessage := "An Authorization header or an Query Param with name token is required."

		handler.ServeHTTP(response, req)
		if response.Code != http.StatusUnauthorized {
			t.Errorf("Response should be %v, but it answer with %v ", http.StatusUnauthorized, response.Code)
		}
		var result helpers.FailedResponse
		err := json.NewDecoder(response.Body).Decode(&result)
		if err != nil {
			t.Fatalf("Response should produce a failed response, but found this: %s", response.Body)
		}

		if result.ErrorDescription != expectedErrorMessage {
			t.Fatalf("Response should produce a message %s, but found this: %s", expectedErrorMessage, result.ErrorDescription)
		}

	})

	t.Run("Should fail for invalid token", func(t *testing.T) {
		token := "eyJhbGciOiJIUzM4NCIsInR5cCI6IkpXVCJ9.eyJwYXNzd29yZCI6IjJhYTYwYThmZjdmY2Q0NzNkMzIxZTAxNDZhZmQ5ZTI2ZGYzOTUxNDciLCJ1c2VybmFtZSI6ImNoaW1lcmFVc2VyIn0.H-bndVdfupspHOWl18i_oLq3pYq0156y7vnUKUTIb_8GvylX0YGZrIjbOP3VLaXD"
		requestURL := "/v1/tokens/check"
		req, _ := http.NewRequest("GET", requestURL, nil)
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
		response := httptest.NewRecorder()
		expectedErrorMessage := "Invalid Authorization token."

		handler.ServeHTTP(response, req)
		if response.Code != http.StatusUnauthorized {
			t.Errorf("Response should be %v, but it answer with %v ", http.StatusUnauthorized, response.Code)
		}
		var result helpers.FailedResponse
		err := json.NewDecoder(response.Body).Decode(&result)
		if err != nil {
			t.Fatalf("Response should produce a failed response, but found this: %s", response.Body)
		}

		if result.ErrorDescription != expectedErrorMessage {
			t.Fatalf("Response should produce a message %s, but found this: %s", expectedErrorMessage, result.ErrorDescription)
		}

	})

	t.Run("Should fail for not parsing the bearer token", func(t *testing.T) {
		token := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6InVzZXJuYW1lIiwicGFzc3dvcmQiOiI1YmFhNjFlNGM5YjkzZjNmMDY4MjI1MGI2Y2Y4MzMxYjdlZTY4ZmQ4In0.nugLnvFL1Znx5xFby5WjIkTUAAbCsqj2ROamzTNsUdqIb1ngBKNkMwvEuQpnK9c-Rki5N3MGYFUVJxpeayLh5AxnixUMvCEkVlZj-M4m-mQUSDP1MmgzvPY7h_jA5e58bjEGnnCqRawXVhXYNcGUZ9hxSu07yPaMhVoXuLIItaYq3nnF6g1s1YPytuvyf_NUmyYW0yzCvr3CMcyXldQN5OO4Jw08Wg1X-Y7S_BPnh9wTBLiSnUdd9R2WQs7k2QfAgKTWPit1xiVHLf76eWJ8-stpXYePtCyvTN3sOOCpqv-tDFvXOGe4QH8AvhYyYTXK-dLwTkPA_E_4l0jBU059Yw"
		requestURL := "/v1/tokens/check"
		req, _ := http.NewRequest("GET", requestURL, nil)
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
		response := httptest.NewRecorder()
		errorExpected := "there was an error, SigningMethod is invalid"

		handler.ServeHTTP(response, req)
		if response.Code != http.StatusInternalServerError {
			t.Errorf("Response should be %v, but it answer with %v ", http.StatusInternalServerError, response.Code)
		}
		var result helpers.FailedResponse
		err := json.NewDecoder(response.Body).Decode(&result)
		if err != nil {
			t.Fatalf("Response should produce a failed response, but found this: %s", response.Body)
		}

		if result.ErrorDescription != errorExpected {
			t.Fatalf("Response should produce a message %s, but found this: %s", errorExpected, result.ErrorDescription)
		}

	})

}

func TestAuthorizeMiddlewareByQueryParam(t *testing.T) {
	handler := ddrouter.New(ddrouter.WithServiceName("appName.http"))
	managerOptions := &manage.ManagerOptions{
		AppName: "salesforce-integration",
	}
	API(handler, managerOptions, apiConfig)

	t.Run("Should fail for not sending query parameter", func(t *testing.T) {
		requestURL := "/v1/tokens/check"
		req, _ := http.NewRequest("GET", requestURL, nil)
		response := httptest.NewRecorder()
		expectedErrorMessage := "An Authorization header or an Query Param with name token is required."

		handler.ServeHTTP(response, req)
		if response.Code != http.StatusUnauthorized {
			t.Errorf("Response should be %v, but it answer with %v ", http.StatusUnauthorized, response.Code)
		}
		var result helpers.FailedResponse
		err := json.NewDecoder(response.Body).Decode(&result)
		if err != nil {
			t.Fatalf("Response should produce a failed response, but found this: %s", response.Body)
		}

		if result.ErrorDescription != expectedErrorMessage {
			t.Fatalf("Response should produce a message %s, but found this: %s", expectedErrorMessage, result.ErrorDescription)
		}

	})

	t.Run("Should respond successfully with token", func(t *testing.T) {
		requestURL := "/v1/tokens/check?token=" + yaloTokenTest
		req, _ := http.NewRequest("GET", requestURL, nil)
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, req)
		if response.Code != http.StatusOK {
			t.Errorf("Response should be %v, but it answer with %v ", http.StatusOK, response.Code)
		}
	})

}

func TestParseBearerToken(t *testing.T) {
	app := &App{
		SecretKey: secretTest,
	}

	t.Run("Should retrieve the JWT with correct signature", func(t *testing.T) {
		token := yaloTokenTest
		jwt, err := app.parseBearerToken(token)

		if err != nil {
			t.Errorf("Response err should %v, but it answer with %v ", nil, err)
		}

		if jwt == nil {
			t.Errorf("Response jwt not should nil, but it answer with %v ", jwt)
		}
	})

	t.Run("Should recover the error by token with invalid signature method", func(t *testing.T) {
		token := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6InVzZXJuYW1lIiwicGFzc3dvcmQiOiI1YmFhNjFlNGM5YjkzZjNmMDY4MjI1MGI2Y2Y4MzMxYjdlZTY4ZmQ4In0.nugLnvFL1Znx5xFby5WjIkTUAAbCsqj2ROamzTNsUdqIb1ngBKNkMwvEuQpnK9c-Rki5N3MGYFUVJxpeayLh5AxnixUMvCEkVlZj-M4m-mQUSDP1MmgzvPY7h_jA5e58bjEGnnCqRawXVhXYNcGUZ9hxSu07yPaMhVoXuLIItaYq3nnF6g1s1YPytuvyf_NUmyYW0yzCvr3CMcyXldQN5OO4Jw08Wg1X-Y7S_BPnh9wTBLiSnUdd9R2WQs7k2QfAgKTWPit1xiVHLf76eWJ8-stpXYePtCyvTN3sOOCpqv-tDFvXOGe4QH8AvhYyYTXK-dLwTkPA_E_4l0jBU059Yw"
		_, err := app.parseBearerToken(token)
		errorExpected := "there was an error, SigningMethod is invalid"

		if err == nil {
			t.Errorf("Response err not should nil, but it answer with %v ", err)
		}

		if err.Error() != errorExpected {
			t.Errorf("Error should be contains %v, but it answer with %v ", errorExpected, err)
		}
	})
}

func TestSignedTokenString(t *testing.T) {
	app := &App{SecretKey: secretTest}

	t.Run("Should retrieve token signed", func(t *testing.T) {
		hash := sha1.New()
		hash.Write([]byte(yaloPasswordTest))
		user := AuthUser{Password: hex.EncodeToString(hash.Sum(nil)), Username: yaloUserTest, Role: Yalo}
		jwt, err := app.SignedTokenString(user)
		tokenExpected := yaloTokenTest

		if err != nil {
			t.Errorf("Response err should be nil, but it answer with %v ", err)
		}

		if jwt != tokenExpected {
			t.Errorf("Response should be %v, but it answer with %v ", tokenExpected, jwt)
		}

	})

	t.Run("Should get error when signing ", func(t *testing.T) {
		singMethod = &jwt.SigningMethodECDSA{}
		user := AuthUser{Password: yaloPasswordTest, Username: yaloUserTest}
		_, err := app.SignedTokenString(user)

		if err == nil {
			t.Errorf("Response err not should nil, but it answer with %v ", err)
		}
		singMethod = jwt.SigningMethodHS384
	})
}
