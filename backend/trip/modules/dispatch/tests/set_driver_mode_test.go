package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatch/dto"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatch/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetDriverMode_RequiresBearer(t *testing.T) {
	router, _, _ := newSetDriverModeRouter(t, true, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/trip/dispatch/driver/mode", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestSetDriverMode_RejectsInvalidBody(t *testing.T) {
	router, sign, _ := newSetDriverModeRouter(t, true, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/trip/dispatch/driver/mode", bytes.NewBufferString(`{
		"mode": "busy",
		"current_lat_long": ["-6.2088", "106.8456"]
	}`))
	req.Header.Set("Authorization", "Bearer "+sign("drv-1", "drv@example.com", "driver"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestSetDriverMode_OnlineSetsStandby(t *testing.T) {
	router, sign, store := newSetDriverModeRouter(t, true, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/trip/dispatch/driver/mode", bytes.NewBufferString(`{
		"mode": "online",
		"current_lat_long": ["-6.2088", "106.8456"]
	}`))
	req.Header.Set("Authorization", "Bearer "+sign("drv-1", "drv@example.com", "driver"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var body struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.True(t, body.Status)
	assert.Equal(t, dto.MESSAGE_SET_DRIVER_MODE_SUCCESS, body.Message)

	nearby, err := store.Nearby(context.Background(), -6.2088, 106.8456, 5, 10)
	require.NoError(t, err)
	require.Len(t, nearby, 1)
	assert.Equal(t, "drv-1", nearby[0].UserID)
}

func TestSetDriverMode_OfflineRemovesStandby(t *testing.T) {
	router, sign, store := newSetDriverModeRouter(t, true, nil)
	require.NoError(t, store.SetStandby(context.Background(), "drv-1", -6.2088, 106.8456))

	req := httptest.NewRequest(http.MethodPost, "/api/trip/dispatch/driver/mode", bytes.NewBufferString(`{
		"mode": "offline",
		"current_lat_long": ["-6.2088", "106.8456"]
	}`))
	req.Header.Set("Authorization", "Bearer "+sign("drv-1", "drv@example.com", "driver"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	nearby, err := store.Nearby(context.Background(), -6.2088, 106.8456, 5, 10)
	require.NoError(t, err)
	assert.Empty(t, nearby)
}

func TestSetDriverMode_DeniesWhenUnauthorized(t *testing.T) {
	router, sign, _ := newSetDriverModeRouter(t, false, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/trip/dispatch/driver/mode", bytes.NewBufferString(`{
		"mode": "online",
		"current_lat_long": ["-6.2088", "106.8456"]
	}`))
	req.Header.Set("Authorization", "Bearer "+sign("drv-1", "drv@example.com", "driver"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestSetDriverModeService_Online(t *testing.T) {
	store := newGeoStore(t)
	svc := service.NewDispatchService(nil, nil, nil, "", store, nil)
	ctx := context.WithValue(context.Background(), "user_id", "drv-1")

	err := svc.SetDriverMode(ctx, dto.SetDriverModeRequest{
		Mode:           dto.DriverModeOnline,
		CurrentLatLong: [2]string{"-6.2088", "106.8456"},
	})
	require.NoError(t, err)

	nearby, err := store.Nearby(context.Background(), -6.2088, 106.8456, 5, 10)
	require.NoError(t, err)
	require.Len(t, nearby, 1)
	assert.Equal(t, "drv-1", nearby[0].UserID)
	assert.InDelta(t, -6.2088, nearby[0].Lat, 0.0001)
	assert.InDelta(t, 106.8456, nearby[0].Lng, 0.0001)
}

func TestSetDriverModeService_Offline(t *testing.T) {
	store := newGeoStore(t)
	require.NoError(t, store.SetStandby(context.Background(), "drv-1", -6.2088, 106.8456))

	svc := service.NewDispatchService(nil, nil, nil, "", store, nil)
	ctx := context.WithValue(context.Background(), "user_id", "drv-1")

	err := svc.SetDriverMode(ctx, dto.SetDriverModeRequest{
		Mode:           dto.DriverModeOffline,
		CurrentLatLong: [2]string{"-6.2088", "106.8456"},
	})
	require.NoError(t, err)

	nearby, err := store.Nearby(context.Background(), -6.2088, 106.8456, 5, 10)
	require.NoError(t, err)
	assert.Empty(t, nearby)
}

func TestSetDriverModeService_InvalidLatLong(t *testing.T) {
	store := newGeoStore(t)
	svc := service.NewDispatchService(nil, nil, nil, "", store, nil)
	ctx := context.WithValue(context.Background(), "user_id", "drv-1")

	err := svc.SetDriverMode(ctx, dto.SetDriverModeRequest{
		Mode:           dto.DriverModeOnline,
		CurrentLatLong: [2]string{"not-a-lat", "106.8456"},
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrInvalidLatLong))
}
