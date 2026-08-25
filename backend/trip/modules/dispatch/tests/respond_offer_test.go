package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/database/entities"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatch/dto"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatch/service"
	wsdto "github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatchws/dto"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func driverOfferCtx() context.Context {
	return context.WithValue(context.Background(), "driver", testDriver())
}

func seedOffer(t *testing.T, svc service.DispatchService, repo *stubDispatchRepo, store interface {
	SetStandby(context.Context, string, float64, float64) error
}, notifier *stubNotifier) uuid.UUID {
	t.Helper()
	require.NoError(t, store.SetStandby(context.Background(), testDriverUserID, -6.2088, 106.8456))
	repo.nearbyProfiles = map[string]dto.NearbyDriverProfile{
		testDriverUserID: testNearbyDriverProfile(),
	}

	result, err := svc.FindDriver(findDriverCtx(), dto.FindDriverRequest{
		PickupLatLong:      [2]string{"-6.2088", "106.8456"},
		DestinationLatLong: [2]string{"-6.1754", "106.8272"},
		VehicleType:        entities.VehicleTypeMotorcycle,
		MaxSize:            1,
	})
	require.NoError(t, err)
	require.NotNil(t, result.TransactionID)
	require.NotEmpty(t, notifier.snapshot())
	return *result.TransactionID
}

func TestRespondOffer_AcceptClaimsOnce(t *testing.T) {
	store := newGeoStore(t)
	repo := &stubDispatchRepo{}
	notifier := &stubNotifier{}
	osrm := httptest.NewServer(osrmOKHandler())
	t.Cleanup(osrm.Close)

	svc := service.NewDispatchService(repo, nil, osrm.Client(), osrm.URL, store, notifier)
	txID := seedOffer(t, svc, repo, store, notifier)

	first, err := svc.RespondOffer(driverOfferCtx(), txID, dto.RespondOfferRequest{Action: dto.OfferActionAccept})
	require.NoError(t, err)
	assert.Equal(t, entities.TransactionStatusAcceptedOffer, first.Status)

	_, err = svc.RespondOffer(driverOfferCtx(), txID, dto.RespondOfferRequest{Action: dto.OfferActionAccept})
	require.Error(t, err)
	assert.ErrorIs(t, err, service.ErrOfferUnavailable)
	assert.Equal(t, 1, repo.claimCalls)

	nearby, err := store.Nearby(context.Background(), -6.2088, 106.8456, 3, 10)
	require.NoError(t, err)
	assert.Empty(t, nearby)
}

func TestRespondOffer_RejectLeavesOfferOpenUntilLast(t *testing.T) {
	store := newGeoStore(t)
	secondDriverUserID := "77777777-7777-7777-7777-777777777777"
	repo := &stubDispatchRepo{
		nearbyProfiles: map[string]dto.NearbyDriverProfile{
			testDriverUserID: testNearbyDriverProfile(),
			secondDriverUserID: {
				UserID:        uuid.MustParse(secondDriverUserID),
				DriverID:      uuid.MustParse("88888888-8888-8888-8888-888888888888"),
				Name:          "Second Driver",
				PhoneNumber:   "081111111111",
				VehicleID:     uuid.MustParse("99999999-9999-9999-9999-999999999999"),
				VehicleName:   "Bike 2",
				LicenseNumber: "B9999YY",
				MaxSize:       2,
				Type:          entities.VehicleTypeMotorcycle,
			},
		},
	}
	notifier := &stubNotifier{}
	osrm := httptest.NewServer(osrmOKHandler())
	t.Cleanup(osrm.Close)

	require.NoError(t, store.SetStandby(context.Background(), testDriverUserID, -6.2088, 106.8456))
	require.NoError(t, store.SetStandby(context.Background(), secondDriverUserID, -6.2090, 106.8458))

	svc := service.NewDispatchService(repo, nil, osrm.Client(), osrm.URL, store, notifier)
	result, err := svc.FindDriver(findDriverCtx(), dto.FindDriverRequest{
		PickupLatLong:      [2]string{"-6.2088", "106.8456"},
		DestinationLatLong: [2]string{"-6.1754", "106.8272"},
		VehicleType:        entities.VehicleTypeMotorcycle,
		MaxSize:            1,
	})
	require.NoError(t, err)
	require.NotNil(t, result.TransactionID)
	require.Len(t, result.Drivers, 2)

	txID := *result.TransactionID
	resp, err := svc.RespondOffer(driverOfferCtx(), txID, dto.RespondOfferRequest{Action: dto.OfferActionReject})
	require.NoError(t, err)
	assert.Equal(t, entities.TransactionStatusOffered, resp.Status)

	secondDriver := entities.Driver{
		ID:        uuid.MustParse("88888888-8888-8888-8888-888888888888"),
		UserID:    uuid.MustParse(secondDriverUserID),
		VehicleID: uuid.MustParse("99999999-9999-9999-9999-999999999999"),
	}
	ctx := context.WithValue(context.Background(), "driver", secondDriver)
	resp, err = svc.RespondOffer(ctx, txID, dto.RespondOfferRequest{Action: dto.OfferActionReject})
	require.NoError(t, err)
	assert.Equal(t, entities.TransactionStatusRejectedOffer, resp.Status)
}

