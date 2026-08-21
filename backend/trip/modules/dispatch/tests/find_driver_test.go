package tests

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/database/entities"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatch/dto"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatch/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindDriver_EmptyList(t *testing.T) {
	router, sign := newFindDriverRouter(t, true, nil)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/trip/dispatch/find-driver?current_location=-6.2088&current_location=106.8456&VehicleType=motorcycle",
		nil,
	)
	req.Header.Set("Authorization", "Bearer "+sign("user-1", "user@example.com", "customer"))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var body struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
		Data    struct {
			Drivers []struct {
				UserID    string     `json:"user_id"`
				DistanceM int        `json:"distance_m"`
				Location  [2]float64 `json:"location"`
			} `json:"drivers"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.True(t, body.Status)
	assert.Empty(t, body.Data.Drivers)
}

func TestFindDriver_ReturnsNearby(t *testing.T) {
	router, sign, store := newFindDriverRouterWithStore(t, true)
	require.NoError(t, store.SetStandby(context.Background(), "drv-1", -6.2088, 106.8456))

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/trip/dispatch/find-driver?current_location=-6.2088&current_location=106.8456&VehicleType=motorcycle",
		nil,
	)
	req.Header.Set("Authorization", "Bearer "+sign("user-1", "user@example.com", "customer"))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var body struct {
		Status bool `json:"status"`
		Data   struct {
			Drivers []struct {
				UserID    string     `json:"user_id"`
				DistanceM int        `json:"distance_m"`
				Location  [2]float64 `json:"location"`
			} `json:"drivers"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Len(t, body.Data.Drivers, 1)
	assert.Equal(t, "drv-1", body.Data.Drivers[0].UserID)
	assert.Equal(t, 0, body.Data.Drivers[0].DistanceM)
}

func TestFindDriver_DeniesWhenUnauthorized(t *testing.T) {
	router, sign := newFindDriverRouter(t, false, nil)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/trip/dispatch/find-driver?current_location=-6.2088&current_location=106.8456&VehicleType=motorcycle",
		nil,
	)
	req.Header.Set("Authorization", "Bearer "+sign("user-1", "drv@example.com", "driver"))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestFindDriverService_EmptyList(t *testing.T) {
	store := newGeoStore(t)
	svc := service.NewDispatchService(nil, nil, nil, "", store)

	result, err := svc.FindDriver(context.Background(), dto.FindDriverRequest{
		CurrentLocation: [2]string{"-6.2088", "106.8456"},
		VehicleType:     entities.VehicleTypeMotorcycle,
	})
	require.NoError(t, err)
	assert.Empty(t, result.Drivers)
}

func TestFindDriverService_ReturnsNearby(t *testing.T) {
	store := newGeoStore(t)
	require.NoError(t, store.SetStandby(context.Background(), "drv-1", -6.2088, 106.8456))

	svc := service.NewDispatchService(nil, nil, nil, "", store)
	result, err := svc.FindDriver(context.Background(), dto.FindDriverRequest{
		CurrentLocation: [2]string{"-6.2088", "106.8456"},
		VehicleType:     entities.VehicleTypeMotorcycle,
	})
	require.NoError(t, err)
	require.Len(t, result.Drivers, 1)
	assert.Equal(t, "drv-1", result.Drivers[0].UserID)
	assert.Equal(t, 0, result.Drivers[0].DistanceM)
	assert.InDelta(t, -6.2088, result.Drivers[0].Location[0], 0.0001)
	assert.InDelta(t, 106.8456, result.Drivers[0].Location[1], 0.0001)
}

func TestFindDriverService_InvalidLatLong(t *testing.T) {
	store := newGeoStore(t)
	svc := service.NewDispatchService(nil, nil, nil, "", store)

	_, err := svc.FindDriver(context.Background(), dto.FindDriverRequest{
		CurrentLocation: [2]string{"not-a-lat", "106.8456"},
		VehicleType:     entities.VehicleTypeMotorcycle,
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrInvalidLatLong))
}
