package handlers

import (
    "bytes"
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"
    "time"

    "fitonex/backend/internal/config"
    "fitonex/backend/internal/models"
    "fitonex/backend/internal/pagination"
    "fitonex/backend/internal/ratelimit"
)

type fakeStorage struct {
	calls []struct {
		Key         string
		ContentType string
		Size        int64
	}
	videoURL string
	thumbURL string
	err      error
}

func (f *fakeStorage) PresignPut(_ context.Context, key, contentType string, sizeBytes int64, _ time.Duration) (string, error) {
	f.calls = append(f.calls, struct {
		Key         string
		ContentType string
		Size        int64
	}{key, contentType, sizeBytes})
	if f.err != nil {
		return "", f.err
	}
	if strings.HasSuffix(key, ".mp4") || strings.HasSuffix(key, ".mov") {
		return f.videoURL, nil
	}
	return f.thumbURL, nil
}

type fakeLimiter struct {
	decision ratelimit.Decision
	err      error
	called   bool
}

func (f *fakeLimiter) Allow(ctx context.Context, key string) (ratelimit.Decision, error) {
	_ = ctx
	_ = key
	f.called = true
	if f.err != nil {
		return ratelimit.Decision{}, f.err
	}
	return f.decision, nil
}

type fakeMachineService struct {
	err error
}

func (f *fakeMachineService) GetByID(id string) (*models.Machine, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &models.Machine{ID: id, Name: "Test Machine"}, nil
}

type fakeVideoService struct {
	createInput struct {
		MachineID   string
		UploaderID  string
		Title       string
		Description *string
		VideoKey    string
		ThumbKey    *string
		DurationSec *int
	}
	createErr error

	listInput struct {
		MachineID string
		Limit     int
		Cursor    *pagination.TimeDescCursor
	}
	listPage pagination.Paginated[models.InstructionVideo]
	listErr  error

	getID    string
	getVideo *models.InstructionVideo
	getErr   error

	likeInput struct {
		VideoID string
		UserID  string
	}
	likeErr error

	unlikeInput struct {
		VideoID string
		UserID  string
	}
	unlikeErr error
}

func (f *fakeVideoService) Create(machineID, uploaderID, title string, description *string, videoKey string, thumbKey *string, durationSec *int) (*models.InstructionVideo, error) {
	f.createInput = struct {
		MachineID   string
		UploaderID  string
		Title       string
		Description *string
		VideoKey    string
		ThumbKey    *string
		DurationSec *int
	}{machineID, uploaderID, title, description, videoKey, thumbKey, durationSec}
	if f.createErr != nil {
		return nil, f.createErr
	}
	return &models.InstructionVideo{ID: "video-1", MachineID: machineID, Title: title, VideoKey: videoKey}, nil
}

func (f *fakeVideoService) ListByMachine(machineID string, limit int, cursor *pagination.TimeDescCursor) (pagination.Paginated[models.InstructionVideo], error) {
	f.listInput = struct {
		MachineID string
		Limit     int
		Cursor    *pagination.TimeDescCursor
	}{machineID, limit, cursor}
	if f.listErr != nil {
		return pagination.Paginated[models.InstructionVideo]{}, f.listErr
	}
	return f.listPage, nil
}

func (f *fakeVideoService) GetByID(id string) (*models.InstructionVideo, error) {
	f.getID = id
	if f.getErr != nil {
		return nil, f.getErr
	}
	if f.getVideo != nil {
		return f.getVideo, nil
	}
	return &models.InstructionVideo{ID: id, Title: "Sample"}, nil
}

func (f *fakeVideoService) LikeVideo(videoID, userID string) error {
	f.likeInput = struct {
		VideoID string
		UserID  string
	}{videoID, userID}
	return f.likeErr
}

func (f *fakeVideoService) UnlikeVideo(videoID, userID string) error {
	f.unlikeInput = struct {
		VideoID string
		UserID  string
	}{videoID, userID}
	return f.unlikeErr
}

