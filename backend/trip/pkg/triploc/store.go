package triploc

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const locationTTL = 24 * time.Hour

type Coords struct {
	Lat float64
	Lng float64
}

type Store struct {
	rdb *redis.Client
}

func NewStore(rdb *redis.Client) *Store {
	return &Store{rdb: rdb}
}

func driverKey(transactionID string) string {
	return fmt.Sprintf("trip:%s:driver", transactionID)
}

func customerKey(transactionID string) string {
	return fmt.Sprintf("trip:%s:customer", transactionID)
}

func (s *Store) SetDriver(ctx context.Context, transactionID string, lat, lng float64) error {
	return s.set(ctx, driverKey(transactionID), lat, lng)
}

func (s *Store) SetCustomer(ctx context.Context, transactionID string, lat, lng float64) error {
	return s.set(ctx, customerKey(transactionID), lat, lng)
}

func (s *Store) GetDriver(ctx context.Context, transactionID string) (Coords, bool, error) {
	return s.get(ctx, driverKey(transactionID))
}

func (s *Store) GetCustomer(ctx context.Context, transactionID string) (Coords, bool, error) {
	return s.get(ctx, customerKey(transactionID))
}

func (s *Store) ClearTrip(ctx context.Context, transactionID string) error {
	return s.rdb.Del(ctx, driverKey(transactionID), customerKey(transactionID)).Err()
}

func (s *Store) set(ctx context.Context, key string, lat, lng float64) error {
	pipe := s.rdb.Pipeline()
	pipe.HSet(ctx, key, map[string]any{
		"lat": strconv.FormatFloat(lat, 'f', -1, 64),
		"lng": strconv.FormatFloat(lng, 'f', -1, 64),
	})
	pipe.Expire(ctx, key, locationTTL)
	_, err := pipe.Exec(ctx)
	return err
}

func (s *Store) get(ctx context.Context, key string) (Coords, bool, error) {
	values, err := s.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return Coords{}, false, err
	}
	if len(values) == 0 {
		return Coords{}, false, nil
	}
	lat, err := strconv.ParseFloat(values["lat"], 64)
	if err != nil {
		return Coords{}, false, err
	}
	lng, err := strconv.ParseFloat(values["lng"], 64)
	if err != nil {
		return Coords{}, false, err
	}
	return Coords{Lat: lat, Lng: lng}, true, nil
}
