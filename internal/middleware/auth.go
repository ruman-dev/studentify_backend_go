package middleware

import (
	"context"
	"net/http"
	"strings"

	"softixa-solutions.com/studentify/internal/utils"
)

type contextKey string

const claimsKey contextKey = "claims"

// Authenticate validates the Bearer access token and stores claims in the request context.
func Authenticate(tokens *utils.TokenManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := strings.TrimSpace(r.Header.Get("Authorization"))
			if header == "" {
				utils.Error(w, http.StatusUnauthorized, "Authorization header is required", utils.ErrInvalidToken)
				return
			}

			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
				utils.Error(w, http.StatusUnauthorized, "Invalid authorization format. Use: Bearer <token>", utils.ErrInvalidToken)
				return
			}

			claims, err := tokens.Parse(strings.TrimSpace(parts[1]))
			if err != nil {
				utils.Error(w, http.StatusUnauthorized, "Invalid or expired token", utils.ErrInvalidToken)
				return
			}

			ctx := context.WithValue(r.Context(), claimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ClaimsFromContext returns JWT claims set by Authenticate middleware.
func ClaimsFromContext(ctx context.Context) (*utils.Claims, bool) {
	claims, ok := ctx.Value(claimsKey).(*utils.Claims)
	return claims, ok
}

// UserIDFromContext returns the authenticated user id, if present.
func UserIDFromContext(ctx context.Context) (string, bool) {
	claims, ok := ClaimsFromContext(ctx)
	if !ok || claims.UserID == "" {
		return "", false
	}
	return claims.UserID, true
}

// RoleFromContext returns the authenticated user role, if present.
func RoleFromContext(ctx context.Context) (string, bool) {
	claims, ok := ClaimsFromContext(ctx)
	if !ok || claims.Role == "" {
		return "", false
	}
	return claims.Role, true
}

// RequireRole ensures the authenticated user has one of the allowed roles.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[strings.ToUpper(role)] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := RoleFromContext(r.Context())
			if !ok {
				utils.Error(w, http.StatusUnauthorized, "Unauthorized", utils.ErrInvalidToken)
				return
			}
			if _, exists := allowed[strings.ToUpper(role)]; !exists {
				utils.Error(w, http.StatusForbidden, "Forbidden", utils.ErrInvalidRole)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
