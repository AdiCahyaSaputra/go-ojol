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

	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/database/entities"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/middlewares"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatch/controller"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatch/dto"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatch/service"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/constants"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/drivergeo"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/jwks"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/session"
	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/samber/do"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type stubEnforcer struct {
	allow bool
}

func (s *stubEnforcer) Enforce(rvals ...interface{}) (bool, error) {
	return s.allow, nil
}

func (s *stubEnforcer) LoadPolicy() error {
	return nil
}

type stubDispatchRepo struct {
	vehicle            *entities.Vehicle
	vehicleErr         error
	vehicleCategories  []dto.VehicleCategory
	vehicleCatErr      error
	pendingErr         error
	nearbyProfiles     map[string]dto.NearbyDriverProfile
	nearbyProfilesErr  error
}

func (s *stubDispatchRepo) VehicleById(id uuid.UUID) (*entities.Vehicle, error) {
	if s.vehicleErr != nil {
		return nil, s.vehicleErr
	}
	if s.vehicle == nil {
		return nil, nil
	}
	if s.vehicle.ID != id {
		return nil, nil
	}
	return s.vehicle, nil
}

func (s *stubDispatchRepo) DistinctVehicleCategories() ([]dto.VehicleCategory, error) {
	if s.vehicleCatErr != nil {
		return nil, s.vehicleCatErr
	}
	if s.vehicleCategories != nil {
		return s.vehicleCategories, nil
	}
	return []dto.VehicleCategory{
		{VehicleType: entities.VehicleTypeMotorcycle, MaxSize: 1},
		{VehicleType: entities.VehicleTypeCar, MaxSize: 4},
	}, nil
}

func (s *stubDispatchRepo) PendingArgoTransaction(req dto.PendingArgoTransaction) error {
	return s.pendingErr
}

func (s *stubDispatchRepo) NearbyDriverProfiles(_ []uuid.UUID, _ entities.VehicleType) (map[string]dto.NearbyDriverProfile, error) {
	if s.nearbyProfilesErr != nil {
		return nil, s.nearbyProfilesErr
	}
	if s.nearbyProfiles != nil {
		return s.nearbyProfiles, nil
	}
	return map[string]dto.NearbyDriverProfile{}, nil
}

func testCustomer() entities.Customer {
	return entities.Customer{
		ID:     uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		UserID: uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		Name:   "Test Customer",
	}
}

func testVehicle(vehicleType entities.VehicleType) *entities.Vehicle {
	return &entities.Vehicle{
		ID:            uuid.MustParse("33333333-3333-3333-3333-333333333333"),
		Name:          "Test Vehicle",
		LicenseNumber: "B1234XX",
		MaxSize:       2,
		Type:          vehicleType,
	}
}

func injectCustomer() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set("customer", testCustomer())
		ctx.Next()
	}
}

func newTestSigner(t *testing.T) (verifier jwks.Verifier, sign func(userID, email, role string) string) {
	t.Helper()

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

	verifier = jwks.NewVerifier(jwksServer.URL, "go-ojol-auth")
	sign = func(userID, email, role string) string {
		token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
			"user_id":    userID,
			"email":      email,
			"role":       role,
			"session_id": "11111111-1111-1111-1111-111111111111",
			"iss":        "go-ojol-auth",
			"exp":        time.Now().Add(time.Minute).Unix(),
			"iat":        time.Now().Unix(),
		})
		token.Header["kid"] = kid
		raw, err := token.SignedString(key)
		require.NoError(t, err)
		return raw
	}
	return verifier, sign
}

func newTestInjector(t *testing.T) *do.Injector {
	t.Helper()
	injector := do.New()
	do.ProvideNamed(injector, constants.DB, func(i *do.Injector) (*gorm.DB, error) {
		return &gorm.DB{}, nil
	})
	return injector
}

