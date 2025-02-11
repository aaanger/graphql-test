package middleware

import (
	"context"
	"errors"
	"github.com/aaanger/graphql-test/pkg/jwt"
	"net/http"
	"strings"
)

func UserIdentity(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")

		if header == "" {
			next.ServeHTTP(w, r)
			return
		}

		headerParts := strings.Split(header, " ")

		if len(headerParts) != 2 {
			http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
			return
		}

		userID, err := jwt.ParseToken(headerParts[1])
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
		}

		ctx := context.WithValue(r.Context(), "userID", userID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserID(ctx context.Context) (int, error) {
	id := ctx.Value("userID")

	if id == nil {
		return 0, errors.New("user id not found")
	}

	userID, ok := id.(int)
	if !ok {
		return 0, errors.New("invalid type of user id")
	}

	return userID, nil
}
