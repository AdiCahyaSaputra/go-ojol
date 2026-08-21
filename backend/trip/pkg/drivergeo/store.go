package drivergeo

import (
	"context"
	"math"

	"github.com/redis/go-redis/v9"
)

const (
	KeyStandby      = "drivers:standby"
	DefaultRadiusKm = 3.0
	DefaultCount    = 10
)

type NearbyDriver struct {
	UserID    string
	DistanceM int
	Lat       float64
	Lng       float64
}

type Store struct {
	rdb *redis.Client
}

func NewStore(rdb *redis.Client) *Store {
	return &Store{rdb: rdb}
}

func (s *Store) SetStandby(ctx context.Context, userId string, lat, lng float64) error {
	return s.rdb.GeoAdd(ctx, KeyStandby, &redis.GeoLocation{
		Name:      userId,
		Longitude: lng,
		Latitude:  lat,
	}).Err()
}

func (s *Store) RemoveStandby(ctx context.Context, userID string) error {
	return s.rdb.ZRem(ctx, KeyStandby, userID).Err()
}

func (s *Store) Nearby(ctx context.Context, lat, lng float64, radiusKm float64, count int) ([]NearbyDriver, error) {
	if radiusKm <= 0 {
		radiusKm = DefaultRadiusKm
	}
	if count <= 0 {
		count = DefaultCount
	}

	q := &redis.GeoSearchLocationQuery{
		GeoSearchQuery: redis.GeoSearchQuery{
			Longitude:  lng,
			Latitude:   lat,
			Count:      count,
			Radius:     radiusKm,
			RadiusUnit: "km",
			Sort:       "ASC",
		},
		WithDist:  true,
		WithCoord: true,
	}

	// go-redis v9.18–v9.22 GeoSearchLocation() duplicates query args
	// (https://github.com/redis/go-redis/pull/3955). Build once via the ctor.
	cmd := redis.NewGeoSearchLocationCmd(ctx, q, "geosearch", KeyStandby)
	if err := s.rdb.Process(ctx, cmd); err != nil {
		return nil, err
	}
	locations, err := cmd.Result()
	if err != nil {
		return nil, err
	}

	drivers := make([]NearbyDriver, 0, len(locations))
	for _, loc := range locations {
		drivers = append(drivers, NearbyDriver{
			UserID:    loc.Name,
			DistanceM: int(math.Round(loc.Dist * 1000)),
			Lat:       loc.Latitude,
			Lng:       loc.Longitude,
		})
	}
	return drivers, nil
}
