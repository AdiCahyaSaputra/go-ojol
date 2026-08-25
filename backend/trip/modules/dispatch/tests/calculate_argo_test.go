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

const validCalculateArgoQuery = "/api/trip/dispatch/customer/calculate-argo?pickup_loc=-6.2088&pickup_loc=106.8456&destination=-6.1754&destination=106.8272"

func TestCalculateArgo_RequiresBearer(t *testing.T) {
	router, _ := newCalculateArgoRouter(t, http.NotFoundHandler(), true)

	req := httptest.NewRequest(http.MethodGet, validCalculateArgoQuery, nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestCalculateArgo_RejectsInvalidToken(t *testing.T) {
	router, _ := newCalculateArgoRouter(t, http.NotFoundHandler(), true)

	req := httptest.NewRequest(http.MethodGet, validCalculateArgoQuery, nil)
	req.Header.Set("Authorization", "Bearer not-a-token")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestCalculateArgo_RejectsInvalidBody(t *testing.T) {
	router, sign := newCalculateArgoRouter(t, http.NotFoundHandler(), true)

	req := httptest.NewRequest(http.MethodGet, "/api/trip/dispatch/customer/calculate-argo?pickup_loc=1000&pickup_loc=106.8456&destination=-6.1754&destination=106.8272", nil)
	req.Header.Set("Authorization", "Bearer "+sign("user-1", "user@example.com", "customer"))
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

	req := httptest.NewRequest(http.MethodGet, validCalculateArgoQuery, nil)
	req.Header.Set("Authorization", "Bearer "+sign("user-1", "user@example.com", "customer"))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var body struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
		Data    struct {
			Distance           int                 `json:"distance"`
			Duration           int                 `json:"duration"`
			Path               [][2]float64        `json:"path"`
			PlatformPercentage int                 `json:"platform_percentage"`
			VehicleOptions     []dto.VehicleOption `json:"vehicle_options"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.True(t, body.Status)
	assert.Equal(t, 2000, body.Data.Distance)
	assert.Equal(t, 180, body.Data.Duration)
	assert.Equal(t, 10, body.Data.PlatformPercentage)
	require.Len(t, body.Data.VehicleOptions, 2)
	assert.Equal(t, dto.VehicleOption{
		VehicleType: entities.VehicleTypeMotorcycle,
		MaxSize:     1,
		TotalFare:   5500,
	}, body.Data.VehicleOptions[0])
	assert.Equal(t, dto.VehicleOption{
		VehicleType: entities.VehicleTypeCar,
		MaxSize:     4,
		TotalFare:   10494,
	}, body.Data.VehicleOptions[1])
	require.Equal(t, [][2]float64{{-6.2088, 106.8456}, {-6.2, 106.84}}, body.Data.Path)
}

func TestCalculateArgo_NoRouteIsBadRequest(t *testing.T) {
	osrm := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"code":"NoRoute","routes":[]}`)
	})
	router, sign := newCalculateArgoRouter(t, osrm, true)

	req := httptest.NewRequest(http.MethodGet, validCalculateArgoQuery, nil)
	req.Header.Set("Authorization", "Bearer "+sign("user-1", "user@example.com", "customer"))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCalculateArgo_OSRMDownIsBadGateway(t *testing.T) {
	osrm := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	router, sign := newCalculateArgoRouter(t, osrm, true)

	req := httptest.NewRequest(http.MethodGet, validCalculateArgoQuery, nil)
	req.Header.Set("Authorization", "Bearer "+sign("user-1", "user@example.com", "customer"))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadGateway, rec.Code)
}

func TestCalculateArgo_DeniesWhenUnauthorized(t *testing.T) {
	router, sign := newCalculateArgoRouter(t, http.NotFoundHandler(), false)

	req := httptest.NewRequest(http.MethodGet, validCalculateArgoQuery, nil)
	req.Header.Set("Authorization", "Bearer "+sign("user-1", "drv@example.com", "driver"))
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

	repo := &stubDispatchRepo{
		vehicleCategories: []dto.VehicleCategory{
			{VehicleType: entities.VehicleTypeMotorcycle, MaxSize: 1},
			{VehicleType: entities.VehicleTypeCar, MaxSize: 4},
			{VehicleType: entities.VehicleTypeCar, MaxSize: 7},
		},
	}
	svc := service.NewDispatchService(repo, nil, osrm.Client(), osrm.URL, nil, nil)

	result, err := svc.CalculateArgo(context.Background(), dto.CalculateArgoRequest{
		PickupLoc:   [2]string{"-6.2088", "106.8456"},
		Destination: [2]string{"-6.1754", "106.8272"},
	})
	require.NoError(t, err)

	require.NotNil(t, gotURL)
	assert.Equal(t, "/route/v1/driving/106.8456,-6.2088;106.8272,-6.1754", gotURL.Path)
	assert.Equal(t, "geojson", gotURL.Query().Get("geometries"))
	assert.Equal(t, "full", gotURL.Query().Get("overview"))

	assert.Equal(t, 2000, result.Distance)
	assert.Equal(t, 180, result.Duration)
	assert.Equal(t, 10, result.PlatformPercentage)
	assert.Equal(t, []dto.VehicleOption{
		{
			VehicleType: entities.VehicleTypeMotorcycle,
			MaxSize:     1,
			TotalFare:   5500,
		},
		{
			VehicleType: entities.VehicleTypeCar,
			MaxSize:     4,
			TotalFare:   10494,
		},
		{
			VehicleType: entities.VehicleTypeCar,
			MaxSize:     7,
			TotalFare:   11088,
		},
	}, result.VehicleOptions)
	require.Equal(t, [][2]float64{{-6.2088, 106.8456}, {-6.2, 106.84}}, result.Path)
}

func TestCalculateArgo_SizeIncrementFare(t *testing.T) {
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

	repo := &stubDispatchRepo{
		vehicleCategories: []dto.VehicleCategory{
			{VehicleType: entities.VehicleTypeCar, MaxSize: 4},
		},
	}
	svc := service.NewDispatchService(repo, nil, osrm.Client(), osrm.URL, nil, nil)

	result, err := svc.CalculateArgo(context.Background(), dto.CalculateArgoRequest{
		PickupLoc:   [2]string{"-6.2088", "106.8456"},
		Destination: [2]string{"-6.1754", "106.8272"},
	})
	require.NoError(t, err)

	assert.Equal(t, 1000, result.Distance)
	assert.Equal(t, []dto.VehicleOption{
		{
			VehicleType: entities.VehicleTypeCar,
			MaxSize:     4,
			TotalFare:   5247,
		},
	}, result.VehicleOptions)
	assert.Empty(t, result.Path)
}

func TestCalculateArgo_NoRoute(t *testing.T) {
	osrm := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"code":"NoRoute","routes":[]}`)
	}))
	t.Cleanup(osrm.Close)

	repo := &stubDispatchRepo{}
	svc := service.NewDispatchService(repo, nil, osrm.Client(), osrm.URL, nil, nil)

	_, err := svc.CalculateArgo(context.Background(), dto.CalculateArgoRequest{
		PickupLoc:   [2]string{"-6.2088", "106.8456"},
		Destination: [2]string{"-6.1754", "106.8272"},
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrNoRoute))
}

func TestCalculateArgo_OSRMUnavailable(t *testing.T) {
	osrm := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(osrm.Close)

	repo := &stubDispatchRepo{}
	svc := service.NewDispatchService(repo, nil, osrm.Client(), osrm.URL, nil, nil)

	_, err := svc.CalculateArgo(context.Background(), dto.CalculateArgoRequest{
		PickupLoc:   [2]string{"-6.2088", "106.8456"},
		Destination: [2]string{"-6.1754", "106.8272"},
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrOSRMUnavailable))
}

func TestCalculateArgo_InvalidJSONFromOSRM(t *testing.T) {
	osrm := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `not-json`)
	}))
	t.Cleanup(osrm.Close)

	repo := &stubDispatchRepo{}
	svc := service.NewDispatchService(repo, nil, osrm.Client(), osrm.URL, nil, nil)

	_, err := svc.CalculateArgo(context.Background(), dto.CalculateArgoRequest{
		PickupLoc:   [2]string{"-6.2088", "106.8456"},
		Destination: [2]string{"-6.1754", "106.8272"},
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrOSRMUnavailable))
}

func TestCalculateArgoResponseJSONIncludesPath(t *testing.T) {
	body, err := json.Marshal(dto.CalculateArgoResponse{
		Distance:           2000,
		Duration:           180,
		Path:               [][2]float64{{-6.2, 106.8}},
		PlatformPercentage: 20,
		VehicleOptions: []dto.VehicleOption{
			{
				VehicleType: entities.VehicleTypeMotorcycle,
				MaxSize:     1,
				TotalFare:   6000,
			},
		},
	})
	require.NoError(t, err)
	assert.Contains(t, string(body), `"path":[[-6.2,106.8]]`)
	assert.Contains(t, string(body), `"vehicle_options":[`)
}
