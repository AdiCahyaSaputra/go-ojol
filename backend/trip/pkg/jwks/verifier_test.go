package jwks

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVerifier_ValidES256Token(t *testing.T) {
	key, kid, server := newJWKSServer(t)
	defer server.Close()

	raw := signToken(t, key, kid, jwt.MapClaims{
		"user_id": "user-1",
		"email":   "user@example.com",
		"role":    "customer",
		"iss":     defaultIssuer,
		"exp":     time.Now().Add(time.Minute).Unix(),
		"iat":     time.Now().Unix(),
	})

	claims, err := NewVerifier(server.URL, defaultIssuer).Verify(raw)
	require.NoError(t, err)
	assert.Equal(t, "user-1", claims.UserID)
	assert.Equal(t, "user@example.com", claims.Email)
	assert.Equal(t, "customer", claims.Role)
}

func TestVerifier_RejectsHMACToken(t *testing.T) {
	_, _, server := newJWKSServer(t)
	defer server.Close()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": "user-1",
		"email":   "user@example.com",
		"role":    "customer",
		"iss":     defaultIssuer,
		"exp":     time.Now().Add(time.Minute).Unix(),
	})
	token.Header["kid"] = "any"
	raw, err := token.SignedString([]byte("not-asymmetric"))
	require.NoError(t, err)

	_, err = NewVerifier(server.URL, defaultIssuer).Verify(raw)
	assert.Error(t, err)
}

func TestVerifier_RejectsWrongKey(t *testing.T) {
	_, kid, server := newJWKSServer(t)
	defer server.Close()

	other, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	raw := signToken(t, other, kid, jwt.MapClaims{
		"user_id": "user-1",
		"email":   "user@example.com",
		"role":    "customer",
		"iss":     defaultIssuer,
		"exp":     time.Now().Add(time.Minute).Unix(),
	})

	_, err = NewVerifier(server.URL, defaultIssuer).Verify(raw)
	assert.Error(t, err)
}

func TestVerifier_RejectsUnknownKid(t *testing.T) {
	key, _, server := newJWKSServer(t)
	defer server.Close()

	raw := signToken(t, key, "missing-kid", jwt.MapClaims{
		"user_id": "user-1",
		"email":   "user@example.com",
		"role":    "customer",
		"iss":     defaultIssuer,
		"exp":     time.Now().Add(time.Minute).Unix(),
	})

	_, err := NewVerifier(server.URL, defaultIssuer).Verify(raw)
	assert.Error(t, err)
}

func TestVerifier_RejectsExpiredToken(t *testing.T) {
	key, kid, server := newJWKSServer(t)
	defer server.Close()

	raw := signToken(t, key, kid, jwt.MapClaims{
		"user_id": "user-1",
		"email":   "user@example.com",
		"role":    "customer",
		"iss":     defaultIssuer,
		"exp":     time.Now().Add(-time.Minute).Unix(),
		"iat":     time.Now().Add(-2 * time.Minute).Unix(),
	})

	_, err := NewVerifier(server.URL, defaultIssuer).Verify(raw)
	assert.Error(t, err)
}

func TestVerifier_RejectsWrongIssuer(t *testing.T) {
	key, kid, server := newJWKSServer(t)
	defer server.Close()

	raw := signToken(t, key, kid, jwt.MapClaims{
		"user_id": "user-1",
		"email":   "user@example.com",
		"role":    "customer",
		"iss":     "other-issuer",
		"exp":     time.Now().Add(time.Minute).Unix(),
	})

	_, err := NewVerifier(server.URL, defaultIssuer).Verify(raw)
	assert.Error(t, err)
}

func newJWKSServer(t *testing.T) (*ecdsa.PrivateKey, string, *httptest.Server) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	kid := "test-kid"
	jwks := JWKS{Keys: []JWK{{
		Kty: "EC",
		Crv: "P-256",
		X:   base64.RawURLEncoding.EncodeToString(key.PublicKey.X.FillBytes(make([]byte, p256CoordSize))),
		Y:   base64.RawURLEncoding.EncodeToString(key.PublicKey.Y.FillBytes(make([]byte, p256CoordSize))),
		Use: "sig",
		Alg: "ES256",
		Kid: kid,
	}}}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(jwks)
	}))

	return key, kid, server
}

func signToken(t *testing.T, key *ecdsa.PrivateKey, kid string, claims jwt.MapClaims) string {
	t.Helper()

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = kid
	raw, err := token.SignedString(key)
	require.NoError(t, err)
	return raw
}
