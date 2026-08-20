package tests

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/middlewares"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatch/controller"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatch/service"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/constants"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/jwks"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/samber/do"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestCalculateArgo_RequiresBearer(t *testing.T) {
	router, _ := newCalculateArgoRouter(t, http.NotFoundHandler(), true)

	req := httptest.NewRequest(http.MethodPost, "/api/trip/dispatch/calculate-argo", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestCalculateArgo_RejectsInvalidToken(t *testing.T) {
	router, _ := newCalculateArgoRouter(t, http.NotFoundHandler(), true)

	req := httptest.NewRequest(http.MethodPost, "/api/trip/dispatch/calculate-argo", bytes.NewBufferString(`{}`))
	req.Header.Set("Authorization", "Bearer not-a-token")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestCalculateArgo_RejectsInvalidBody(t *testing.T) {
	router, sign := newCalculateArgoRouter(t, http.NotFoundHandler(), true)

	req := httptest.NewRequest(http.MethodPost, "/api/trip/dispatch/calculate-argo", bytes.NewBufferString(`{
		"pickup_loc": ["1000", "106.8456"],
		"destination": ["-6.1754", "106.8272"],
		"vehicle_type": "motorcycle"
	}`))
	req.Header.Set("Authorization", "Bearer "+sign("user-1", "user@example.com", "customer"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCalculateArgo_ReturnsQuote(t *testing.T) {
	osrm := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"code": "Ok",
			"routes": [{
				"distance": 2000,
				"duration": 180,
				"geometry": {
					"type": "LineString",
					"coordinates": [[106.8456, -6.2088], [106.84, -6.2]]
				}
			}]
		}`)
	})
	router, sign := newCalculateArgoRouter(t, osrm, true)

	req := httptest.NewRequest(http.MethodPost, "/api/trip/dispatch/calculate-argo", bytes.NewBufferString(`{
		"pickup_loc": ["-6.2088", "106.8456"],
		"destination": ["-6.1754", "106.8272"],
		"vehicle_type": "motorcycle"
	}`))
	req.Header.Set("Authorization", "Bearer "+sign("user-1", "user@example.com", "customer"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var body struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
		Data    struct {
			Distance           int          `json:"distance"`
			Duration           int          `json:"duration"`
			Path               [][2]float64 `json:"path"`
			FarePerDistance    int          `json:"fare_per_distance"`
			PlatformPercentage int          `json:"platform_percentage"`
			TotalFare          int          `json:"total_fare"`
			VehicleType        string       `json:"vehicle_type"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.True(t, body.Status)
	assert.Equal(t, 2000, body.Data.Distance)
	assert.Equal(t, 180, body.Data.Duration)
	assert.Equal(t, 2500, body.Data.FarePerDistance)
	assert.Equal(t, 10, body.Data.PlatformPercentage)
	assert.Equal(t, 5500, body.Data.TotalFare)
	assert.Equal(t, "motorcycle", body.Data.VehicleType)
	require.Equal(t, [][2]float64{{-6.2088, 106.8456}, {-6.2, 106.84}}, body.Data.Path)
}

func TestCalculateArgo_NoRouteIsBadRequest(t *testing.T) {
	osrm := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"code":"NoRoute","routes":[]}`)
	})
	router, sign := newCalculateArgoRouter(t, osrm, true)

	req := httptest.NewRequest(http.MethodPost, "/api/trip/dispatch/calculate-argo", bytes.NewBufferString(`{
		"pickup_loc": ["-6.2088", "106.8456"],
		"destination": ["-6.1754", "106.8272"],
		"vehicle_type": "car"
	}`))
	req.Header.Set("Authorization", "Bearer "+sign("user-1", "user@example.com", "customer"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCalculateArgo_OSRMDownIsBadGateway(t *testing.T) {
	osrm := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	router, sign := newCalculateArgoRouter(t, osrm, true)

	req := httptest.NewRequest(http.MethodPost, "/api/trip/dispatch/calculate-argo", bytes.NewBufferString(`{
		"pickup_loc": ["-6.2088", "106.8456"],
		"destination": ["-6.1754", "106.8272"],
		"vehicle_type": "car"
	}`))
	req.Header.Set("Authorization", "Bearer "+sign("user-1", "user@example.com", "customer"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadGateway, rec.Code)
}

func TestCalculateArgo_DeniesWhenUnauthorized(t *testing.T) {
	router, sign := newCalculateArgoRouter(t, http.NotFoundHandler(), false)

	req := httptest.NewRequest(http.MethodPost, "/api/trip/dispatch/calculate-argo", bytes.NewBufferString(`{
		"pickup_loc": ["-6.2088", "106.8456"],
		"destination": ["-6.1754", "106.8272"],
		"vehicle_type": "motorcycle"
	}`))
	req.Header.Set("Authorization", "Bearer "+sign("user-1", "drv@example.com", "driver"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func newCalculateArgoRouter(t *testing.T, osrmHandler http.Handler, allow bool) (*gin.Engine, func(userID, email, role string) string) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	osrmServer := httptest.NewServer(osrmHandler)
	t.Cleanup(osrmServer.Close)

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

	injector := do.New()
	do.ProvideNamed(injector, constants.DB, func(i *do.Injector) (*gorm.DB, error) {
		return &gorm.DB{}, nil
	})

	dispatchSvc := service.NewDispatchService(nil, nil, osrmServer.Client(), osrmServer.URL)
	dispatchCtrl := controller.NewDispatchController(injector, dispatchSvc)
	verifier := jwks.NewVerifier(jwksServer.URL, "go-ojol-auth")

	router := gin.New()
	router.POST(
		"/api/trip/dispatch/calculate-argo",
		middlewares.Authenticate(verifier),
		middlewares.Authorize(&stubEnforcer{allow: allow}, constants.ENUM_RESOURCE_TRIP, constants.ENUM_ACTION_CREATE),
		dispatchCtrl.CalculateArgo,
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
