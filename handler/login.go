package handler

import (
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"log"
	"net/http"
	"strings"
	"time"
)

func (handler *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var loginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	err := json.NewDecoder(r.Body).Decode(&loginRequest)
	if err != nil {
		log.Printf("Error decoding login request: %v\n", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if loginRequest.Username == handler.Config.GetString("Login.username") &&
		loginRequest.Password == handler.Config.GetString("Login.password") {
		tokenString, err := handler.generateJWTToken(loginRequest.Username)
		if err != nil {
			log.Printf("Error generating JWT token: %v\n", err)
			http.Error(w, "Error generating JWT token", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
		log.Println("Admin Login Succeded")
	} else {
		log.Println("Invalid login attempt with username:", loginRequest.Username)
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
	}
}

func (handler *Handler) isValidToken(r *http.Request) bool {
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		return false
	}

	tokenParts := strings.SplitN(tokenString, " ", 2)
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		return false
	}

	token, err := jwt.Parse(tokenParts[1], func(token *jwt.Token) (interface{}, error) {
		return []byte(handler.Config.GetString("Jwt_key")), nil
	})

	if err != nil || !token.Valid {
		log.Println("Token validation failed:", err)
		return false
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.Println("Invalid claims format")
		return false
	}
	expClaim, ok := claims["exp"]
	if !ok {
		log.Println("Missing 'exp' claim")
		return false
	}
	expTime, err := time.Parse(time.RFC3339, expClaim.(string))
	if err != nil {
		log.Println("Error parsing 'exp' claim:", err)
		return false
	}
	currentTimeLocal := time.Now().Local()
	expirationTimeLocal := expTime.Local()
	if currentTimeLocal.After(expirationTimeLocal) {
		log.Println("Token has expired.")
		return false
	}
	return true
}
func (handler *Handler) generateJWTToken(username string) (string, error) {
	key := []byte(handler.Config.GetString("Jwt_key"))
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(time.Hour * 1),
	})
	tokenString, err := token.SignedString(key)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
