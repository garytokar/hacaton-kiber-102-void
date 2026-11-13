package middleware

import (
	"context"
	"net/http"
	"strings"

	"bashkir-backend/internal/pkg/jwt"

	"github.com/google/uuid"
)

type contextKey string

const (
	UserIDKey contextKey = "user_id"
	RoleKey   contextKey = "role"
)

func Auth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error":"missing authorization header"}`, http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, `{"error":"invalid authorization header format"}`, http.StatusUnauthorized)
				return
			}

			claims, err := jwt.ValidateToken(parts[1], secret)
			if err != nil {
				http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, RoleKey, claims.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserID(ctx context.Context) uuid.UUID {
	val, ok := ctx.Value(UserIDKey).(uuid.UUID)
	if !ok {
		return uuid.Nil
	}
	return val
}

func GetUserRole(ctx context.Context) string {
	val, ok := ctx.Value(RoleKey).(string)
	if !ok {
		return ""
	}
	return val
}

// RequireRoles проверяет, что текущий аутентифицированный пользователь обладает одной из разрешенных ролей.
func RequireRoles(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole := GetUserRole(r.Context())
			if userRole == "" {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			for _, rRole := range roles {
				if strings.EqualFold(userRole, rRole) {
					next.ServeHTTP(w, r)
					return
				}
			}
			http.Error(w, `{"error":"forbidden: insufficient permissions"}`, http.StatusForbidden)
		})
	}
}