func TestRespondOffer_HTTPAccept(t *testing.T) {
	store := newGeoStore(t)
	repo := &stubDispatchRepo{}
	notifier := &stubNotifier{}
	router, sign, svc := newRespondOfferRouter(t, true, store, repo, notifier)
	txID := seedOffer(t, svc, repo, store, notifier)

	body, err := json.Marshal(map[string]string{"action": "accept"})
	require.NoError(t, err)
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/trip/dispatch/driver/offers/"+txID.String()+"/respond",
		bytes.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+sign(testDriverUserID, "drv@example.com", "driver"))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var res struct {
		Status bool `json:"status"`
		Data   struct {
			TransactionID string `json:"transaction_id"`
			Status        string `json:"status"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	assert.True(t, res.Status)
	assert.Equal(t, txID.String(), res.Data.TransactionID)
	assert.Equal(t, string(entities.TransactionStatusAcceptedOffer), res.Data.Status)
}

func TestRespondOffer_HTTPConflictOnSecondAccept(t *testing.T) {
	store := newGeoStore(t)
	repo := &stubDispatchRepo{}
	notifier := &stubNotifier{}
	router, sign, svc := newRespondOfferRouter(t, true, store, repo, notifier)
	txID := seedOffer(t, svc, repo, store, notifier)

	_, err := svc.RespondOffer(driverOfferCtx(), txID, dto.RespondOfferRequest{Action: dto.OfferActionAccept})
	require.NoError(t, err)

	body, err := json.Marshal(map[string]string{"action": "accept"})
	require.NoError(t, err)
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/trip/dispatch/driver/offers/"+txID.String()+"/respond",
		bytes.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+sign(testDriverUserID, "drv@example.com", "driver"))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)
}

func TestRespondOffer_AcceptNotifiesOthersOfferTaken(t *testing.T) {
	store := newGeoStore(t)
	secondDriverUserID := "77777777-7777-7777-7777-777777777777"
	repo := &stubDispatchRepo{
		nearbyProfiles: map[string]dto.NearbyDriverProfile{
			testDriverUserID: testNearbyDriverProfile(),
			secondDriverUserID: {
				UserID:        uuid.MustParse(secondDriverUserID),
				DriverID:      uuid.MustParse("88888888-8888-8888-8888-888888888888"),
				Name:          "Second Driver",
				PhoneNumber:   "081111111111",
				VehicleID:     uuid.MustParse("99999999-9999-9999-9999-999999999999"),
				VehicleName:   "Bike 2",
				LicenseNumber: "B9999YY",
				MaxSize:       2,
				Type:          entities.VehicleTypeMotorcycle,
			},
		},
	}
	notifier := &stubNotifier{}
	osrm := httptest.NewServer(osrmOKHandler())
	t.Cleanup(osrm.Close)

	require.NoError(t, store.SetStandby(context.Background(), testDriverUserID, -6.2088, 106.8456))
	require.NoError(t, store.SetStandby(context.Background(), secondDriverUserID, -6.2090, 106.8458))

	svc := service.NewDispatchService(repo, nil, osrm.Client(), osrm.URL, store, notifier)
	result, err := svc.FindDriver(findDriverCtx(), dto.FindDriverRequest{
		PickupLatLong:      [2]string{"-6.2088", "106.8456"},
		DestinationLatLong: [2]string{"-6.1754", "106.8272"},
		VehicleType:        entities.VehicleTypeMotorcycle,
		MaxSize:            1,
	})
	require.NoError(t, err)
	txID := *result.TransactionID

	_, err = svc.RespondOffer(driverOfferCtx(), txID, dto.RespondOfferRequest{Action: dto.OfferActionAccept})
	require.NoError(t, err)

	var taken []notifiedMessage
	for _, msg := range notifier.snapshot() {
		if msg.Msg.Type == wsdto.TypeOfferTaken {
			taken = append(taken, msg)
		}
	}
	require.Len(t, taken, 1)
	assert.Equal(t, secondDriverUserID, taken[0].UserID)
}