func TestGetUploadURLSuccess(t *testing.T) {
	storage := &fakeStorage{
		videoURL: "https://example.com/video",
		thumbURL: "https://example.com/thumb",
	}
	h := &Handlers{
		config: &config.Config{MaxUploadMB: 100},
	}
	h.SetObjectStorage(storage)

	body, _ := json.Marshal(UploadURLRequest{
		MachineID:   "machine-1",
		Title:       "Deep Squat",
		Description: "demo",
		ContentType: "video/mp4",
		Bytes:       1024,
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/videos/upload-url", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), "userID", "user-1"))
	res := httptest.NewRecorder()

	h.GetUploadURL(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}

	var payload UploadURLResponse
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if payload.UploadURL != storage.videoURL {
		t.Fatalf("unexpected upload_url %s", payload.UploadURL)
	}
	if payload.ThumbUploadURL != storage.thumbURL {
		t.Fatalf("unexpected thumb url %s", payload.ThumbUploadURL)
	}
	if len(storage.calls) != 2 {
		t.Fatalf("expected 2 storage calls, got %d", len(storage.calls))
	}
}

func TestGetUploadURLRateLimited(t *testing.T) {
	h := &Handlers{
		config: &config.Config{MaxUploadMB: 100},
	}
	h.SetObjectStorage(&fakeStorage{
		videoURL: "https://example.com/video",
		thumbURL: "https://example.com/thumb",
	})
	h.SetUploadLimiter(&fakeLimiter{
		decision: ratelimit.Decision{Allowed: false, RetryAfter: 5 * time.Second},
	})

	body, _ := json.Marshal(UploadURLRequest{
		MachineID:   "machine-1",
		Title:       "Deep Squat",
		ContentType: "video/mp4",
		Bytes:       2048,
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/videos/upload-url", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), "userID", "user-1"))
	res := httptest.NewRecorder()

	h.GetUploadURL(res, req)

	if res.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", res.Code)
	}
}

func TestFinalizeVideoSuccess(t *testing.T) {
	videoSvc := &fakeVideoService{}
	machineSvc := &fakeMachineService{}

	h := &Handlers{
		config: &config.Config{},
	}
	h.SetVideoService(videoSvc)
	h.SetMachineService(machineSvc)

	body, _ := json.Marshal(FinalizeVideoRequest{
		MachineID:   "machine-1",
		Title:       "  Deep Squat ",
		Description: "  demo ",
		VideoKey:    "videos/abc.mp4",
		ThumbKey:    "thumbs/abc.jpg",
		DurationSec: 120,
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/videos/finalize", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), "userID", "user-1"))
	res := httptest.NewRecorder()

	h.FinalizeVideo(res, req)

	if res.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", res.Code)
	}

	if videoSvc.createInput.Title != "Deep Squat" {
		t.Fatalf("expected trimmed title, got %q", videoSvc.createInput.Title)
	}
	if videoSvc.createInput.Description == nil || *videoSvc.createInput.Description != "demo" {
		t.Fatalf("expected trimmed description, got %v", videoSvc.createInput.Description)
	}
	if videoSvc.createInput.DurationSec == nil || *videoSvc.createInput.DurationSec != 120 {
		t.Fatalf("unexpected duration %v", videoSvc.createInput.DurationSec)
	}
}

func TestFinalizeVideoValidation(t *testing.T) {
	h := &Handlers{
		config: &config.Config{},
	}

	body, _ := json.Marshal(FinalizeVideoRequest{
		MachineID: "",
		Title:     "",
		VideoKey:  "",
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/videos/finalize", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), "userID", "user-1"))
	res := httptest.NewRecorder()

	h.FinalizeVideo(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.Code)
	}
}

func TestGetVideosSuccess(t *testing.T) {
	videoSvc := &fakeVideoService{
		listPage: pagination.Paginated[models.InstructionVideo]{
			Items: []models.InstructionVideo{
				{ID: "v1", CreatedAt: time.Now()},
				{ID: "v2", CreatedAt: time.Now().Add(-time.Minute)},
			},
			HasMore: true,
			NextCursor: "cursor123",
		},
	}

	h := &Handlers{
		config: &config.Config{},
	}
	h.SetVideoService(videoSvc)

	req := httptest.NewRequest(http.MethodGet, "/v1/videos?machine_id=machine-1&limit=30", nil)
	res := httptest.NewRecorder()

	h.GetVideos(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}

	if videoSvc.listInput.Limit != 30 {
		t.Fatalf("expected limit 30, got %d", videoSvc.listInput.Limit)
	}

	var payload pagination.Paginated[models.InstructionVideo]
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !payload.HasMore || payload.NextCursor != "cursor123" {
		t.Fatalf("unexpected pagination payload: %+v", payload)
	}
}

func TestGetVideosInvalidCursor(t *testing.T) {
	h := &Handlers{
		config: &config.Config{},
	}
	h.SetVideoService(&fakeVideoService{})

	req := httptest.NewRequest(http.MethodGet, "/v1/videos?machine_id=machine-1&cursor=invalid", nil)
	res := httptest.NewRecorder()

	h.GetVideos(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.Code)
	}
}
