package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type JWTManager struct {
	secretKey     string
	tokenDuration time.Duration
}

// struct with the jwt standard claims
type UserClaim struct {
	jwt.StandardClaims
	Username string `json:"username"`
	Type     string `json:"type"`
	Avatar   string `json:"avatar"`
}

type authHandler struct {
	next http.Handler
}

// Create and sign a new jwt token for user
func (manager *JWTManager) Generate(username, userType, avatar string) (string, error) {

	claims := UserClaim{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(manager.tokenDuration).Unix(),
		},
		Username: username,
		Type:     userType,
		Avatar:   avatar,
	}

	// create a unsigned token
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	// sign the token with the secretKey
	return token.SignedString([]byte(manager.secretKey))
}

// Verify the token
func (manager *JWTManager) Verify(accessToken string) (*UserClaim, error) {

	token, err := jwt.ParseWithClaims(
		accessToken,
		&UserClaim{},
		func(token *jwt.Token) (interface{}, error) {
			_, ok := token.Method.(*jwt.SigningMethodHMAC)
			if !ok {
				return nil, fmt.Errorf("Error with token sign method")
			}
			return []byte(manager.secretKey), nil
		},
	)

	if err != nil {
		return nil, fmt.Errorf("Invalid token: %w", err)
	}

	// convert the claims in a *UserClaim object
	claims, ok := token.Claims.(*UserClaim)
	if !ok {
		return nil, fmt.Errorf("Invalid token claim")
	}

	return claims, nil
}

// MustAuth is a "constructor" for the authHandler type
func MustAuth(handler http.Handler) http.Handler {
	return &authHandler{next: handler}
}

func (h *authHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", corsServer)
	w.Header().Set("Access-Control-Allow-Headers", "Authorization")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")

	// "handle" the cors pre-flight request (https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS#preflighted_requests)
	if r.Method == "OPTIONS" {
		return
	}

	token := strings.Fields(r.Header.Get("Authorization"))[1]
	jwt := &JWTManager{secretKey, 0}

	_, err := jwt.Verify(token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// success - call the next handler
	h.next.ServeHTTP(w, r)
}
