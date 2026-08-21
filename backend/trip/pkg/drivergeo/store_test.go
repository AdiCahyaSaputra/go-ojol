package drivergeo

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore_SetStandbyNearbyRemove(t *testing.T) {
	mr := miniredis.RunT(t)
	store := NewStore(redis.NewClient(&redis.Options{Addr: mr.Addr()}))
	ctx := context.Background()

	require.NoError(t, store.SetStandby(ctx, "drv-1", -6.2088, 106.8456))
	require.NoError(t, store.SetStandby(ctx, "drv-far", -6.1754, 106.8272))

	nearby, err := store.Nearby(ctx, -6.2088, 106.8456, 0.5, 10)
	require.NoError(t, err)
	require.Len(t, nearby, 1)
	assert.Equal(t, "drv-1", nearby[0].UserID)
	assert.Equal(t, 0, nearby[0].DistanceM)
	assert.InDelta(t, -6.2088, nearby[0].Lat, 0.0001)
	assert.InDelta(t, 106.8456, nearby[0].Lng, 0.0001)

	require.NoError(t, store.Remove(ctx, "drv-1"))

	nearby, err = store.Nearby(ctx, -6.2088, 106.8456, 0.5, 10)
	require.NoError(t, err)
	assert.Empty(t, nearby)
}

func TestStore_NearbyEmpty(t *testing.T) {
	mr := miniredis.RunT(t)
	store := NewStore(redis.NewClient(&redis.Options{Addr: mr.Addr()}))

	nearby, err := store.Nearby(context.Background(), -6.2088, 106.8456, DefaultRadiusKm, DefaultCount)
	require.NoError(t, err)
	assert.Empty(t, nearby)
}
