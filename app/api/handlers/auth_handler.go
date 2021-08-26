package handlers

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/context"
	"github.com/julienschmidt/httprouter"
	"yalochat.com/salesforce-integration/base/helpers"
)

// User structure that will allow validating and generating jwt
type AuthUser struct {
	Username string   `json:"username,omitempty"`
	Password string   `json:"password,omitempty"`
	Role     RoleType `json:"role,omitempty"`
}

//Structure that will allow returning a token as a response
type tokenResult struct {
	Token string `json:"token"`
}

// Status that a conversation can have
type RoleType string

const (
	Yalo       RoleType = "YALO_ROLE"
	Salesforce RoleType = "SALESFORCE_ROLE"
	userKey             = "authUser"
	roleKey             = "role"
)

// Signing method to create JWT
var singMethod jwt.SigningMethod = jwt.SigningMethodHS384

// The authenticate function will generate a JWT necessary to be able to use the endpoints,
// sending the username and password in the request, and validating that they are the same
// as the environment variables.
func (app *App) authenticate(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var userRequest AuthUser
	if err := json.NewDecoder(r.Body).Decode(&userRequest); err != nil {
		helpers.WriteFailedResponse(w, http.StatusBadRequest, helpers.InvalidPayload+" : "+err.Error())
		return
	}

	user, err := app.settingRoleByUsername(userRequest)
	if err != nil {
		helpers.WriteFailedResponse(w, http.StatusUnauthorized, err.Error())
		return
	}

	hash := sha1.New()
	hash.Write([]byte(user.Password))
	user.Password = hex.EncodeToString(hash.Sum(nil))
	if !app.validateCredentials(*user) {
		helpers.WriteFailedResponse(w, http.StatusUnauthorized, "Invalid credentials.")
		return
	}
	token, err := app.SignedTokenString(*user)
	if err != nil {
		helpers.WriteFailedResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	helpers.WriteSuccessResponse(w, &tokenResult{Token: token})
}

// Function getUserByToken will return the username and role if the token sent in
//the Authorization header is valid and can be used in the other endpoints
func (app *App) getUserByToken(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	user := context.Get(req, userKey).(AuthUser)
	user.Password = ""
	helpers.WriteSuccessResponse(w, user)
}

// authorizeMiddleware validates the token if it is sent by the Authorization header or query param
func (app *App) authorizeMiddleware(next httprouter.Handle, roles []RoleType) httprouter.Handle {
	return httprouter.Handle(func(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
		authorizationHeader := req.Header.Get("Authorization")
		if authorizationHeader != "" {
			bearerToken := strings.Split(authorizationHeader, " ")
			if len(bearerToken) == 2 {
				if bearerToken[0] == "Bearer" {
					app.validToken(w, req, params, next, bearerToken[1], roles)
					return
				}
			}
			helpers.WriteFailedResponse(w, http.StatusUnauthorized, "Invalid Authorization token, header invalid.")
			return
		}

		if len(req.URL.Query()) > 0 {
			token := req.URL.Query().Get("token")
			if token != "" {
				app.validToken(w, req, params, next, token, roles)
				return
			}
		}

		helpers.WriteFailedResponse(w, http.StatusUnauthorized, "An Authorization header or an Query Param with name token is required.")
	})
}

func (app *App) validToken(w http.ResponseWriter, req *http.Request, params httprouter.Params, next httprouter.Handle, token string, roles []RoleType) {
	tokenParse, err := app.parseBearerToken(token)
	if err != nil {
		helpers.WriteFailedResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	if tokenParse.Valid {
		var user AuthUser
		tokenClaims, _ := helpers.MarshalJSON(tokenParse.Claims)
		json.Unmarshal(tokenClaims, &user)
		if app.validateCredentials(user) {
			for _, role := range roles {
				if user.Role == role {
					context.Set(req, userKey, user)
					next(w, req, params)
					return
				}
			}
			helpers.WriteFailedResponse(w, http.StatusForbidden, "Invalid Authorization token.")
			return
		}
	}
	helpers.WriteFailedResponse(w, http.StatusUnauthorized, "Invalid Authorization token.")
}

func (app *App) parseBearerToken(bearerToken string) (*jwt.Token, error) {
	return jwt.Parse(bearerToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("there was an error, SigningMethod is invalid")
		}
		return []byte(app.SecretKey), nil
	})
}

func (app *App) validateCredentials(user AuthUser) bool {
	switch user.Role {
	case Yalo:
		if user.Username == app.YaloUsername && user.Password == app.YaloPassword {
			return true
		}
	// Salesforce Role
	case Salesforce:
		if user.Username == app.SalesforceUsername && user.Password == app.SalesforcePassword {
			return true
		}
	}
	return false
}

// signedTokenString signs the jwt with the signature method and the secret of the api
func (app *App) SignedTokenString(user AuthUser) (string, error) {
	token := jwt.NewWithClaims(singMethod, jwt.MapClaims{
		"username": user.Username,
		"password": user.Password,
		"role":     user.Role,
	})

	signedToken, err := token.SignedString([]byte(app.SecretKey))
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

func (app *App) settingRoleByUsername(user AuthUser) (*AuthUser, error) {

	if user.Username == app.YaloUsername {
		user.Role = Yalo
	}

	if user.Username == app.SalesforceUsername {
		user.Role = Salesforce
	}

	if user.Role == "" {
		return nil, fmt.Errorf("username Invalid")
	}
	return &user, nil
}
