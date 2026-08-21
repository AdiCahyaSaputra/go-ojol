package tests

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/database/entities"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatch/dto"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatch/service"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/drivergeo"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalculateArgo_MotorcycleFareAndPath(t *testing.T) {
	var gotURL *url.URL
	osrm := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotURL = r.URL
		assert.Equal(t, "go-ojol/trip", r.Header.Get("User-Agent"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"code": "Ok",
			"routes": [{
				"distance": 2000,
				"duration": 180.4,
				"geometry": {
					"type": "LineString",
					"coordinates": [[106.8456, -6.2088], [106.84, -6.2]]
				}
			}]
		}`)
	}))
	t.Cleanup(osrm.Close)

	svc := service.NewDispatchService(nil, nil, osrm.Client(), osrm.URL, nil)
	result, err := svc.CalculateArgo(context.Background(), dto.CalculateArgoRequest{
		PickupLoc:   [2]string{"-6.2088", "106.8456"},
		Destination: [2]string{"-6.1754", "106.8272"},
		VehicleType: entities.VehicleTypeMotorcycle,
	})
	require.NoError(t, err)

	require.NotNil(t, gotURL)
	assert.Equal(t, "/route/v1/driving/106.8456,-6.2088;106.8272,-6.1754", gotURL.Path)
	assert.Equal(t, "geojson", gotURL.Query().Get("geometries"))
	assert.Equal(t, "full", gotURL.Query().Get("overview"))

	assert.Equal(t, 2000, result.Distance)
	assert.Equal(t, 180, result.Duration)
	assert.Equal(t, 2500, result.FarePerDistance)
	assert.Equal(t, 10, result.PlatformPercentage)
	assert.Equal(t, 5500, result.TotalFare)
	assert.Equal(t, entities.VehicleTypeMotorcycle, result.VehicleType)
	require.Equal(t, [][2]float64{{-6.2088, 106.8456}, {-6.2, 106.84}}, result.Path)
}

func TestCalculateArgo_CarFare(t *testing.T) {
	osrm := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"code": "Ok",
			"routes": [{
				"distance": 1000,
				"duration": 60,
				"geometry": {"type": "LineString", "coordinates": []}
			}]
		}`)
	}))
	t.Cleanup(osrm.Close)

	svc := service.NewDispatchService(nil, nil, osrm.Client(), osrm.URL, nil)
	result, err := svc.CalculateArgo(context.Background(), dto.CalculateArgoRequest{
		PickupLoc:   [2]string{"-6.2088", "106.8456"},
		Destination: [2]string{"-6.1754", "106.8272"},
		VehicleType: entities.VehicleTypeCar,
	})
	require.NoError(t, err)

	assert.Equal(t, 1000, result.Distance)
	assert.Equal(t, 4500, result.FarePerDistance)
	assert.Equal(t, 4950, result.TotalFare)
	assert.Empty(t, result.Path)
}

func TestCalculateArgo_NoRoute(t *testing.T) {
	osrm := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"code":"NoRoute","routes":[]}`)
	}))
	t.Cleanup(osrm.Close)

	svc := service.NewDispatchService(nil, nil, osrm.Client(), osrm.URL, nil)
	_, err := svc.CalculateArgo(context.Background(), dto.CalculateArgoRequest{
		PickupLoc:   [2]string{"-6.2088", "106.8456"},
		Destination: [2]string{"-6.1754", "106.8272"},
		VehicleType: entities.VehicleTypeMotorcycle,
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrNoRoute))
}

func TestCalculateArgo_OSRMUnavailable(t *testing.T) {
	osrm := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(osrm.Close)

	svc := service.NewDispatchService(nil, nil, osrm.Client(), osrm.URL, nil)
	_, err := svc.CalculateArgo(context.Background(), dto.CalculateArgoRequest{
		PickupLoc:   [2]string{"-6.2088", "106.8456"},
		Destination: [2]string{"-6.1754", "106.8272"},
		VehicleType: entities.VehicleTypeMotorcycle,
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrOSRMUnavailable))
}

func TestCalculateArgo_InvalidJSONFromOSRM(t *testing.T) {
	osrm := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `not-json`)
	}))
	t.Cleanup(osrm.Close)

	svc := service.NewDispatchService(nil, nil, osrm.Client(), osrm.URL, nil)
	_, err := svc.CalculateArgo(context.Background(), dto.CalculateArgoRequest{
		PickupLoc:   [2]string{"-6.2088", "106.8456"},
		Destination: [2]string{"-6.1754", "106.8272"},
		VehicleType: entities.VehicleTypeMotorcycle,
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrOSRMUnavailable))
}

func TestCalculateArgoResponseJSONIncludesPath(t *testing.T) {
	body, err := json.Marshal(dto.CalculateArgoResponse{
		Distance:           2000,
		Duration:           180,
		Path:               [][2]float64{{-6.2, 106.8}},
		FarePerDistance:    2500,
		PlatformPercentage: 20,
		TotalFare:          6000,
		VehicleType:        entities.VehicleTypeMotorcycle,
	})
	require.NoError(t, err)
	assert.Contains(t, string(body), `"path":[[-6.2,106.8]]`)
}

func TestFindDriverService_EmptyList(t *testing.T) {
	store := newGeoStore(t)
	svc := service.NewDispatchService(nil, nil, nil, "", store)

	result, err := svc.FindDriver(context.Background(), dto.FindDriverRequest{
		CurrentLocation: [2]string{"-6.2088", "106.8456"},
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
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrInvalidLatLong))
}

func newGeoStore(t *testing.T) *drivergeo.Store {
	t.Helper()
	mr := miniredis.RunT(t)
	return drivergeo.NewStore(redis.NewClient(&redis.Options{Addr: mr.Addr()}))
}
