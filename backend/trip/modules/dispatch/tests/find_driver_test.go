package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/database/entities"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatch/dto"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatch/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testDriverUserID = "44444444-4444-4444-4444-444444444444"
	testDriverID     = "55555555-5555-5555-5555-555555555555"
	testVehicleID    = "66666666-6666-6666-6666-666666666666"
)

func testNearbyDriverProfile() dto.NearbyDriverProfile {
	return dto.NearbyDriverProfile{
		UserID:        uuid.MustParse(testDriverUserID),
		DriverID:      uuid.MustParse(testDriverID),
		Name:          "Test Driver",
		PhoneNumber:   "081234567890",
		VehicleID:     uuid.MustParse(testVehicleID),
		VehicleName:   "Test Motorcycle",
		LicenseNumber: "B1234XX",
		MaxSize:       2,
		Type:          entities.VehicleTypeMotorcycle,
	}
}

func findDriverRequestBody(t *testing.T) *bytes.Reader {
	t.Helper()
	raw, err := json.Marshal(map[string]any{
		"current_lat_long": []string{"-6.2088", "106.8456"},
		"vehicle_type":     entities.VehicleTypeMotorcycle,
	})
	require.NoError(t, err)
	return bytes.NewReader(raw)
}

func TestFindDriver_EmptyList(t *testing.T) {
	router, sign := newFindDriverRouter(t, true, nil)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/trip/dispatch/customer/find-driver",
		findDriverRequestBody(t),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+sign("user-1", "user@example.com", "customer"))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var body struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
		Data    struct {
			Drivers []struct {
				DistanceM int        `json:"distance_m"`
				Location  [2]float64 `json:"location"`
				Profile   struct {
					UserID string `json:"user_id"`
				} `json:"profile"`
			} `json:"drivers"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.True(t, body.Status)
	assert.Empty(t, body.Data.Drivers)
}

func TestFindDriver_ReturnsNearby(t *testing.T) {
	store := newGeoStore(t)
	repo := &stubDispatchRepo{
		nearbyProfiles: map[string]dto.NearbyDriverProfile{
			testDriverUserID: testNearbyDriverProfile(),
		},
	}
	router, sign, _ := newFindDriverRouterWithStoreAndAllow(t, true, store, repo)
	require.NoError(t, store.SetStandby(context.Background(), testDriverUserID, -6.2088, 106.8456))

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/trip/dispatch/customer/find-driver",
		findDriverRequestBody(t),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+sign("user-1", "user@example.com", "customer"))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var body struct {
		Status bool `json:"status"`
		Data   struct {
			Drivers []struct {
				DistanceM int        `json:"distance_m"`
				Location  [2]float64 `json:"location"`
				Profile   struct {
					UserID      string `json:"user_id"`
					Name        string `json:"name"`
					VehicleName string `json:"vehicle_name"`
				} `json:"profile"`
			} `json:"drivers"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Len(t, body.Data.Drivers, 1)
	assert.Equal(t, testDriverUserID, body.Data.Drivers[0].Profile.UserID)
	assert.Equal(t, "Test Driver", body.Data.Drivers[0].Profile.Name)
	assert.Equal(t, "Test Motorcycle", body.Data.Drivers[0].Profile.VehicleName)
	assert.Equal(t, 0, body.Data.Drivers[0].DistanceM)
}

func TestFindDriver_DeniesWhenUnauthorized(t *testing.T) {
	router, sign := newFindDriverRouter(t, false, nil)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/trip/dispatch/customer/find-driver",
		findDriverRequestBody(t),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+sign("user-1", "drv@example.com", "driver"))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestFindDriverService_EmptyList(t *testing.T) {
	store := newGeoStore(t)
	svc := service.NewDispatchService(&stubDispatchRepo{}, nil, nil, "", store)

	result, err := svc.FindDriver(context.Background(), dto.FindDriverRequest{
		CurrentLatLong: [2]string{"-6.2088", "106.8456"},
		VehicleType:    entities.VehicleTypeMotorcycle,
	})
	require.NoError(t, err)
	assert.Empty(t, result.Drivers)
}

func TestFindDriverService_ReturnsNearby(t *testing.T) {
	store := newGeoStore(t)
	require.NoError(t, store.SetStandby(context.Background(), testDriverUserID, -6.2088, 106.8456))

	repo := &stubDispatchRepo{
		nearbyProfiles: map[string]dto.NearbyDriverProfile{
			testDriverUserID: testNearbyDriverProfile(),
		},
	}
	svc := service.NewDispatchService(repo, nil, nil, "", store)
	result, err := svc.FindDriver(context.Background(), dto.FindDriverRequest{
		CurrentLatLong: [2]string{"-6.2088", "106.8456"},
		VehicleType:    entities.VehicleTypeMotorcycle,
	})
	require.NoError(t, err)
	require.Len(t, result.Drivers, 1)
	assert.Equal(t, testDriverUserID, result.Drivers[0].Profile.UserID.String())
	assert.Equal(t, "Test Driver", result.Drivers[0].Profile.Name)
	assert.Equal(t, 0, result.Drivers[0].DistanceM)
	assert.InDelta(t, -6.2088, result.Drivers[0].Location[0], 0.0001)
	assert.InDelta(t, 106.8456, result.Drivers[0].Location[1], 0.0001)
}

func TestFindDriverService_InvalidLatLong(t *testing.T) {
	store := newGeoStore(t)
	svc := service.NewDispatchService(&stubDispatchRepo{}, nil, nil, "", store)

	_, err := svc.FindDriver(context.Background(), dto.FindDriverRequest{
		CurrentLatLong: [2]string{"not-a-lat", "106.8456"},
		VehicleType:    entities.VehicleTypeMotorcycle,
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrInvalidLatLong))
}
