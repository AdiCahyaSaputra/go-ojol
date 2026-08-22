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
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/middlewares"
	dispatchcontroller "github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatch/controller"
	dispatchservice "github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatch/service"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatchws/controller"
	wsdto "github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatchws/dto"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatchws/service"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/constants"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/drivergeo"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/jwks"
	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/samber/do"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestServeWS_DeniesWhenUnauthorized(t *testing.T) {
	server, sign, _, _ := newDispatchWSServer(t, false)
	defer server.Close()

	_, resp, err := websocket.DefaultDialer.Dial(
		wsURL(server.URL),
		http.Header{"Authorization": []string{"Bearer " + sign("drv-1", "drv@example.com", "driver")}},
	)
	require.Error(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestServeWS_StandbyThenFindDriver(t *testing.T) {
	server, sign, _, rdb := newDispatchWSServer(t, true)
	defer server.Close()

	token := sign("drv-1", "drv@example.com", "driver")
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

	score, err := rdb.ZScore(context.Background(), drivergeo.KeyStandby, "drv-1").Result()
	require.NoError(t, err)
	assert.NotZero(t, score)

	customer := sign("cst-1", "cst@example.com", "customer")
	req, err := http.NewRequest(
		http.MethodGet,
		server.URL+"/api/trip/dispatch/customer/find-driver?current_location=-6.2088&current_location=106.8456&VehicleType=motorcycle",
		nil,
	)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+customer)
	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()
	require.Equal(t, http.StatusOK, res.StatusCode)

	var body struct {
		Data struct {
			Drivers []struct {
				UserID string `json:"user_id"`
			} `json:"drivers"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(res.Body).Decode(&body))
	require.Len(t, body.Data.Drivers, 1)
	assert.Equal(t, "drv-1", body.Data.Drivers[0].UserID)
}

func TestServeWS_AcceptsQueryToken(t *testing.T) {
	server, sign, _, _ := newDispatchWSServer(t, true)
	defer server.Close()

	token := sign("drv-1", "drv@example.com", "driver")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL(server.URL)+"?token="+url.QueryEscape(token), nil)
	require.NoError(t, err)
	conn.Close()
}

func TestServeWS_RemoveOnDisconnect(t *testing.T) {
	server, sign, _, rdb := newDispatchWSServer(t, true)
	defer server.Close()

	conn, _, err := websocket.DefaultDialer.Dial(
		wsURL(server.URL),
		http.Header{"Authorization": []string{"Bearer " + sign("drv-1", "drv@example.com", "driver")}},
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
		err := rdb.ZScore(context.Background(), drivergeo.KeyStandby, "drv-1").Err()
		return err == redis.Nil
	}, 2*time.Second, 20*time.Millisecond)
}

func TestServeWS_SecondConnectReplacesFirst(t *testing.T) {
	server, sign, _, rdb := newDispatchWSServer(t, true)
	defer server.Close()

	token := sign("drv-1", "drv@example.com", "driver")
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

	score, err := rdb.ZScore(context.Background(), drivergeo.KeyStandby, "drv-1").Result()
	require.NoError(t, err)
	assert.NotZero(t, score)
}

func newDispatchWSServer(t *testing.T, allow bool) (*httptest.Server, func(userID, email, role string) string, *drivergeo.Store, *redis.Client) {
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
	wsCtrl := controller.NewDispatchWSController(service.NewDispatchWSService(store))
	dispatchSvc := dispatchservice.NewDispatchService(nil, nil, nil, "", store)
	dispatchCtrl := dispatchcontroller.NewDispatchController(injector, dispatchSvc)

	router := gin.New()
	router.GET(
		"/api/trip/dispatch/ws",
		middlewares.Authenticate(verifier),
		middlewares.Authorize(&stubEnforcer{allow: allow}, constants.ENUM_ROLE_DRIVER, constants.ENUM_RESOURCE_TRIP, constants.ENUM_ACTION_UPDATE),
		wsCtrl.ServeWS,
	)
	router.GET(
		"/api/trip/dispatch/customer/find-driver",
		middlewares.Authenticate(verifier),
		middlewares.Authorize(&stubEnforcer{allow: true}, constants.ENUM_ROLE_CUSTOMER, constants.ENUM_RESOURCE_DISPATCH, constants.ENUM_ACTION_READ),
		dispatchCtrl.FindDriver,
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

	return httptest.NewServer(router), sign, store, rdb
}

func wsURL(httpURL string) string {
	return "ws" + strings.TrimPrefix(httpURL, "http") + "/api/trip/dispatch/ws"
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
