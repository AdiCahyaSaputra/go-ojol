package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/database/entities"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatch/dto"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatch/repository"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/drivergeo"
	"gorm.io/gorm"
)

const (
	defaultOSRMBaseURL = "https://router.project-osrm.org"
	osrmUserAgent      = "go-ojol/trip"
	osrmTimeout        = 10 * time.Second

	farePerKmMotorcycle = 2500 // IDR
	farePerKmCar        = 4500 // IDR
	platformPercentage  = 10
)

var (
	ErrNoRoute         = errors.New("no route found")
	ErrOSRMUnavailable = errors.New("routing service unavailable")
	ErrInvalidLatLong  = errors.New("invalid lat long")
	ErrUnknownVehicle  = errors.New("unknown vehicle type")
	ErrLocationStore   = errors.New("location store unavailable")
)

type DispatchService interface {
	CalculateArgo(ctx context.Context, req dto.CalculateArgoRequest) (dto.CalculateArgoResponse, error)
	FindDriver(ctx context.Context, req dto.FindDriverRequest) (dto.FindDriverResponse, error)
}

type dispatchService struct {
	dispatchRepository repository.DispatchRepository
	db                 *gorm.DB
	httpClient         *http.Client
	osrmBaseURL        string
	locations          *drivergeo.Store
}

func NewDispatchService(
	dispatchRepo repository.DispatchRepository,
	db *gorm.DB,
	httpClient *http.Client,
	osrmBaseURL string,
	locations *drivergeo.Store,
) DispatchService {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: osrmTimeout}
	}
	if osrmBaseURL == "" {
		osrmBaseURL = defaultOSRMBaseURL
	}

	return &dispatchService{
		dispatchRepository: dispatchRepo,
		db:                 db,
		httpClient:         httpClient,
		osrmBaseURL:        strings.TrimRight(osrmBaseURL, "/"),
		locations:          locations,
	}
}

func (s *dispatchService) CalculateArgo(ctx context.Context, req dto.CalculateArgoRequest) (dto.CalculateArgoResponse, error) {
	vehicle, err := s.dispatchRepository.VehicleById(req.VehicleId)

	if err != nil || vehicle == nil {
		return dto.CalculateArgoResponse{}, err
	}

	pickupLat, pickupLng, err := parseLatLong(req.PickupLoc)
	if err != nil {
		return dto.CalculateArgoResponse{}, err
	}
	destLat, destLng, err := parseLatLong(req.Destination)
	if err != nil {
		return dto.CalculateArgoResponse{}, err
	}

	farePerDistance, err := farePerDistanceFor(vehicle.Type)
	if err != nil {
		return dto.CalculateArgoResponse{}, err
	}

	route, err := s.getOSRMRoute(ctx, pickupLat, pickupLng, destLat, destLng)
	if err != nil {
		return dto.CalculateArgoResponse{}, err
	}

	customer, ok := ctx.Value("customer").(entities.Customer)

	if !ok {
		return dto.CalculateArgoResponse{}, errors.New(dto.MESSAGE_CUSTOMER_NOT_FOUND_CTX)
	}

	distance := int(math.Round(route.Distance))
	duration := int(math.Round(route.Duration))
	base := int(math.Round(float64(distance) / 1000 * float64(farePerDistance)))
	total := base + base*platformPercentage/100

	err = s.dispatchRepository.PendingArgoTransaction(dto.PendingArgoTransaction{
		CustomerID:         customer.ID,
		VehicleID:          vehicle.ID,
		PickupLatLong:      [2]string{formatCoord(pickupLat), formatCoord(pickupLng)},
		LastLatLong:        [2]string{formatCoord(pickupLat), formatCoord(pickupLng)},
		DestinationLatLong: [2]string{formatCoord(destLat), formatCoord(destLng)},
		Distance:           distance,
		FarePerDistance:    farePerDistance,
		PlatformPercentage: platformPercentage,
		TotalFare:          total,
	})

	if err != nil {
		return dto.CalculateArgoResponse{}, err
	}

	return dto.CalculateArgoResponse{
		Distance:           distance,
		Duration:           duration,
		Path:               geoJSONToLatLngPath(route.Geometry.Coordinates),
		FarePerDistance:    farePerDistance,
		PlatformPercentage: platformPercentage,
		TotalFare:          total,
	}, nil
}

