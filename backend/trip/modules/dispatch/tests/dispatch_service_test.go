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

	svc := service.NewDispatchService(nil, nil, osrm.Client(), osrm.URL)
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

	svc := service.NewDispatchService(nil, nil, osrm.Client(), osrm.URL)
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

	svc := service.NewDispatchService(nil, nil, osrm.Client(), osrm.URL)
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

	svc := service.NewDispatchService(nil, nil, osrm.Client(), osrm.URL)
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

	svc := service.NewDispatchService(nil, nil, osrm.Client(), osrm.URL)
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
