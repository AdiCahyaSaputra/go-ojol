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

	"github.com/google/uuid"

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
	platformPct         = 10
	maxSizeFareStepPct  = 2
)

var (
	ErrNoRoute         = errors.New("no route found")
	ErrOSRMUnavailable = errors.New("routing service unavailable")
	ErrInvalidLatLong  = errors.New("invalid lat long")
	ErrLocationStore   = errors.New("location store unavailable")
)

type DispatchService interface {
	// Customer
	CalculateArgo(ctx context.Context, req dto.CalculateArgoRequest) (dto.CalculateArgoResponse, error)
	FindDriver(ctx context.Context, req dto.FindDriverRequest) (dto.FindDriverResponse, error)

	// Driver
	SetDriverMode(ctx context.Context, req dto.SetDriverModeRequest) error
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
	pickupLat, pickupLng, err := parseLatLong(req.PickupLoc)
	if err != nil {
		return dto.CalculateArgoResponse{}, err
	}
	destLat, destLng, err := parseLatLong(req.Destination)
	if err != nil {
		return dto.CalculateArgoResponse{}, err
	}

	route, err := s.getOSRMRoute(ctx, pickupLat, pickupLng, destLat, destLng)
	if err != nil {
		return dto.CalculateArgoResponse{}, err
	}

	categories, err := s.dispatchRepository.DistinctVehicleCategories()
	if err != nil {
		return dto.CalculateArgoResponse{}, err
	}

	distance := int(math.Round(route.Distance))
	duration := int(math.Round(route.Duration))
	vehicleOptions := make([]dto.VehicleOption, 0, len(categories))

	for _, category := range categories {
		farePerDistance, ok := farePerDistanceFor(category.VehicleType)
		if !ok {
			continue
		}

		vehicleOptions = append(vehicleOptions, dto.VehicleOption{
			VehicleType: category.VehicleType,
			MaxSize:     category.MaxSize,
			TotalFare:   calculateTotalFare(distance, farePerDistance, category.MaxSize),
		})
	}

	return dto.CalculateArgoResponse{
		Distance:           distance,
		Duration:           duration,
		Path:               geoJSONToLatLngPath(route.Geometry.Coordinates),
		PlatformPercentage: platformPct,
		VehicleOptions:     vehicleOptions,
	}, nil
}

func (s *dispatchService) FindDriver(ctx context.Context, req dto.FindDriverRequest) (dto.FindDriverResponse, error) {
	if s.locations == nil {
		return dto.FindDriverResponse{}, ErrLocationStore
	}

	lat, lng, err := parseLatLong(req.CurrentLatLong)
	if err != nil {
		return dto.FindDriverResponse{}, err
	}

	nearby, err := s.locations.Nearby(ctx, lat, lng, drivergeo.DefaultRadiusKm, drivergeo.DefaultCount)
	if err != nil {
		return dto.FindDriverResponse{}, err
	}

	driverUserIds := make([]uuid.UUID, 0, len(nearby))

	for _, driver := range nearby {
		driverUserIds = append(driverUserIds, uuid.MustParse(driver.UserID))
	}

	nearbyDriverProfiles, err := s.dispatchRepository.NearbyDriverProfiles(driverUserIds, req.VehicleType)

	if err != nil {
		return dto.FindDriverResponse{}, err
	}

	drivers := make([]dto.NearbyDriver, 0, len(nearby))

	for _, driver := range nearby {
		if driverProfile, ok := nearbyDriverProfiles[driver.UserID]; ok {
			drivers = append(drivers, dto.NearbyDriver{
				DistanceM: driver.DistanceM,
				Location:  [2]float64{driver.Lat, driver.Lng},
				Profile:   driverProfile,
			})
		}
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

func (s dispatchService) SetDriverMode(ctx context.Context, req dto.SetDriverModeRequest) error {
	userId := ctx.Value("user_id").(string)

	if userId == "" {
		return errors.New(dto.MESSAGE_DRIVE_USER_ID_CONTEXT_NOT_FOUND)
	}

	lat, long, err := parseLatLong(req.CurrentLatLong)

	if err != nil {
		return err
	}

	switch req.Mode {
	case dto.DriverModeOnline:
		err = s.locations.SetStandby(ctx, userId, lat, long)

		if err != nil {
			return err
		}
	case dto.DriverModeOffline:
		err = s.locations.RemoveStandby(ctx, userId)

		if err != nil {
			return err
		}
	}

	return nil
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

func farePerDistanceFor(vehicleType entities.VehicleType) (int, bool) {
	switch vehicleType {
	case entities.VehicleTypeMotorcycle:
		return farePerKmMotorcycle, true
	case entities.VehicleTypeCar:
		return farePerKmCar, true
	default:
		return 0, false
	}
}

func calculateTotalFare(distance, farePerDistance, maxSize int) int {
	sizePct := 100
	if maxSize > 1 {
		sizePct += (maxSize - 1) * maxSizeFareStepPct
	}

	numerator := int64(distance) * int64(farePerDistance) * int64(sizePct) * int64(100+platformPct)
	denominator := int64(1000 * 100 * 100)

	return int((numerator + denominator - 1) / denominator)
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