func (s *dispatchService) FindDriver(ctx context.Context, req dto.FindDriverRequest) (dto.FindDriverResponse, error) {
	if s.locations == nil {
		return dto.FindDriverResponse{}, ErrLocationStore
	}

	lat, lng, err := parseLatLong(req.CurrentLocation)
	if err != nil {
		return dto.FindDriverResponse{}, err
	}

	nearby, err := s.locations.Nearby(ctx, lat, lng, drivergeo.DefaultRadiusKm, drivergeo.DefaultCount)
	if err != nil {
		return dto.FindDriverResponse{}, err
	}

	drivers := make([]dto.NearbyDriver, 0, len(nearby))
	for _, driver := range nearby {
		drivers = append(drivers, dto.NearbyDriver{
			UserID:    driver.UserID,
			DistanceM: driver.DistanceM,
			Location:  [2]float64{driver.Lat, driver.Lng},
		})
	}

	return dto.FindDriverResponse{Drivers: drivers}, nil
}

type osrmRouteResponse struct {
	Code   string      `json:"code"`
	Routes []osrmRoute `json:"routes"`
}

type osrmRoute struct {
	Distance float64      `json:"distance"`
	Duration float64      `json:"duration"`
	Geometry osrmGeometry `json:"geometry"`
}

type osrmGeometry struct {
	Coordinates [][]float64 `json:"coordinates"`
}

func (s *dispatchService) getOSRMRoute(ctx context.Context, pickupLat, pickupLng, destLat, destLng float64) (osrmRoute, error) {
	url := fmt.Sprintf(
		"%s/route/v1/driving/%s,%s;%s,%s?geometries=geojson&overview=full",
		s.osrmBaseURL,
		formatCoord(pickupLng), // Notice that we pass longitude first
		formatCoord(pickupLat),
		formatCoord(destLng),
		formatCoord(destLat),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return osrmRoute{}, fmt.Errorf("%w: %v", ErrOSRMUnavailable, err)
	}
	req.Header.Set("User-Agent", osrmUserAgent)

	res, err := s.httpClient.Do(req)
	if err != nil {
		return osrmRoute{}, fmt.Errorf("%w: %v", ErrOSRMUnavailable, err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(io.LimitReader(res.Body, 8<<20))
	if err != nil {
		return osrmRoute{}, fmt.Errorf("%w: %v", ErrOSRMUnavailable, err)
	}

	if res.StatusCode >= 500 {
		return osrmRoute{}, fmt.Errorf("%w: status %d", ErrOSRMUnavailable, res.StatusCode)
	}
	if res.StatusCode >= 400 {
		return osrmRoute{}, ErrNoRoute
	}

	var parsed osrmRouteResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return osrmRoute{}, fmt.Errorf("%w: %v", ErrOSRMUnavailable, err)
	}

	if parsed.Code != "Ok" || len(parsed.Routes) == 0 {
		return osrmRoute{}, ErrNoRoute
	}

	return parsed.Routes[0], nil
}

func parseLatLong(pair [2]string) (lat float64, lng float64, err error) {
	lat, errLat := strconv.ParseFloat(pair[0], 64)
	lng, errLng := strconv.ParseFloat(pair[1], 64)
	if errLat != nil || errLng != nil {
		return 0, 0, ErrInvalidLatLong
	}
	return lat, lng, nil
}

func formatCoord(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func farePerDistanceFor(vehicleType entities.VehicleType) (int, error) {
	switch vehicleType {
	case entities.VehicleTypeMotorcycle:
		return farePerKmMotorcycle, nil
	case entities.VehicleTypeCar:
		return farePerKmCar, nil
	default:
		return 0, ErrUnknownVehicle
	}
}

// The geoJSON response is in reverse [long, lat] so we need to re-reverse it
func geoJSONToLatLngPath(coordinates [][]float64) [][2]float64 {
	path := make([][2]float64, 0, len(coordinates))
	for _, coord := range coordinates {
		if len(coord) < 2 {
			continue
		}
		path = append(path, [2]float64{coord[1], coord[0]})
	}
	return path
}
