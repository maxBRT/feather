package auth

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc"
	"github.com/golang-jwt/jwt/v4"
	"github.com/joho/godotenv"
)

type AuthProvider struct {
	Issuer   string
	Audience string
	JWKSUrl  string
	jwks     *keyfunc.JWKS
}

// Creates a new AuthProvider instance with configuration values
// retrieved from environment variables. It initializes the AuthProvider
// with JWKS_URL, ISSUER, and AUDIENCE.
func NewAuthProvider() (*AuthProvider, error) {
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	jwksURL := os.Getenv("JWKS_URL")
	issuer := os.Getenv("ISSUER")
	audience := os.Getenv("AUDIENCE")
	if issuer == "" || audience == "" || jwksURL == "" {
		return nil, errors.New("missing ISSUER/AUDIENCE/JWKS_URL")
	}

	jwks, err := keyfunc.Get(jwksURL, keyfunc.Options{
		RefreshInterval:   5 * time.Minute,
		RefreshUnknownKID: true,
	})

	if err != nil {
		return nil, err
	}
	return &AuthProvider{
		JWKSUrl:  jwksURL,
		Issuer:   issuer,
		Audience: audience,
		jwks:     jwks,
	}, nil
}

// Verify a bearer token and return claims.
func (ap *AuthProvider) verify(tokenString string) (jwt.MapClaims, error) {
	tok, err := jwt.Parse(tokenString, ap.jwks.Keyfunc)
	if err != nil || !tok.Valid {
		return nil, errors.New("invalid token")
	}
	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("bad claims")
	}
	if !claims.VerifyIssuer(ap.Issuer, true) {
		return nil, errors.New("bad issuer")
	}
	if !claims.VerifyAudience(ap.Audience, true) {
		return nil, errors.New("bad audience")
	}
	if !claims.VerifyExpiresAt(time.Now().Unix(), true) {
		return nil, errors.New("expired")
	}
	return claims, nil
}

// Middleware: extracts Bearer, verifies, injects claims into context.
func (ap *AuthProvider) JWTMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := r.Header.Get("Authorization")
		if h == "" || !strings.HasPrefix(h, "Bearer ") {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		claims, err := ap.verify(strings.TrimPrefix(h, "Bearer "))
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), "userClaims", claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
