package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/saved_address/dto"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type envelope struct {
	Status  bool            `json:"status"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func TestSavedAddressCRUD_CreateListGetUpdateDelete(t *testing.T) {
	tr := newSavedAddressRouter(t, true, testCustomer())
	customer := testCustomer()
	header := authHeader(tr.sign, customer)

	createBody := map[string]any{
		"name":              "Home",
		"lat_long":          []string{"-6.2088", "106.8456"},
		"is_default_pickup": true,
	}
	raw, err := json.Marshal(createBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/trip/saved-addresses", bytes.NewReader(raw))
	req.Header.Set("Authorization", header)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	tr.router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var createEnv envelope
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &createEnv))
	assert.True(t, createEnv.Status)
	assert.Equal(t, dto.MESSAGE_SUCCESS_CREATE_SAVED_ADDRESS, createEnv.Message)

	var created dto.SavedAddressResponse
	require.NoError(t, json.Unmarshal(createEnv.Data, &created))
	assert.Equal(t, "Home", created.Name)
	assert.True(t, created.IsDefaultPickup)
	assert.NotEqual(t, uuid.Nil, created.ID)

	req = httptest.NewRequest(http.MethodGet, "/api/trip/saved-addresses", nil)
	req.Header.Set("Authorization", header)
	w = httptest.NewRecorder()
	tr.router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var listEnv envelope
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listEnv))
	var listed []dto.SavedAddressResponse
	require.NoError(t, json.Unmarshal(listEnv.Data, &listed))
	require.Len(t, listed, 1)
	assert.Equal(t, created.ID, listed[0].ID)

	req = httptest.NewRequest(http.MethodGet, "/api/trip/saved-addresses/"+created.ID.String(), nil)
	req.Header.Set("Authorization", header)
	w = httptest.NewRecorder()
	tr.router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	updateBody := map[string]any{
		"name":              "Office",
		"lat_long":          []string{"-6.1754", "106.8272"},
		"is_default_pickup": false,
	}
	raw, err = json.Marshal(updateBody)
	require.NoError(t, err)
	req = httptest.NewRequest(http.MethodPut, "/api/trip/saved-addresses/"+created.ID.String(), bytes.NewReader(raw))
	req.Header.Set("Authorization", header)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	tr.router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var updateEnv envelope
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &updateEnv))
	var updated dto.SavedAddressResponse
	require.NoError(t, json.Unmarshal(updateEnv.Data, &updated))
	assert.Equal(t, "Office", updated.Name)
	assert.False(t, updated.IsDefaultPickup)
	assert.Equal(t, [2]string{"-6.1754", "106.8272"}, updated.LatLong)

	req = httptest.NewRequest(http.MethodDelete, "/api/trip/saved-addresses/"+created.ID.String(), nil)
	req.Header.Set("Authorization", header)
	w = httptest.NewRecorder()
	tr.router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	req = httptest.NewRequest(http.MethodGet, "/api/trip/saved-addresses/"+created.ID.String(), nil)
	req.Header.Set("Authorization", header)
	w = httptest.NewRecorder()
	tr.router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestSavedAddress_CrossCustomerIsolation(t *testing.T) {
	customer := testCustomer()
	tr := newSavedAddressRouter(t, true, customer)
	other := seedAddress(tr.repo, otherCustomer().ID, "Other Home", true)

	req := httptest.NewRequest(http.MethodGet, "/api/trip/saved-addresses/"+other.ID.String(), nil)
	req.Header.Set("Authorization", authHeader(tr.sign, customer))
	w := httptest.NewRecorder()
	tr.router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)

	req = httptest.NewRequest(http.MethodGet, "/api/trip/saved-addresses", nil)
	req.Header.Set("Authorization", authHeader(tr.sign, customer))
	w = httptest.NewRecorder()
	tr.router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var listEnv envelope
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listEnv))
	var listed []dto.SavedAddressResponse
	require.NoError(t, json.Unmarshal(listEnv.Data, &listed))
	assert.Empty(t, listed)
}

func TestSavedAddress_ClearDefaultPickupOnCreate(t *testing.T) {
	customer := testCustomer()
	tr := newSavedAddressRouter(t, true, customer)
	existing := seedAddress(tr.repo, customer.ID, "Old Default", true)

	body := map[string]any{
		"name":              "New Default",
		"lat_long":          []string{"-6.1754", "106.8272"},
		"is_default_pickup": true,
	}
	raw, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/trip/saved-addresses", bytes.NewReader(raw))
	req.Header.Set("Authorization", authHeader(tr.sign, customer))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	tr.router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	tr.repo.mu.Lock()
	defer tr.repo.mu.Unlock()
	assert.False(t, tr.repo.addresses[existing.ID].IsDefaultPickup)

	var createEnv envelope
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &createEnv))
	var created dto.SavedAddressResponse
	require.NoError(t, json.Unmarshal(createEnv.Data, &created))
	assert.True(t, created.IsDefaultPickup)
	assert.True(t, tr.repo.addresses[created.ID].IsDefaultPickup)
}

func TestSavedAddress_ClearDefaultPickupOnUpdate(t *testing.T) {
	customer := testCustomer()
	tr := newSavedAddressRouter(t, true, customer)
	defaultAddr := seedAddress(tr.repo, customer.ID, "Default", true)
	other := seedAddress(tr.repo, customer.ID, "Secondary", false)

	body := map[string]any{
		"name":              "Secondary Now Default",
		"lat_long":          []string{"-6.1754", "106.8272"},
		"is_default_pickup": true,
	}
	raw, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPut, "/api/trip/saved-addresses/"+other.ID.String(), bytes.NewReader(raw))
	req.Header.Set("Authorization", authHeader(tr.sign, customer))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	tr.router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	tr.repo.mu.Lock()
	defer tr.repo.mu.Unlock()
	assert.False(t, tr.repo.addresses[defaultAddr.ID].IsDefaultPickup)
	assert.True(t, tr.repo.addresses[other.ID].IsDefaultPickup)
}

func TestSavedAddress_ForbiddenWhenCasbinDenies(t *testing.T) {
	tr := newSavedAddressRouter(t, false, testCustomer())
	req := httptest.NewRequest(http.MethodGet, "/api/trip/saved-addresses", nil)
	req.Header.Set("Authorization", authHeader(tr.sign, testCustomer()))
	w := httptest.NewRecorder()
	tr.router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}
