package tests

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/database/entities"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/middlewares"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/trip/controller"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/trip/dto"
	tripService "github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/trip/service"
	wsdto "github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatchws/dto"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/constants"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/jwks"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/session"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/triploc"
	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/samber/do"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubTripRepo struct {
	mu           sync.Mutex
	transactions map[uuid.UUID]*entities.Transaction
}

func tripTestCustomer() entities.Customer {
	return entities.Customer{
		ID:     uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		UserID: uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		Name:   "Test Customer",
	}
}

func tripTestDriver() entities.Driver {
	return entities.Driver{
		ID:        uuid.MustParse("55555555-5555-5555-5555-555555555555"),
		UserID:    uuid.MustParse("44444444-4444-4444-4444-444444444444"),
		VehicleID: uuid.MustParse("66666666-6666-6666-6666-666666666666"),
		Name:      "Test Driver",
	}
}

func newAcceptedTrip(customerID, driverID, vehicleID uuid.UUID) *entities.Transaction {
	return &entities.Transaction{
		ID:                  uuid.New(),
		CustomerID:          &customerID,
		DriverID:            &driverID,
		VehicleID:           &vehicleID,
		PickupLatLong:       []string{"-6.2088", "106.8456"},
		DestinationLatLong:  []string{"-6.1754", "106.8272"},
		DriverLastLatLong:   []string{"-6.2088", "106.8456"},
		CustomerLastLatLong: []string{"-6.2088", "106.8456"},
		Distance:            2000,
		TotalFare:           25000,
		Status:              entities.TransactionStatusAcceptedOffer,
		Customer: &entities.Customer{
			ID:     customerID,
			UserID: uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			Name:   "Test Customer",
		},
		Driver: &entities.Driver{
			ID:     driverID,
			UserID: uuid.MustParse("44444444-4444-4444-4444-444444444444"),
			Name:   "Test Driver",
		},
		Vehicle: &entities.Vehicle{
			ID:            vehicleID,
			Name:          "Test Vehicle",
			LicenseNumber: "B1234XX",
			Type:          entities.VehicleTypeMotorcycle,
		},
	}
}

func (s *stubTripRepo) seed(txn *entities.Transaction) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.transactions == nil {
		s.transactions = map[uuid.UUID]*entities.Transaction{}
	}
	s.transactions[txn.ID] = txn
}

func (s *stubTripRepo) ActiveByCustomerID(customerID uuid.UUID) (*entities.Transaction, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, txn := range s.transactions {
		if txn.CustomerID != nil && *txn.CustomerID == customerID &&
			(txn.Status == entities.TransactionStatusAcceptedOffer || txn.Status == entities.TransactionStatusOnTheWay) {
			return txn, nil
		}
	}
	return nil, errNoActiveTrip
}

func (s *stubTripRepo) ActiveByDriverID(driverID uuid.UUID) (*entities.Transaction, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, txn := range s.transactions {
		if txn.DriverID != nil && *txn.DriverID == driverID &&
			(txn.Status == entities.TransactionStatusAcceptedOffer || txn.Status == entities.TransactionStatusOnTheWay) {
			return txn, nil
		}
	}
	return nil, errNoActiveTrip
}

func (s *stubTripRepo) TransactionByID(txID uuid.UUID) (*entities.Transaction, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	txn, ok := s.transactions[txID]
	if !ok {
		return nil, errTripNotFound
	}
	return txn, nil
}

func (s *stubTripRepo) TransactionWithRelations(txID uuid.UUID) (*entities.Transaction, error) {
	return s.TransactionByID(txID)
}

func (s *stubTripRepo) UpdateStatus(txID uuid.UUID, from, to entities.TransactionStatus, extra map[string]any) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	txn, ok := s.transactions[txID]
	if !ok || txn.Status != from {
		return false, nil
	}
	txn.Status = to
	if paidAt, ok := extra["paid_at"].(time.Time); ok {
		txn.PaidAt = &paidAt
	}
	return true, nil
}

