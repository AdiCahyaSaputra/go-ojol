package tests

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/database/entities"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/middlewares"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/saved_address/controller"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/saved_address/repository"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/saved_address/service"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/constants"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/jwks"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/session"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/samber/do"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
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

type memorySavedAddressRepo struct {
	mu        sync.Mutex
	addresses map[uuid.UUID]entities.SavedAddress
}

func newMemorySavedAddressRepo() *memorySavedAddressRepo {
	return &memorySavedAddressRepo{
		addresses: map[uuid.UUID]entities.SavedAddress{},
	}
}

func (r *memorySavedAddressRepo) ListByCustomer(_ context.Context, _ *gorm.DB, customerID uuid.UUID) ([]entities.SavedAddress, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]entities.SavedAddress, 0)
	for _, address := range r.addresses {
		if address.CustomerID == customerID {
			out = append(out, address)
		}
	}
	return out, nil
}

func (r *memorySavedAddressRepo) GetByIDAndCustomer(_ context.Context, _ *gorm.DB, id, customerID uuid.UUID) (entities.SavedAddress, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	address, ok := r.addresses[id]
	if !ok || address.CustomerID != customerID {
		return entities.SavedAddress{}, gorm.ErrRecordNotFound
	}
	return address, nil
}

func (r *memorySavedAddressRepo) Create(_ context.Context, _ *gorm.DB, address entities.SavedAddress) (entities.SavedAddress, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if address.ID == uuid.Nil {
		address.ID = uuid.New()
	}
	now := time.Now().UTC()
	address.CreatedAt = now
	address.UpdatedAt = now
	r.addresses[address.ID] = address
	return address, nil
}

func (r *memorySavedAddressRepo) Update(_ context.Context, _ *gorm.DB, address entities.SavedAddress) (entities.SavedAddress, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	existing, ok := r.addresses[address.ID]
	if !ok {
		return entities.SavedAddress{}, gorm.ErrRecordNotFound
	}
	address.CreatedAt = existing.CreatedAt
	address.UpdatedAt = time.Now().UTC()
	r.addresses[address.ID] = address
	return address, nil
}

func (r *memorySavedAddressRepo) Delete(_ context.Context, _ *gorm.DB, id, customerID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	address, ok := r.addresses[id]
	if !ok || address.CustomerID != customerID {
		return gorm.ErrRecordNotFound
	}
	delete(r.addresses, id)
	return nil
}

func (r *memorySavedAddressRepo) ClearDefaultPickup(_ context.Context, _ *gorm.DB, customerID uuid.UUID, exceptID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for id, address := range r.addresses {
		if address.CustomerID != customerID || !address.IsDefaultPickup {
			continue
		}
		if exceptID != uuid.Nil && id == exceptID {
			continue
		}
		address.IsDefaultPickup = false
		r.addresses[id] = address
	}
	return nil
}

var _ repository.SavedAddressRepository = (*memorySavedAddressRepo)(nil)

func testCustomer() entities.Customer {
	return entities.Customer{
		ID:     uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		UserID: uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		Name:   "Test Customer",
	}
}

func otherCustomer() entities.Customer {
	return entities.Customer{
		ID:     uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
		UserID: uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
		Name:   "Other Customer",
	}
}

func injectCustomer(customer entities.Customer) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set("customer", customer)
		ctx.Next()
	}
}

func newTestSigner(t *testing.T) (jwks.Verifier, func(userID, email, role string) string) {
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

	verifier := jwks.NewVerifier(jwksServer.URL, "go-ojol-auth")
	sign := func(userID, email, role string) string {
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

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	require.NoError(t, err)
	return db
}

func newTestInjector(t *testing.T, db *gorm.DB) *do.Injector {
	t.Helper()
	injector := do.New()
	do.ProvideNamed(injector, constants.DB, func(i *do.Injector) (*gorm.DB, error) {
		return db, nil
	})
	return injector
}

type testRouter struct {
	router *gin.Engine
	sign   func(userID, email, role string) string
	repo   *memorySavedAddressRepo
}

func newSavedAddressRouter(t *testing.T, allow bool, customer entities.Customer) testRouter {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db := newTestDB(t)
	verifier, sign := newTestSigner(t)
	injector := newTestInjector(t, db)
	repo := newMemorySavedAddressRepo()
	svc := service.NewSavedAddressService(repo, db)
	ctrl := controller.NewSavedAddressController(injector, svc)

	authn := middlewares.Authenticate(verifier, session.AlwaysActive())
	role := constants.ENUM_ROLE_CUSTOMER
	resource := constants.ENUM_RESOURCE_SAVED_ADDRESS
	inject := injectCustomer(customer)

	router := gin.New()
	group := router.Group("/api/trip/saved-addresses")
	{
		group.GET("",
			authn,
			middlewares.Authorize(&stubEnforcer{allow: allow}, role, resource, constants.ENUM_ACTION_READ),
			inject,
			ctrl.List,
		)
		group.GET("/:id",
			authn,
			middlewares.Authorize(&stubEnforcer{allow: allow}, role, resource, constants.ENUM_ACTION_READ),
			inject,
			ctrl.GetByID,
		)
		group.POST("",
			authn,
			middlewares.Authorize(&stubEnforcer{allow: allow}, role, resource, constants.ENUM_ACTION_CREATE),
			inject,
			ctrl.Create,
		)
		group.PUT("/:id",
			authn,
			middlewares.Authorize(&stubEnforcer{allow: allow}, role, resource, constants.ENUM_ACTION_UPDATE),
			inject,
			ctrl.Update,
		)
		group.DELETE("/:id",
			authn,
			middlewares.Authorize(&stubEnforcer{allow: allow}, role, resource, constants.ENUM_ACTION_DELETE),
			inject,
			ctrl.Delete,
		)
	}

	return testRouter{router: router, sign: sign, repo: repo}
}

func seedAddress(repo *memorySavedAddressRepo, customerID uuid.UUID, name string, isDefault bool) entities.SavedAddress {
	address := entities.SavedAddress{
		ID:              uuid.New(),
		CustomerID:      customerID,
		Name:            name,
		LatLong:         pq.StringArray{"-6.2088", "106.8456"},
		IsDefaultPickup: isDefault,
		Timestamp: entities.Timestamp{
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		},
	}
	repo.addresses[address.ID] = address
	return address
}

func authHeader(sign func(userID, email, role string) string, customer entities.Customer) string {
	return "Bearer " + sign(customer.UserID.String(), "cst@example.com", constants.ENUM_ROLE_CUSTOMER)
}