func newGeoStore(t *testing.T) *drivergeo.Store {
	t.Helper()
	mr := miniredis.RunT(t)
	return drivergeo.NewStore(redis.NewClient(&redis.Options{Addr: mr.Addr()}))
}

func newCalculateArgoRouter(t *testing.T, osrmHandler http.Handler, allow bool) (*gin.Engine, func(userID, email, role string) string) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	osrmServer := httptest.NewServer(osrmHandler)
	t.Cleanup(osrmServer.Close)

	verifier, sign := newTestSigner(t)
	injector := newTestInjector(t)

	repo := &stubDispatchRepo{}
	dispatchSvc := service.NewDispatchService(repo, nil, osrmServer.Client(), osrmServer.URL, nil)
	dispatchCtrl := controller.NewDispatchController(injector, dispatchSvc)

	router := gin.New()
	router.GET(
		"/api/trip/dispatch/customer/calculate-argo",
		middlewares.Authenticate(verifier, session.AlwaysActive()),
		middlewares.Authorize(&stubEnforcer{allow: allow}, constants.ENUM_ROLE_CUSTOMER, constants.ENUM_RESOURCE_DISPATCH, constants.ENUM_ACTION_READ),
		injectCustomer(),
		dispatchCtrl.CalculateArgo,
	)

	return router, sign
}

func newFindDriverRouter(t *testing.T, allow bool, store *drivergeo.Store) (*gin.Engine, func(userID, email, role string) string) {
	t.Helper()
	router, sign, _ := newFindDriverRouterWithStoreAndAllow(t, allow, store, nil)
	return router, sign
}

func newFindDriverRouterWithStore(t *testing.T, allow bool) (*gin.Engine, func(userID, email, role string) string, *drivergeo.Store) {
	t.Helper()
	store := newGeoStore(t)
	router, sign, _ := newFindDriverRouterWithStoreAndAllow(t, allow, store, nil)
	return router, sign, store
}

func newFindDriverRouterWithStoreAndAllow(
	t *testing.T,
	allow bool,
	store *drivergeo.Store,
	repo *stubDispatchRepo,
) (*gin.Engine, func(userID, email, role string) string, *drivergeo.Store) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	if store == nil {
		store = newGeoStore(t)
	}
	if repo == nil {
		repo = &stubDispatchRepo{}
	}

	verifier, sign := newTestSigner(t)
	injector := newTestInjector(t)

	dispatchSvc := service.NewDispatchService(repo, nil, nil, "", store)
	dispatchCtrl := controller.NewDispatchController(injector, dispatchSvc)

	router := gin.New()
	router.POST(
		"/api/trip/dispatch/customer/find-driver",
		middlewares.Authenticate(verifier, session.AlwaysActive()),
		middlewares.Authorize(&stubEnforcer{allow: allow}, constants.ENUM_ROLE_CUSTOMER, constants.ENUM_RESOURCE_DISPATCH, constants.ENUM_ACTION_READ),
		dispatchCtrl.FindDriver,
	)

	return router, sign, store
}

func newSetDriverModeRouter(t *testing.T, allow bool, store *drivergeo.Store) (*gin.Engine, func(userID, email, role string) string, *drivergeo.Store) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	if store == nil {
		store = newGeoStore(t)
	}

	verifier, sign := newTestSigner(t)
	injector := newTestInjector(t)

	dispatchSvc := service.NewDispatchService(nil, nil, nil, "", store)
	dispatchCtrl := controller.NewDispatchController(injector, dispatchSvc)

	router := gin.New()
	router.POST(
		"/api/trip/dispatch/driver/mode",
		middlewares.Authenticate(verifier, session.AlwaysActive()),
		middlewares.Authorize(&stubEnforcer{allow: allow}, constants.ENUM_ROLE_DRIVER, constants.ENUM_RESOURCE_DISPATCH, constants.ENUM_ACTION_UPDATE),
		dispatchCtrl.SetDriverMode,
	)

	return router, sign, store
}
