package tests

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/modules/auth/service"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestJWTService(t *testing.T) (service.JWTService, *ecdsa.PrivateKey) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	svc, err := service.NewJWTServiceFromKey(key, "", "test-issuer")
	require.NoError(t, err)

	return svc, key
}

func TestJWTService_GenerateAndValidateAccessToken(t *testing.T) {
	svc, _ := newTestJWTService(t)

	token, err := svc.GenerateAccessToken("user-1", "admin@example.com", "admin", "session-1")
	require.NoError(t, err)
	require.NotEmpty(t, token)

	parsed, err := svc.ValidateToken(token)
	require.NoError(t, err)
	assert.True(t, parsed.Valid)

	userID, err := svc.GetUserIDByToken(token)
	require.NoError(t, err)
	assert.Equal(t, "user-1", userID)

	email, err := svc.GetEmailByToken(token)
	require.NoError(t, err)
	assert.Equal(t, "admin@example.com", email)

	sessionID, err := svc.GetSessionIDByToken(token)
	require.NoError(t, err)
	assert.Equal(t, "session-1", sessionID)
}

func TestJWTService_JWKSMatchesSignedHeader(t *testing.T) {
	svc, _ := newTestJWTService(t)

	token, err := svc.GenerateAccessToken("user-1", "user@example.com", "customer", "session-1")
	require.NoError(t, err)

	jwks := svc.JWKS()
	require.Len(t, jwks.Keys, 1)
	assert.Equal(t, "EC", jwks.Keys[0].Kty)
	assert.Equal(t, "P-256", jwks.Keys[0].Crv)
	assert.Equal(t, "sig", jwks.Keys[0].Use)
	assert.Equal(t, "ES256", jwks.Keys[0].Alg)
	assert.NotEmpty(t, jwks.Keys[0].X)
	assert.NotEmpty(t, jwks.Keys[0].Y)
	assert.NotEmpty(t, jwks.Keys[0].Kid)

	parser := new(jwt.Parser)
	unverified, _, err := parser.ParseUnverified(token, jwt.MapClaims{})
	require.NoError(t, err)
	assert.Equal(t, "ES256", unverified.Header["alg"])
	assert.Equal(t, jwks.Keys[0].Kid, unverified.Header["kid"])
}

func TestJWTService_RejectsHMACToken(t *testing.T) {
	svc, _ := newTestJWTService(t)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": "user-1",
		"role":    "user",
	})
	signed, err := token.SignedString([]byte("not-asymmetric"))
	require.NoError(t, err)

	_, err = svc.ValidateToken(signed)
	assert.Error(t, err)
}

func TestJWTService_RejectsTokenSignedWithDifferentKey(t *testing.T) {
	svc, _ := newTestJWTService(t)

	otherKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"user_id": "user-1",
		"role":    "user",
	})
	signed, err := token.SignedString(otherKey)
	require.NoError(t, err)

	_, err = svc.ValidateToken(signed)
	assert.Error(t, err)
}

func TestNewJWTService_LoadsPEMFromPath(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	der, err := x509.MarshalECPrivateKey(key)
	require.NoError(t, err)

	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: der})
	path := filepath.Join(t.TempDir(), "ec-private.pem")
	require.NoError(t, os.WriteFile(path, pemBytes, 0o600))

	t.Setenv("JWT_PRIVATE_KEY_PATH", path)
	t.Setenv("JWT_ISSUER", "test-issuer")
	t.Setenv("JWT_KID", "test-kid")

	svc, err := service.NewJWTService()
	require.NoError(t, err)

	jwks := svc.JWKS()
	require.Len(t, jwks.Keys, 1)
	assert.Equal(t, "test-kid", jwks.Keys[0].Kid)

	token, err := svc.GenerateAccessToken("user-1", "user@example.com", "customer", "session-1")
	require.NoError(t, err)

	parsed, err := svc.ValidateToken(token)
	require.NoError(t, err)
	assert.True(t, parsed.Valid)
}

func TestJWTService_GenerateRefreshToken(t *testing.T) {
	svc, _ := newTestJWTService(t)

	token, expiresAt, err := svc.GenerateRefreshToken()
	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.True(t, expiresAt.After(time.Now()))
}

func TestNewJWTServiceFromKey_RejectsNonP256(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	require.NoError(t, err)

	_, err = service.NewJWTServiceFromKey(key, "", "test")
	assert.Error(t, err)
}
