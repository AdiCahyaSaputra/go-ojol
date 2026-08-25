package tests

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/database/entities"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/middlewares"
	dispatchcontroller "github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatch/controller"
	dispatchdto "github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatch/dto"
	dispatchservice "github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatch/service"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatchws/controller"
	wsdto "github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatchws/dto"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatchws/service"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/constants"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/drivergeo"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/jwks"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/session"
	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/samber/do"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

const (
	testDriverUserID   = "44444444-4444-4444-4444-444444444444"
	testCustomerUserID = "33333333-3333-3333-3333-333333333333"
	testDriverID       = "55555555-5555-5555-5555-555555555555"
	testVehicleID      = "66666666-6666-6666-6666-666666666666"
	testCustomerID     = "11111111-1111-1111-1111-111111111111"
)

func TestServeWS_DeniesWhenUnauthorized(t *testing.T) {
	server, sign, _, _ := newDispatchWSServer(t, false, true)
	defer server.Close()

	_, resp, err := websocket.DefaultDialer.Dial(
		wsURL(server.URL),
		http.Header{"Authorization": []string{"Bearer " + sign(testDriverUserID, "drv@example.com", "driver")}},
	)
	require.Error(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestServeWS_StandbyThenFindDriver(t *testing.T) {
	server, sign, _, _ := newDispatchWSServer(t, true, true)
	defer server.Close()

	token := sign(testDriverUserID, "drv@example.com", "driver")
	conn, _, err := websocket.DefaultDialer.Dial(
		wsURL(server.URL),
		http.Header{"Authorization": []string{"Bearer " + token}},
	)
	require.NoError(t, err)
	defer conn.Close()

	require.NoError(t, conn.WriteJSON(wsdto.ClientMessage{
		Type: wsdto.TypeStandby,
		Lat:  -6.2088,
		Lng:  106.8456,
	}))

	var ack wsdto.ServerMessage
	require.NoError(t, conn.ReadJSON(&ack))
	assert.Equal(t, wsdto.TypeStandbyOK, ack.Type)

	txID := postFindDriver(t, server.URL, sign(testCustomerUserID, "cst@example.com", "customer"))

	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	var offerMsg wsdto.ServerMessage
	require.NoError(t, conn.ReadJSON(&offerMsg))
	assert.Equal(t, wsdto.TypeTripOffer, offerMsg.Type)
	require.NotNil(t, offerMsg.Offer)
	assert.Equal(t, txID, offerMsg.Offer.TransactionID)
	assert.Equal(t, "Test Customer", offerMsg.Offer.CustomerName)
}

func TestServeWS_RejectsQueryToken(t *testing.T) {
	server, sign, _, _ := newDispatchWSServer(t, true, true)
	defer server.Close()

	token := sign(testDriverUserID, "drv@example.com", "driver")
	_, resp, err := websocket.DefaultDialer.Dial(wsURL(server.URL)+"?token="+url.QueryEscape(token), nil)
	require.Error(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestServeWS_RemoveOnDisconnect(t *testing.T) {
	server, sign, _, rdb := newDispatchWSServer(t, true, true)
	defer server.Close()

	conn, _, err := websocket.DefaultDialer.Dial(
		wsURL(server.URL),
		http.Header{"Authorization": []string{"Bearer " + sign(testDriverUserID, "drv@example.com", "driver")}},
	)
	require.NoError(t, err)

	require.NoError(t, conn.WriteJSON(wsdto.ClientMessage{
		Type: wsdto.TypeStandby,
		Lat:  -6.2088,
		Lng:  106.8456,
	}))
	var ack wsdto.ServerMessage
	require.NoError(t, conn.ReadJSON(&ack))
	require.Equal(t, wsdto.TypeStandbyOK, ack.Type)

	require.NoError(t, conn.Close())
	require.Eventually(t, func() bool {
		err := rdb.ZScore(context.Background(), drivergeo.KeyStandby, testDriverUserID).Err()
		return err == redis.Nil
	}, 2*time.Second, 20*time.Millisecond)
}

func TestServeWS_SecondConnectReplacesFirst(t *testing.T) {
	server, sign, _, rdb := newDispatchWSServer(t, true, true)
	defer server.Close()

	token := sign(testDriverUserID, "drv@example.com", "driver")
	first, _, err := websocket.DefaultDialer.Dial(
		wsURL(server.URL),
		http.Header{"Authorization": []string{"Bearer " + token}},
	)
	require.NoError(t, err)

	require.NoError(t, first.WriteJSON(wsdto.ClientMessage{
		Type: wsdto.TypeStandby,
		Lat:  -6.2088,
		Lng:  106.8456,
	}))
	var ack wsdto.ServerMessage
	require.NoError(t, first.ReadJSON(&ack))
	require.Equal(t, wsdto.TypeStandbyOK, ack.Type)

	second, _, err := websocket.DefaultDialer.Dial(
		wsURL(server.URL),
		http.Header{"Authorization": []string{"Bearer " + token}},
	)
	require.NoError(t, err)
	defer second.Close()

	require.NoError(t, second.WriteJSON(wsdto.ClientMessage{
		Type: wsdto.TypeStandby,
		Lat:  -6.2090,
		Lng:  106.8458,
	}))
	require.NoError(t, second.ReadJSON(&ack))
	require.Equal(t, wsdto.TypeStandbyOK, ack.Type)

	_ = first.Close()
	time.Sleep(100 * time.Millisecond)

	score, err := rdb.ZScore(context.Background(), drivergeo.KeyStandby, testDriverUserID).Result()
	require.NoError(t, err)
	assert.NotZero(t, score)
}

func TestCustomerWS_DeniesDriverToken(t *testing.T) {
	server, sign, _, _ := newDispatchWSServer(t, true, true)
	defer server.Close()

	_, resp, err := websocket.DefaultDialer.Dial(
		customerWSURL(server.URL),
		http.Header{"Authorization": []string{"Bearer " + sign(testDriverUserID, "drv@example.com", "driver")}},
	)
	require.Error(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestCustomerWS_DeniesWhenUnauthorized(t *testing.T) {
	server, sign, _, _ := newDispatchWSServer(t, true, false)
	defer server.Close()

	_, resp, err := websocket.DefaultDialer.Dial(
		customerWSURL(server.URL),
		http.Header{"Authorization": []string{"Bearer " + sign(testCustomerUserID, "cst@example.com", "customer")}},
	)
	require.Error(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestCustomerWS_WaitingThenDriverMatched(t *testing.T) {
	server, sign, _, _ := newDispatchWSServer(t, true, true)
	defer server.Close()

	driverConn := dialDriverStandby(t, server.URL, sign)
	defer driverConn.Close()

	customerConn := dialCustomerWS(t, server.URL, sign)
	defer customerConn.Close()

	txID := postFindDriver(t, server.URL, sign(testCustomerUserID, "cst@example.com", "customer"))

	waiting := readWSMessage(t, customerConn, 2*time.Second)
	assert.Equal(t, wsdto.TypeWaiting, waiting.Type)
	assert.Equal(t, txID, waiting.TransactionID)

	offer := readWSMessage(t, driverConn, 2*time.Second)
	assert.Equal(t, wsdto.TypeTripOffer, offer.Type)

	postRespondOffer(t, server.URL, sign(testDriverUserID, "drv@example.com", "driver"), txID, "accept")

	matched := readWSMessage(t, customerConn, 2*time.Second)
	assert.Equal(t, wsdto.TypeDriverMatched, matched.Type)
	assert.Equal(t, txID, matched.TransactionID)
	require.NotNil(t, matched.MatchedDriver)
	assert.Equal(t, testDriverUserID, matched.MatchedDriver.UserID)
	assert.Equal(t, "Test Driver", matched.MatchedDriver.Name)
}

func TestCustomerWS_OfferExpired(t *testing.T) {
	server, sign, _, _ := newDispatchWSServer(t, true, true, dispatchservice.WithOfferTTL(50*time.Millisecond))
	defer server.Close()

	driverConn := dialDriverStandby(t, server.URL, sign)
	defer driverConn.Close()

	customerConn := dialCustomerWS(t, server.URL, sign)
	defer customerConn.Close()

	txID := postFindDriver(t, server.URL, sign(testCustomerUserID, "cst@example.com", "customer"))
	_ = readWSMessage(t, customerConn, 2*time.Second) // waiting
	_ = readWSMessage(t, driverConn, 2*time.Second)   // trip_offer

	expired := readWSMessage(t, customerConn, 2*time.Second)
	assert.Equal(t, wsdto.TypeOfferExpired, expired.Type)
	assert.Equal(t, txID, expired.TransactionID)
}

func TestCustomerWS_OfferRejected(t *testing.T) {
	server, sign, _, _ := newDispatchWSServer(t, true, true)
	defer server.Close()

	driverConn := dialDriverStandby(t, server.URL, sign)
	defer driverConn.Close()

	customerConn := dialCustomerWS(t, server.URL, sign)
	defer customerConn.Close()

	txID := postFindDriver(t, server.URL, sign(testCustomerUserID, "cst@example.com", "customer"))
	_ = readWSMessage(t, customerConn, 2*time.Second)
	_ = readWSMessage(t, driverConn, 2*time.Second)

	postRespondOffer(t, server.URL, sign(testDriverUserID, "drv@example.com", "driver"), txID, "reject")

	rejected := readWSMessage(t, customerConn, 2*time.Second)
	assert.Equal(t, wsdto.TypeOfferRejected, rejected.Type)
	assert.Equal(t, txID, rejected.TransactionID)
}

func TestCustomerWS_RetryAfterExpire(t *testing.T) {
	server, sign, _, _ := newDispatchWSServer(t, true, true, dispatchservice.WithOfferTTL(50*time.Millisecond))
	defer server.Close()

	driverConn := dialDriverStandby(t, server.URL, sign)
	defer driverConn.Close()

	customerConn := dialCustomerWS(t, server.URL, sign)
	defer customerConn.Close()

	_ = postFindDriver(t, server.URL, sign(testCustomerUserID, "cst@example.com", "customer"))
	_ = readWSMessage(t, customerConn, 2*time.Second) // waiting
	_ = readWSMessage(t, driverConn, 2*time.Second)   // trip_offer
	_ = readWSMessage(t, customerConn, 2*time.Second) // offer_expired
	_ = readWSMessage(t, driverConn, 2*time.Second)   // offer_expired

	require.NoError(t, customerConn.WriteJSON(wsdto.ClientMessage{Type: wsdto.TypeRetry}))

	waiting := readWSMessage(t, customerConn, 2*time.Second)
	assert.Equal(t, wsdto.TypeWaiting, waiting.Type)
	require.NotEmpty(t, waiting.TransactionID)

	offer := readWSMessage(t, driverConn, 2*time.Second)
	assert.Equal(t, wsdto.TypeTripOffer, offer.Type)
	assert.Equal(t, waiting.TransactionID, offer.TransactionID)
}

func TestCustomerWS_RetryWhileActiveReturnsError(t *testing.T) {
	server, sign, _, _ := newDispatchWSServer(t, true, true)
	defer server.Close()

	driverConn := dialDriverStandby(t, server.URL, sign)
	defer driverConn.Close()

	customerConn := dialCustomerWS(t, server.URL, sign)
	defer customerConn.Close()

	_ = postFindDriver(t, server.URL, sign(testCustomerUserID, "cst@example.com", "customer"))
	_ = readWSMessage(t, customerConn, 2*time.Second)
	_ = readWSMessage(t, driverConn, 2*time.Second)

	require.NoError(t, customerConn.WriteJSON(wsdto.ClientMessage{Type: wsdto.TypeRetry}))

	errMsg := readWSMessage(t, customerConn, 2*time.Second)
	assert.Equal(t, wsdto.TypeError, errMsg.Type)
	assert.Equal(t, dispatchdto.MESSAGE_OFFER_STILL_ACTIVE, errMsg.Message)
}

func TestCustomerWS_RetryWithoutPriorSearch(t *testing.T) {
	server, sign, _, _ := newDispatchWSServer(t, true, true)
	defer server.Close()

	customerConn := dialCustomerWS(t, server.URL, sign)
	defer customerConn.Close()

	require.NoError(t, customerConn.WriteJSON(wsdto.ClientMessage{Type: wsdto.TypeRetry}))

	errMsg := readWSMessage(t, customerConn, 2*time.Second)
	assert.Equal(t, wsdto.TypeError, errMsg.Type)
	assert.Equal(t, dispatchdto.MESSAGE_NO_LAST_SEARCH, errMsg.Message)
}

func TestCustomerWS_RetryNoDrivers(t *testing.T) {
	server, sign, store, _ := newDispatchWSServer(t, true, true, dispatchservice.WithOfferTTL(50*time.Millisecond))
	defer server.Close()

	driverConn := dialDriverStandby(t, server.URL, sign)
	defer driverConn.Close()

	customerConn := dialCustomerWS(t, server.URL, sign)
	defer customerConn.Close()

	_ = postFindDriver(t, server.URL, sign(testCustomerUserID, "cst@example.com", "customer"))
	_ = readWSMessage(t, customerConn, 2*time.Second)
	_ = readWSMessage(t, driverConn, 2*time.Second)
	_ = readWSMessage(t, customerConn, 2*time.Second) // expired
	_ = readWSMessage(t, driverConn, 2*time.Second)

	require.NoError(t, store.RemoveStandby(context.Background(), testDriverUserID))
	require.NoError(t, customerConn.WriteJSON(wsdto.ClientMessage{Type: wsdto.TypeRetry}))

	msg := readWSMessage(t, customerConn, 2*time.Second)
	assert.Equal(t, wsdto.TypeNoDrivers, msg.Type)
}

func dialDriverStandby(t *testing.T, httpURL string, sign func(string, string, string) string) *websocket.Conn {
	t.Helper()
	conn, _, err := websocket.DefaultDialer.Dial(
		wsURL(httpURL),
		http.Header{"Authorization": []string{"Bearer " + sign(testDriverUserID, "drv@example.com", "driver")}},
	)
	require.NoError(t, err)
	require.NoError(t, conn.WriteJSON(wsdto.ClientMessage{
		Type: wsdto.TypeStandby,
		Lat:  -6.2088,
		Lng:  106.8456,
	}))
	var ack wsdto.ServerMessage
	require.NoError(t, conn.ReadJSON(&ack))
	require.Equal(t, wsdto.TypeStandbyOK, ack.Type)
	return conn
}

func dialCustomerWS(t *testing.T, httpURL string, sign func(string, string, string) string) *websocket.Conn {
	t.Helper()
	conn, _, err := websocket.DefaultDialer.Dial(
		customerWSURL(httpURL),
		http.Header{"Authorization": []string{"Bearer " + sign(testCustomerUserID, "cst@example.com", "customer")}},
	)
	require.NoError(t, err)
	return conn
}

func readWSMessage(t *testing.T, conn *websocket.Conn, timeout time.Duration) wsdto.ServerMessage {
	t.Helper()
	_ = conn.SetReadDeadline(time.Now().Add(timeout))
	var msg wsdto.ServerMessage
	require.NoError(t, conn.ReadJSON(&msg))
	return msg
}

func postFindDriver(t *testing.T, httpURL, token string) string {
	t.Helper()
	reqBody, err := json.Marshal(map[string]any{
		"pickup_lat_long":      []string{"-6.2088", "106.8456"},
		"destination_lat_long": []string{"-6.1754", "106.8272"},
		"vehicle_type":         "motorcycle",
		"max_size":             1,
	})
	require.NoError(t, err)
	req, err := http.NewRequest(
		http.MethodPost,
		httpURL+"/api/trip/dispatch/customer/find-driver",
		bytes.NewReader(reqBody),
	)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()
	require.Equal(t, http.StatusOK, res.StatusCode)

	var body struct {
		Data struct {
			TransactionID string `json:"transaction_id"`
			Drivers       []any  `json:"drivers"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(res.Body).Decode(&body))
	require.NotEmpty(t, body.Data.Drivers)
	require.NotEmpty(t, body.Data.TransactionID)
	return body.Data.TransactionID
}

func postRespondOffer(t *testing.T, httpURL, token, txID, action string) {
	t.Helper()
	reqBody, err := json.Marshal(map[string]string{"action": action})
	require.NoError(t, err)
	req, err := http.NewRequest(
		http.MethodPost,
		httpURL+"/api/trip/dispatch/driver/offers/"+txID+"/respond",
		bytes.NewReader(reqBody),
	)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()
	require.Equal(t, http.StatusOK, res.StatusCode)
}

func newDispatchWSServer(
	t *testing.T,
	driverAllow bool,
	customerAllow bool,
	opts ...dispatchservice.Option,
) (*httptest.Server, func(userID, email, role string) string, *drivergeo.Store, *redis.Client) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	store := drivergeo.NewStore(rdb)

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

	verifier := jwks.NewVerifier(jwksServer.URL, "go-ojol-auth")
	wsSvc := service.NewDispatchWSService(store)
	wsCtrl := controller.NewDispatchWSController(wsSvc)

	osrmServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"code": "Ok",
			"routes": [{
				"distance": 2000,
				"duration": 180,
				"geometry": {
					"type": "LineString",
					"coordinates": [[106.8456, -6.2088], [106.8272, -6.1754]]
				}
			}]
		}`))
	}))
	t.Cleanup(osrmServer.Close)

	repo := &wsStubDispatchRepo{
		nearbyProfiles: map[string]dispatchdto.NearbyDriverProfile{
			testDriverUserID: {
				UserID:        uuid.MustParse(testDriverUserID),
				DriverID:      uuid.MustParse(testDriverID),
				Name:          "Test Driver",
				PhoneNumber:   "081234567890",
				VehicleID:     uuid.MustParse(testVehicleID),
				VehicleName:   "Test Motorcycle",
				LicenseNumber: "B1234XX",
				MaxSize:       2,
				Type:          "motorcycle",
			},
		},
		vehicle: &entities.Vehicle{
			ID:            uuid.MustParse(testVehicleID),
			Name:          "Test Motorcycle",
			LicenseNumber: "B1234XX",
			MaxSize:       2,
			Type:          entities.VehicleTypeMotorcycle,
		},
	}
	dispatchSvc := dispatchservice.NewDispatchService(
		repo,
		nil,
		osrmServer.Client(),
		osrmServer.URL,
		store,
		wsSvc,
		opts...,
	)
	wsSvc.SetOfferRetrier(dispatchSvc)
	dispatchCtrl := dispatchcontroller.NewDispatchController(injector, dispatchSvc)

	router := gin.New()
	router.GET(
		"/api/trip/dispatch/ws",
		middlewares.AuthenticateWS(verifier, session.AlwaysActive()),
		middlewares.Authorize(&stubEnforcer{allow: driverAllow}, constants.ENUM_ROLE_DRIVER, constants.ENUM_RESOURCE_TRIP, constants.ENUM_ACTION_UPDATE),
		wsCtrl.ServeWS,
	)
	router.GET(
		"/api/trip/dispatch/customer/ws",
		middlewares.AuthenticateWS(verifier, session.AlwaysActive()),
		middlewares.Authorize(&stubEnforcer{allow: customerAllow}, constants.ENUM_ROLE_CUSTOMER, constants.ENUM_RESOURCE_DISPATCH, constants.ENUM_ACTION_CREATE),
		wsCtrl.ServeCustomerWS,
	)
	router.POST(
		"/api/trip/dispatch/customer/find-driver",
		middlewares.Authenticate(verifier, session.AlwaysActive()),
		middlewares.Authorize(&stubEnforcer{allow: true}, constants.ENUM_ROLE_CUSTOMER, constants.ENUM_RESOURCE_DISPATCH, constants.ENUM_ACTION_CREATE),
		func(ctx *gin.Context) {
			ctx.Set("customer", entities.Customer{
				ID:     uuid.MustParse(testCustomerID),
				UserID: uuid.MustParse(testCustomerUserID),
				Name:   "Test Customer",
			})
			ctx.Next()
		},
		dispatchCtrl.FindDriver,
	)
	router.POST(
		"/api/trip/dispatch/driver/offers/:transaction_id/respond",
		middlewares.Authenticate(verifier, session.AlwaysActive()),
		middlewares.Authorize(&stubEnforcer{allow: true}, constants.ENUM_ROLE_DRIVER, constants.ENUM_RESOURCE_DISPATCH, constants.ENUM_ACTION_UPDATE),
		func(ctx *gin.Context) {
			ctx.Set("driver", entities.Driver{
				ID:          uuid.MustParse(testDriverID),
				UserID:      uuid.MustParse(testDriverUserID),
				VehicleID:   uuid.MustParse(testVehicleID),
				Name:        "Test Driver",
				PhoneNumber: "081234567890",
			})
			ctx.Next()
		},
		dispatchCtrl.RespondOffer,
	)

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

	return httptest.NewServer(router), sign, store, rdb
}

func wsURL(httpURL string) string {
	return "ws" + strings.TrimPrefix(httpURL, "http") + "/api/trip/dispatch/ws"
}

func customerWSURL(httpURL string) string {
	return "ws" + strings.TrimPrefix(httpURL, "http") + "/api/trip/dispatch/customer/ws"
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

type wsStubDispatchRepo struct {
	mu             sync.Mutex
	nearbyProfiles map[string]dispatchdto.NearbyDriverProfile
	vehicle        *entities.Vehicle
	created        *entities.Transaction
	statusByID     map[uuid.UUID]entities.TransactionStatus
}

func (s *wsStubDispatchRepo) VehicleById(id uuid.UUID) (*entities.Vehicle, error) {
	if s.vehicle == nil || s.vehicle.ID != id {
		return nil, nil
	}
	return s.vehicle, nil
}

func (s *wsStubDispatchRepo) DistinctVehicleCategories() ([]dispatchdto.VehicleCategory, error) {
	return []dispatchdto.VehicleCategory{
		{VehicleType: entities.VehicleTypeMotorcycle, MaxSize: 1},
	}, nil
}

func (s *wsStubDispatchRepo) NearbyDriverProfiles(_ []uuid.UUID, _ entities.VehicleType) (map[string]dispatchdto.NearbyDriverProfile, error) {
	return s.nearbyProfiles, nil
}

func (s *wsStubDispatchRepo) CreateOfferedTransaction(req dispatchdto.CreateOfferedTransaction) (*entities.Transaction, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	txn := &entities.Transaction{
		ID:                 uuid.New(),
		CustomerID:         &req.CustomerID,
		PickupLatLong:      req.PickupLatLong[:],
		DestinationLatLong: req.DestinationLatLong[:],
		LastLatLong:        req.LastLatLong[:],
		Distance:           req.Distance,
		FarePerDistance:    req.FarePerDistance,
		PlatformPercentage: req.PlatformPercentage,
		TotalFare:          req.TotalFare,
		Status:             entities.TransactionStatusOffered,
	}
	s.created = txn
	if s.statusByID == nil {
		s.statusByID = map[uuid.UUID]entities.TransactionStatus{}
	}
	s.statusByID[txn.ID] = entities.TransactionStatusOffered
	return txn, nil
}

func (s *wsStubDispatchRepo) ClaimOffer(txID, driverID, vehicleID uuid.UUID) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.statusByID == nil {
		return false, nil
	}
	status, ok := s.statusByID[txID]
	if !ok || status != entities.TransactionStatusOffered {
		return false, nil
	}
	s.statusByID[txID] = entities.TransactionStatusAcceptedOffer
	return true, nil
}

func (s *wsStubDispatchRepo) ExpireOffer(txID uuid.UUID) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.statusByID == nil {
		return false, nil
	}
	status, ok := s.statusByID[txID]
	if !ok || status != entities.TransactionStatusOffered {
		return false, nil
	}
	s.statusByID[txID] = entities.TransactionStatusExpired
	return true, nil
}

func (s *wsStubDispatchRepo) MarkRejectedOffer(txID uuid.UUID) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.statusByID == nil {
		return false, nil
	}
	status, ok := s.statusByID[txID]
	if !ok || status != entities.TransactionStatusOffered {
		return false, nil
	}
	s.statusByID[txID] = entities.TransactionStatusRejectedOffer
	return true, nil
}

func (s *wsStubDispatchRepo) TransactionByID(txID uuid.UUID) (*entities.Transaction, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	status, ok := s.statusByID[txID]
	if !ok {
		return nil, errors.New(dispatchdto.MESSAGE_OFFER_NOT_FOUND)
	}
	return &entities.Transaction{ID: txID, Status: status}, nil
}
