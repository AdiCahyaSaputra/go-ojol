package tests

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

	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/middlewares"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/trip/controller"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/constants"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/jwks"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProtected_RequiresBearer(t *testing.T) {
	router, _ := newProtectedRouter(t, true)

	req := httptest.NewRequest(http.MethodGet, "/api/trip/protected", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestProtected_RejectsInvalidToken(t *testing.T) {
	router, _ := newProtectedRouter(t, true)

	req := httptest.NewRequest(http.MethodGet, "/api/trip/protected", nil)
	req.Header.Set("Authorization", "Bearer not-a-token")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestProtected_ReturnsClaims(t *testing.T) {
	router, sign := newProtectedRouter(t, true)
	raw := sign("user-1", "user@example.com", "customer")

	req := httptest.NewRequest(http.MethodGet, "/api/trip/protected", nil)
	req.Header.Set("Authorization", "Bearer "+raw)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var body struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
		Data    struct {
			UserID string `json:"user_id"`
			Email  string `json:"email"`
			Role   string `json:"role"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.True(t, body.Status)
	assert.Equal(t, "user-1", body.Data.UserID)
	assert.Equal(t, "user@example.com", body.Data.Email)
	assert.Equal(t, "customer", body.Data.Role)
}

func TestProtected_DeniesWhenUnauthorized(t *testing.T) {
	router, sign := newProtectedRouter(t, false)
	raw := sign("user-1", "drv@example.com", "driver")

	req := httptest.NewRequest(http.MethodGet, "/api/trip/protected", nil)
	req.Header.Set("Authorization", "Bearer "+raw)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func newProtectedRouter(t *testing.T, allow bool) (*gin.Engine, func(userID, email, role string) string) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	kid := "test-kid"
	jwksBody := jwks.JWKS{Keys: []jwks.JWK{{
		Kty: "EC",
		Crv: "P-256",
		X:   base64.RawURLEncoding.EncodeToString(key.PublicKey.X.FillBytes(make([]byte, 32))),
		Y:   base64.RawURLEncoding.EncodeToString(key.PublicKey.Y.FillBytes(make([]byte, 32))),
		Use: "sig",
		Alg: "ES256",
		Kid: kid,
	}}}

	jwksServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(jwksBody)
	}))
	t.Cleanup(jwksServer.Close)

	verifier := jwks.NewVerifier(jwksServer.URL, "go-ojol-auth")
	router := gin.New()
	router.GET(
		"/api/trip/protected",
		middlewares.Authenticate(verifier),
		middlewares.Authorize(&stubEnforcer{allow: allow}, "", constants.ENUM_RESOURCE_TRIP, constants.ENUM_ACTION_READ),
		controller.NewTripController().Protected,
	)

	sign := func(userID, email, role string) string {
		token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
			"user_id": userID,
			"email":   email,
			"role":    role,
			"iss":     "go-ojol-auth",
			"exp":     time.Now().Add(time.Minute).Unix(),
			"iat":     time.Now().Unix(),
		})
		token.Header["kid"] = kid
		raw, err := token.SignedString(key)
		require.NoError(t, err)
		return raw
	}

	return router, sign
}

type stubEnforcer struct {
	allow bool
}

func (s *stubEnforcer) Enforce(rvals ...interface{}) (bool, error) {
	return s.allow, nil
}

func (s *stubEnforcer) LoadPolicy() error {
	return nil
}
