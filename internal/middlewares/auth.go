package middlewares

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserId   string  `json:"userId"`
	Email    string  `json:"email"`
	Role     string  `json:"role"`
	Username *string `json:"username"`
	Avatar   string  `json:"avatar"`
	jwt.RegisteredClaims
}

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("authorization")
		if token == "" {
			http.Error(w, "auth token not found", http.StatusUnauthorized)
			return
		}
		token = strings.TrimPrefix(token, "Bearer ")
		parsedToken, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("ACCESS_SECRET")), nil
		})
		if err != nil {
			http.Error(w, "unverified token signature", http.StatusUnauthorized)
			return
		}
		if claims, ok := parsedToken.Claims.(*Claims); ok {
			ctx := context.WithValue(r.Context(), "decoded", claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}

func AuthAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("authorization")
		if token == "" {
			http.Error(w, "auth token not found", http.StatusUnauthorized)
			return
		}
		token = strings.TrimPrefix(token, "Bearer ")
		parsedToken, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("ACCESS_SECRET")), nil
		})
		if err != nil {
			http.Error(w, "unverified token signature", http.StatusBadRequest)
			return
		}
		if claims, ok := parsedToken.Claims.(*Claims); ok {
			if claims.Role != "ADMIN" {
				http.Error(w, "user is not authorised to perform this operation", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), "decoded", claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}