func (s *stubTripRepo) UpdateDriverLastLatLong(txID uuid.UUID, lat, lng string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	txn, ok := s.transactions[txID]
	if !ok {
		return errTripNotFound
	}
	txn.DriverLastLatLong = []string{lat, lng}
	return nil
}

func (s *stubTripRepo) UpdateCustomerLastLatLong(txID uuid.UUID, lat, lng string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	txn, ok := s.transactions[txID]
	if !ok {
		return errTripNotFound
	}
	txn.CustomerLastLatLong = []string{lat, lng}
	return nil
}

var (
	errTripNotFound = &tripSimpleError{msg: dto.MESSAGE_TRANSACTION_NOT_FOUND}
	errNoActiveTrip = &tripSimpleError{msg: dto.MESSAGE_NO_ACTIVE_TRANSACTION}
)

type tripSimpleError struct{ msg string }

func (e *tripSimpleError) Error() string { return e.msg }

type recordingNotifier struct {
	mu sync.Mutex
}

func (n *recordingNotifier) Notify(userID string, msg wsdto.ServerMessage) bool {
	return true
}

func tripTestSigner(t *testing.T) (jwks.Verifier, func(userID, email, role string) string) {
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

func injectTripCustomer() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set("customer", tripTestCustomer())
		ctx.Next()
	}
}

func injectTripDriver() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set("driver", tripTestDriver())
		ctx.Next()
	}
}

func injectTripProfileByRole() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		role, _ := ctx.Get("role")
		switch role {
		case "customer":
			injectTripCustomer()(ctx)
		case "driver":
			injectTripDriver()(ctx)
		default:
			ctx.Next()
		}
	}
}

func newTripRouter(t *testing.T, repo *stubTripRepo, allow bool) (*gin.Engine, func(userID, email, role string) string) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	mr := miniredis.RunT(t)
	store := triploc.NewStore(redis.NewClient(&redis.Options{Addr: mr.Addr()}))
	svc := tripService.NewTripService(repo, store, &recordingNotifier{})

	verifier, sign := tripTestSigner(t)
	injector := do.New()
	tripCtrl := controller.NewTripController(injector, svc)

	router := gin.New()
	auth := middlewares.Authenticate(verifier, session.AlwaysActive())
	read := middlewares.Authorize(&tripStubEnforcer{allow: allow}, "", constants.ENUM_RESOURCE_TRIP, constants.ENUM_ACTION_READ)
	update := middlewares.Authorize(&tripStubEnforcer{allow: allow}, "", constants.ENUM_RESOURCE_TRIP, constants.ENUM_ACTION_UPDATE)
	driverUpdate := middlewares.Authorize(&tripStubEnforcer{allow: allow}, constants.ENUM_ROLE_DRIVER, constants.ENUM_RESOURCE_TRIP, constants.ENUM_ACTION_UPDATE)

	router.GET("/api/trip/transactions/active", auth, injectTripProfileByRole(), read, tripCtrl.GetActive)
	router.POST("/api/trip/transactions/:id/start", auth, injectTripDriver(), driverUpdate, tripCtrl.StartTrip)
	router.POST("/api/trip/transactions/:id/complete", auth, injectTripDriver(), driverUpdate, tripCtrl.CompleteTrip)
	router.POST("/api/trip/transactions/:id/cancel", auth, injectTripProfileByRole(), update, tripCtrl.CancelTrip)

	return router, sign
}

type tripStubEnforcer struct {
	allow bool
}

func (s *tripStubEnforcer) Enforce(rvals ...interface{}) (bool, error) {
	return s.allow, nil
}

func (s *tripStubEnforcer) LoadPolicy() error {
	return nil
}

