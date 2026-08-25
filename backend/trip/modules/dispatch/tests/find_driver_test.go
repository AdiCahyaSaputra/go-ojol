package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/database/entities"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatch/dto"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatch/service"
	wsdto "github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatchws/dto"
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
		"pickup_lat_long":      []string{"-6.2088", "106.8456"},
		"destination_lat_long": []string{"-6.1754", "106.8272"},
		"vehicle_type":         entities.VehicleTypeMotorcycle,
		"max_size":             1,
	})
	require.NoError(t, err)
	return bytes.NewReader(raw)
}

func findDriverCtx() context.Context {
	return context.WithValue(context.Background(), "customer", testCustomer())
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

func TestFindDriver_ReturnsNearbyAndCreatesOffer(t *testing.T) {
	store := newGeoStore(t)
	repo := &stubDispatchRepo{
		nearbyProfiles: map[string]dto.NearbyDriverProfile{
			testDriverUserID: testNearbyDriverProfile(),
		},
	}
	notifier := &stubNotifier{}
	router, sign, _, _ := newFindDriverRouterWithDeps(t, true, store, repo, notifier, nil)
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
			TransactionID string `json:"transaction_id"`
			ExpiresInSec  int    `json:"expires_in_sec"`
			Drivers       []struct {
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
	assert.NotEmpty(t, body.Data.TransactionID)
	assert.Equal(t, dto.OfferTTLSeconds, body.Data.ExpiresInSec)
	assert.Equal(t, testDriverUserID, body.Data.Drivers[0].Profile.UserID)
	assert.Equal(t, "Test Driver", body.Data.Drivers[0].Profile.Name)
	assert.Equal(t, "Test Motorcycle", body.Data.Drivers[0].Profile.VehicleName)
	assert.Equal(t, 0, body.Data.Drivers[0].DistanceM)

	msgs := notifier.snapshot()
	require.Len(t, msgs, 2)
	assert.Equal(t, testDriverUserID, msgs[0].UserID)
	assert.Equal(t, wsdto.TypeTripOffer, msgs[0].Msg.Type)
	require.NotNil(t, msgs[0].Msg.Offer)
	assert.Equal(t, "Test Customer", msgs[0].Msg.Offer.CustomerName)
	assert.Equal(t, 2000, msgs[0].Msg.Offer.DistanceM)
	assert.Equal(t, testCustomer().UserID.String(), msgs[1].UserID)
	assert.Equal(t, wsdto.TypeWaiting, msgs[1].Msg.Type)
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
	svc := service.NewDispatchService(&stubDispatchRepo{}, nil, nil, "", store, nil)

	result, err := svc.FindDriver(findDriverCtx(), dto.FindDriverRequest{
		PickupLatLong:      [2]string{"-6.2088", "106.8456"},
		DestinationLatLong: [2]string{"-6.1754", "106.8272"},
		VehicleType:        entities.VehicleTypeMotorcycle,
		MaxSize:            1,
	})
	require.NoError(t, err)
	assert.Empty(t, result.Drivers)
	assert.Nil(t, result.TransactionID)
}

func TestFindDriverService_ReturnsNearby(t *testing.T) {
	store := newGeoStore(t)
	require.NoError(t, store.SetStandby(context.Background(), testDriverUserID, -6.2088, 106.8456))

	osrm := httptest.NewServer(osrmOKHandler())
	t.Cleanup(osrm.Close)

	repo := &stubDispatchRepo{
		nearbyProfiles: map[string]dto.NearbyDriverProfile{
			testDriverUserID: testNearbyDriverProfile(),
		},
	}
	notifier := &stubNotifier{}
	svc := service.NewDispatchService(repo, nil, osrm.Client(), osrm.URL, store, notifier)
	result, err := svc.FindDriver(findDriverCtx(), dto.FindDriverRequest{
		PickupLatLong:      [2]string{"-6.2088", "106.8456"},
		DestinationLatLong: [2]string{"-6.1754", "106.8272"},
		VehicleType:        entities.VehicleTypeMotorcycle,
		MaxSize:            1,
	})
	require.NoError(t, err)
	require.Len(t, result.Drivers, 1)
	require.NotNil(t, result.TransactionID)
	assert.Equal(t, testDriverUserID, result.Drivers[0].Profile.UserID.String())
	assert.Equal(t, "Test Driver", result.Drivers[0].Profile.Name)
	assert.Equal(t, 0, result.Drivers[0].DistanceM)
	assert.InDelta(t, -6.2088, result.Drivers[0].Location[0], 0.0001)
	assert.InDelta(t, 106.8456, result.Drivers[0].Location[1], 0.0001)
	require.Len(t, notifier.snapshot(), 2)
}

func TestFindDriverService_InvalidLatLong(t *testing.T) {
	store := newGeoStore(t)
	svc := service.NewDispatchService(&stubDispatchRepo{}, nil, nil, "", store, nil)

	_, err := svc.FindDriver(findDriverCtx(), dto.FindDriverRequest{
		PickupLatLong:      [2]string{"not-a-lat", "106.8456"},
		DestinationLatLong: [2]string{"-6.1754", "106.8272"},
		VehicleType:        entities.VehicleTypeMotorcycle,
		MaxSize:            1,
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrInvalidLatLong))
}

func TestFindDriverService_RequiresCustomer(t *testing.T) {
	store := newGeoStore(t)
	svc := service.NewDispatchService(&stubDispatchRepo{}, nil, nil, "", store, nil)

	_, err := svc.FindDriver(context.Background(), dto.FindDriverRequest{
		PickupLatLong:      [2]string{"-6.2088", "106.8456"},
		DestinationLatLong: [2]string{"-6.1754", "106.8272"},
		VehicleType:        entities.VehicleTypeMotorcycle,
		MaxSize:            1,
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrCustomerNotInCtx))
}

func TestFindDriverService_ExpiresOffer(t *testing.T) {
	store := newGeoStore(t)
	require.NoError(t, store.SetStandby(context.Background(), testDriverUserID, -6.2088, 106.8456))

	osrm := httptest.NewServer(osrmOKHandler())
	t.Cleanup(osrm.Close)

	repo := &stubDispatchRepo{
		nearbyProfiles: map[string]dto.NearbyDriverProfile{
			testDriverUserID: testNearbyDriverProfile(),
		},
	}
	notifier := &stubNotifier{}
	svc := service.NewDispatchService(
		repo,
		nil,
		osrm.Client(),
		osrm.URL,
		store,
		notifier,
		service.WithOfferTTL(20*time.Millisecond),
	)

	result, err := svc.FindDriver(findDriverCtx(), dto.FindDriverRequest{
		PickupLatLong:      [2]string{"-6.2088", "106.8456"},
		DestinationLatLong: [2]string{"-6.1754", "106.8272"},
		VehicleType:        entities.VehicleTypeMotorcycle,
		MaxSize:            1,
	})
	require.NoError(t, err)
	require.NotNil(t, result.TransactionID)

	require.Eventually(t, func() bool {
		for _, msg := range notifier.snapshot() {
			if msg.Msg.Type == wsdto.TypeOfferExpired {
				return true
			}
		}
		return false
	}, time.Second, 10*time.Millisecond)

	txn, err := repo.TransactionByID(*result.TransactionID)
	require.NoError(t, err)
	assert.Equal(t, entities.TransactionStatusExpired, txn.Status)
}
