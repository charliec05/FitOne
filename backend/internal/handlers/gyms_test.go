package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"fitonex/backend/internal/config"
	"fitonex/backend/internal/httpx"
	"fitonex/backend/internal/models"
	"fitonex/backend/internal/pagination"
)

type fakeGymsService struct {
	page        pagination.Paginated[models.NearbyGym]
	err         error
	called      bool
	lat         float64
	lng         float64
	radiusKm    float64
	limit       int
	cursor      *pagination.DistanceAscCursor
}

func (f *fakeGymsService) GetNearby(lat, lng, radiusKm float64, limit int, cursor *pagination.DistanceAscCursor) (pagination.Paginated[models.NearbyGym], error) {
	f.called = true
	f.lat = lat
	f.lng = lng
	f.radiusKm = radiusKm
	f.limit = limit
	if cursor != nil {
		value := *cursor
		f.cursor = &value
	} else {
		f.cursor = nil
	}

	if f.err != nil {
		return pagination.Paginated[models.NearbyGym]{}, f.err
	}

	return f.page, nil
}

func TestGetNearbyGymsSuccess(t *testing.T) {
	service := &fakeGymsService{
		page: pagination.Paginated[models.NearbyGym]{
			Items: []models.NearbyGym{
				{
					ID:            "gym-1",
					Name:          "Downtown Gym",
					Lat:           47.61,
					Lng:           -122.33,
					Address:       "123 Main St",
					DistanceM:     250,
					MachinesCount: 15,
					AvgRating:     floatPtr(4.5),
					PriceFromCents: intPtr(2999),
				},
				{
					ID:            "gym-2",
					Name:          "Uptown Gym",
					Lat:           47.62,
					Lng:           -122.31,
					Address:       "456 Pine St",
					DistanceM:     450,
					MachinesCount: 10,
				},
			},
			HasMore: false,
		},
	}

	h := &Handlers{
		config:      &config.Config{},
		gymsService: service,
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/gyms/nearby?lat=47.6&lng=-122.3", nil)
	res := httptest.NewRecorder()

	h.GetNearbyGyms(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.Code)
	}

	if !service.called {
		t.Fatal("expected service to be called")
	}

	if service.lat != 47.6 || service.lng != -122.3 {
		t.Fatalf("unexpected coordinates: lat=%f lng=%f", service.lat, service.lng)
	}
	if service.radiusKm != 5 {
		t.Fatalf("expected default radius 5, got %f", service.radiusKm)
	}
	if service.limit != 20 {
		t.Fatalf("expected default limit 20, got %d", service.limit)
	}
	if service.cursor != nil {
		t.Fatal("expected nil cursor")
	}

	var payload pagination.Paginated[models.NearbyGym]
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(payload.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(payload.Items))
	}
	if payload.HasMore {
		t.Fatal("expected has_more to be false")
	}
	if payload.NextCursor != "" {
		t.Fatal("expected next_cursor to be empty")
	}
}

func TestGetNearbyGymsInvalidCursor(t *testing.T) {
	service := &fakeGymsService{}
	h := &Handlers{
		config:      &config.Config{},
		gymsService: service,
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/gyms/nearby?lat=47.6&lng=-122.3&cursor=not-base64", nil)
	res := httptest.NewRecorder()

	h.GetNearbyGyms(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", res.Code)
	}
	if service.called {
		t.Fatal("expected service not to be called on invalid cursor")
	}
}

func TestGetNearbyGymsServiceError(t *testing.T) {
	service := &fakeGymsService{
		err: httpx.NewError(http.StatusInternalServerError, httpx.ErrorCodeInternal, "boom"),
	}
	h := &Handlers{
		config:      &config.Config{},
		gymsService: service,
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/gyms/nearby?lat=47.6&lng=-122.3", nil)
	res := httptest.NewRecorder()

	h.GetNearbyGyms(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", res.Code)
	}
}

func TestGetNearbyGymsServiceUnavailable(t *testing.T) {
	h := &Handlers{
		config: &config.Config{},
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/gyms/nearby?lat=47.6&lng=-122.3", nil)
	res := httptest.NewRecorder()

	h.GetNearbyGyms(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", res.Code)
	}
}

func TestGetNearbyGymsInvalidRadius(t *testing.T) {
	service := &fakeGymsService{}
	h := &Handlers{
		config:      &config.Config{},
		gymsService: service,
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/gyms/nearby?lat=47.6&lng=-122.3&radius_km=-1", nil)
	res := httptest.NewRecorder()

	h.GetNearbyGyms(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", res.Code)
	}
}

func TestGetNearbyGymsCursorParsing(t *testing.T) {
	cursorValue := pagination.DistanceAscCursor{
		DistanceM: 100,
		ID:        "gym-1",
	}
	cursor, err := pagination.EncodeCursor(cursorValue)
	if err != nil {
		t.Fatalf("failed to encode cursor: %v", err)
	}

	service := &fakeGymsService{
		err: httpx.NewError(http.StatusInternalServerError, httpx.ErrorCodeInternal, "fail"),
	}
	h := &Handlers{
		config:      &config.Config{},
		gymsService: service,
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/gyms/nearby?lat=47.6&lng=-122.3&cursor="+cursor, nil)
	res := httptest.NewRecorder()

	h.GetNearbyGyms(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", res.Code)
	}

	if service.cursor == nil {
		t.Fatal("expected cursor to be passed to service")
	}
	if service.cursor.DistanceM != 100 || service.cursor.ID != "gym-1" {
		t.Fatalf("unexpected cursor value: %+v", service.cursor)
	}
}

func TestGetNearbyGymsMissingParams(t *testing.T) {
	h := &Handlers{
		config: &config.Config{},
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/gyms/nearby?lat=47.6", nil)
	res := httptest.NewRecorder()

	h.GetNearbyGyms(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", res.Code)
	}
}

func TestGetNearbyGymsInvalidLat(t *testing.T) {
	h := &Handlers{
		config: &config.Config{},
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/gyms/nearby?lat=abc&lng=-122.3", nil)
	res := httptest.NewRecorder()

	h.GetNearbyGyms(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", res.Code)
	}
}

func TestGetNearbyGymsInvalidLng(t *testing.T) {
	h := &Handlers{
		config: &config.Config{},
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/gyms/nearby?lat=47.6&lng=200", nil)
	res := httptest.NewRecorder()

	h.GetNearbyGyms(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", res.Code)
	}
}

func TestGetNearbyGymsLimitClamp(t *testing.T) {
	service := &fakeGymsService{
		err: errors.New("should not reach here"),
	}
	h := &Handlers{
		config:      &config.Config{},
		gymsService: service,
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/gyms/nearby?lat=47.6&lng=-122.3&limit=100", nil)
	res := httptest.NewRecorder()

	h.GetNearbyGyms(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500 due to injected error, got %d", res.Code)
	}
	if service.limit != 50 {
		t.Fatalf("expected limit to be clamped to 50, got %d", service.limit)
	}
}

func floatPtr(v float64) *float64 {
	return &v
}

func intPtr(v int) *int {
	return &v
}
