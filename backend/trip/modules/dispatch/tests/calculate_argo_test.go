package tests

import (
	"bytes"
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

func TestCalculateArgo_RequiresBearer(t *testing.T) {
	router, _ := newCalculateArgoRouter(t, http.NotFoundHandler(), true)

	req := httptest.NewRequest(http.MethodPost, "/api/trip/dispatch/calculate-argo", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestCalculateArgo_RejectsInvalidToken(t *testing.T) {
	router, _ := newCalculateArgoRouter(t, http.NotFoundHandler(), true)

	req := httptest.NewRequest(http.MethodPost, "/api/trip/dispatch/calculate-argo", bytes.NewBufferString(`{}`))
	req.Header.Set("Authorization", "Bearer not-a-token")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestCalculateArgo_RejectsInvalidBody(t *testing.T) {
	router, sign := newCalculateArgoRouter(t, http.NotFoundHandler(), true)
	vehicle := testVehicle(entities.VehicleTypeMotorcycle)

	req := httptest.NewRequest(http.MethodPost, "/api/trip/dispatch/calculate-argo", bytes.NewBufferString(`{
		"pickup_loc": ["1000", "106.8456"],
		"destination": ["-6.1754", "106.8272"],
		"vehicle_id": "`+vehicle.ID.String()+`"
	}`))
	req.Header.Set("Authorization", "Bearer "+sign("user-1", "user@example.com", "customer"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCalculateArgo_ReturnsQuote(t *testing.T) {
	osrm := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"code": "Ok",
			"routes": [{
				"distance": 2000,
				"duration": 180,
				"geometry": {
					"type": "LineString",
					"coordinates": [[106.8456, -6.2088], [106.84, -6.2]]
				}
			}]
		}`)
	})
	router, sign := newCalculateArgoRouter(t, osrm, true)
	vehicle := testVehicle(entities.VehicleTypeMotorcycle)

	req := httptest.NewRequest(http.MethodPost, "/api/trip/dispatch/calculate-argo", bytes.NewBufferString(`{
		"pickup_loc": ["-6.2088", "106.8456"],
		"destination": ["-6.1754", "106.8272"],
		"vehicle_id": "`+vehicle.ID.String()+`"
	}`))
	req.Header.Set("Authorization", "Bearer "+sign("user-1", "user@example.com", "customer"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var body struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
		Data    struct {
			Distance           int          `json:"distance"`
			Duration           int          `json:"duration"`
			Path               [][2]float64 `json:"path"`
			FarePerDistance    int          `json:"fare_per_distance"`
			PlatformPercentage int          `json:"platform_percentage"`
			TotalFare          int          `json:"total_fare"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.True(t, body.Status)
	assert.Equal(t, 2000, body.Data.Distance)
	assert.Equal(t, 180, body.Data.Duration)
	assert.Equal(t, 2500, body.Data.FarePerDistance)
	assert.Equal(t, 10, body.Data.PlatformPercentage)
	assert.Equal(t, 5500, body.Data.TotalFare)
	require.Equal(t, [][2]float64{{-6.2088, 106.8456}, {-6.2, 106.84}}, body.Data.Path)
}

func TestCalculateArgo_NoRouteIsBadRequest(t *testing.T) {
	osrm := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"code":"NoRoute","routes":[]}`)
	})
	router, sign := newCalculateArgoRouter(t, osrm, true)
	vehicle := testVehicle(entities.VehicleTypeMotorcycle)

	req := httptest.NewRequest(http.MethodPost, "/api/trip/dispatch/calculate-argo", bytes.NewBufferString(`{
		"pickup_loc": ["-6.2088", "106.8456"],
		"destination": ["-6.1754", "106.8272"],
		"vehicle_id": "`+vehicle.ID.String()+`"
	}`))
	req.Header.Set("Authorization", "Bearer "+sign("user-1", "user@example.com", "customer"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCalculateArgo_OSRMDownIsBadGateway(t *testing.T) {
	osrm := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	router, sign := newCalculateArgoRouter(t, osrm, true)
	vehicle := testVehicle(entities.VehicleTypeMotorcycle)

	req := httptest.NewRequest(http.MethodPost, "/api/trip/dispatch/calculate-argo", bytes.NewBufferString(`{
		"pickup_loc": ["-6.2088", "106.8456"],
		"destination": ["-6.1754", "106.8272"],
		"vehicle_id": "`+vehicle.ID.String()+`"
	}`))
	req.Header.Set("Authorization", "Bearer "+sign("user-1", "user@example.com", "customer"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadGateway, rec.Code)
}

func TestCalculateArgo_DeniesWhenUnauthorized(t *testing.T) {
	router, sign := newCalculateArgoRouter(t, http.NotFoundHandler(), false)
	vehicle := testVehicle(entities.VehicleTypeMotorcycle)

	req := httptest.NewRequest(http.MethodPost, "/api/trip/dispatch/calculate-argo", bytes.NewBufferString(`{
		"pickup_loc": ["-6.2088", "106.8456"],
		"destination": ["-6.1754", "106.8272"],
		"vehicle_id": "`+vehicle.ID.String()+`"
	}`))
	req.Header.Set("Authorization", "Bearer "+sign("user-1", "drv@example.com", "driver"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

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

	vehicle := testVehicle(entities.VehicleTypeMotorcycle)
	repo := &stubDispatchRepo{vehicle: vehicle}
	svc := service.NewDispatchService(repo, nil, osrm.Client(), osrm.URL, nil)
	ctx := context.WithValue(context.Background(), "customer", testCustomer())

	result, err := svc.CalculateArgo(ctx, dto.CalculateArgoRequest{
		PickupLoc:   [2]string{"-6.2088", "106.8456"},
		Destination: [2]string{"-6.1754", "106.8272"},
		VehicleId:   vehicle.ID,
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

	vehicle := testVehicle(entities.VehicleTypeCar)
	repo := &stubDispatchRepo{vehicle: vehicle}
	svc := service.NewDispatchService(repo, nil, osrm.Client(), osrm.URL, nil)
	ctx := context.WithValue(context.Background(), "customer", testCustomer())

	result, err := svc.CalculateArgo(ctx, dto.CalculateArgoRequest{
		PickupLoc:   [2]string{"-6.2088", "106.8456"},
		Destination: [2]string{"-6.1754", "106.8272"},
		VehicleId:   vehicle.ID,
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

	vehicle := testVehicle(entities.VehicleTypeMotorcycle)
	repo := &stubDispatchRepo{vehicle: vehicle}
	svc := service.NewDispatchService(repo, nil, osrm.Client(), osrm.URL, nil)
	ctx := context.WithValue(context.Background(), "customer", testCustomer())

	_, err := svc.CalculateArgo(ctx, dto.CalculateArgoRequest{
		PickupLoc:   [2]string{"-6.2088", "106.8456"},
		Destination: [2]string{"-6.1754", "106.8272"},
		VehicleId:   vehicle.ID,
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrNoRoute))
}

func TestCalculateArgo_OSRMUnavailable(t *testing.T) {
	osrm := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(osrm.Close)

	vehicle := testVehicle(entities.VehicleTypeMotorcycle)
	repo := &stubDispatchRepo{vehicle: vehicle}
	svc := service.NewDispatchService(repo, nil, osrm.Client(), osrm.URL, nil)
	ctx := context.WithValue(context.Background(), "customer", testCustomer())

	_, err := svc.CalculateArgo(ctx, dto.CalculateArgoRequest{
		PickupLoc:   [2]string{"-6.2088", "106.8456"},
		Destination: [2]string{"-6.1754", "106.8272"},
		VehicleId:   vehicle.ID,
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrOSRMUnavailable))
}

func TestCalculateArgo_InvalidJSONFromOSRM(t *testing.T) {
	osrm := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `not-json`)
	}))
	t.Cleanup(osrm.Close)

	vehicle := testVehicle(entities.VehicleTypeMotorcycle)
	repo := &stubDispatchRepo{vehicle: vehicle}
	svc := service.NewDispatchService(repo, nil, osrm.Client(), osrm.URL, nil)
	ctx := context.WithValue(context.Background(), "customer", testCustomer())

	_, err := svc.CalculateArgo(ctx, dto.CalculateArgoRequest{
		PickupLoc:   [2]string{"-6.2088", "106.8456"},
		Destination: [2]string{"-6.1754", "106.8272"},
		VehicleId:   vehicle.ID,
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
	})
	require.NoError(t, err)
	assert.Contains(t, string(body), `"path":[[-6.2,106.8]]`)
}