func TestGetActive_ReturnsTripForCustomer(t *testing.T) {
	repo := &stubTripRepo{}
	customer := tripTestCustomer()
	driver := tripTestDriver()
	txn := newAcceptedTrip(customer.ID, driver.ID, driver.VehicleID)
	repo.seed(txn)

	router, sign := newTripRouter(t, repo, true)
	raw := sign(customer.UserID.String(), "customer@example.com", "customer")

	req := httptest.NewRequest(http.MethodGet, "/api/trip/transactions/active", nil)
	req.Header.Set("Authorization", "Bearer "+raw)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var body struct {
		Data dto.TransactionResponse `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, txn.ID.String(), body.Data.ID)
	assert.Equal(t, string(entities.TransactionStatusAcceptedOffer), body.Data.Status)
}

func TestStartTrip_TransitionsToOnTheWay(t *testing.T) {
	repo := &stubTripRepo{}
	customer := tripTestCustomer()
	driver := tripTestDriver()
	txn := newAcceptedTrip(customer.ID, driver.ID, driver.VehicleID)
	repo.seed(txn)

	router, sign := newTripRouter(t, repo, true)
	raw := sign(driver.UserID.String(), "driver@example.com", "driver")

	req := httptest.NewRequest(http.MethodPost, "/api/trip/transactions/"+txn.ID.String()+"/start", nil)
	req.Header.Set("Authorization", "Bearer "+raw)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, entities.TransactionStatusOnTheWay, txn.Status)
}

func TestCompleteTrip_SetsPaidAtAndCompleted(t *testing.T) {
	repo := &stubTripRepo{}
	customer := tripTestCustomer()
	driver := tripTestDriver()
	txn := newAcceptedTrip(customer.ID, driver.ID, driver.VehicleID)
	txn.Status = entities.TransactionStatusOnTheWay
	repo.seed(txn)

	router, sign := newTripRouter(t, repo, true)
	raw := sign(driver.UserID.String(), "driver@example.com", "driver")

	req := httptest.NewRequest(http.MethodPost, "/api/trip/transactions/"+txn.ID.String()+"/complete", nil)
	req.Header.Set("Authorization", "Bearer "+raw)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, entities.TransactionStatusCompleted, txn.Status)
	require.NotNil(t, txn.PaidAt)
}

func TestStartTrip_RejectsWrongDriver(t *testing.T) {
	repo := &stubTripRepo{}
	customer := tripTestCustomer()
	driver := tripTestDriver()
	txn := newAcceptedTrip(customer.ID, driver.ID, driver.VehicleID)
	repo.seed(txn)

	router, sign := newTripRouter(t, repo, true)
	raw := sign(customer.UserID.String(), "customer@example.com", "customer")

	req := httptest.NewRequest(http.MethodPost, "/api/trip/transactions/"+txn.ID.String()+"/start", nil)
	req.Header.Set("Authorization", "Bearer "+raw)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestCompleteTrip_RejectsFromAcceptedOffer(t *testing.T) {
	repo := &stubTripRepo{}
	customer := tripTestCustomer()
	driver := tripTestDriver()
	txn := newAcceptedTrip(customer.ID, driver.ID, driver.VehicleID)
	repo.seed(txn)

	router, sign := newTripRouter(t, repo, true)
	raw := sign(driver.UserID.String(), "driver@example.com", "driver")

	req := httptest.NewRequest(http.MethodPost, "/api/trip/transactions/"+txn.ID.String()+"/complete", nil)
	req.Header.Set("Authorization", "Bearer "+raw)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)
}

func TestHandleDriverTripLocation_PersistsCoordinates(t *testing.T) {
	repo := &stubTripRepo{}
	customer := tripTestCustomer()
	driver := tripTestDriver()
	txn := newAcceptedTrip(customer.ID, driver.ID, driver.VehicleID)
	repo.seed(txn)

	mr := miniredis.RunT(t)
	store := triploc.NewStore(redis.NewClient(&redis.Options{Addr: mr.Addr()}))
	svc := tripService.NewTripService(repo, store, &recordingNotifier{})

	err := svc.HandleDriverTripLocation(context.Background(), driver.UserID.String(), txn.ID, -6.21, 106.84)
	require.NoError(t, err)
	require.Len(t, txn.DriverLastLatLong, 2)
	assert.Equal(t, "-6.21000000", txn.DriverLastLatLong[0])
	assert.Equal(t, "106.84000000", txn.DriverLastLatLong[1])

	coords, ok, err := store.GetDriver(context.Background(), txn.ID.String())
	require.NoError(t, err)
	require.True(t, ok)
	assert.InDelta(t, -6.21, coords.Lat, 0.0001)
}
