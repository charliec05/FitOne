package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"fitonex/backend/internal/httpx"
	"fitonex/backend/internal/models"
	"fitonex/backend/internal/moderation"
	"fitonex/backend/internal/pagination"
	"fitonex/backend/internal/storage"
	videosstore "fitonex/backend/internal/store/videos"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

const (
	uploadURLTTL         = 15 * time.Minute
	thumbnailContentType = "image/jpeg"
	defaultVideoLimit    = 20
	maxVideoLimit        = 50
)

var allowedVideoContentTypes = map[string]string{
	"video/mp4":       ".mp4",
	"video/quicktime": ".mov",
}

// UploadURLRequest represents the upload URL request payload.
type UploadURLRequest struct {
	MachineID   string `json:"machine_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	ContentType string `json:"content_type"`
	Bytes       int64  `json:"bytes"`
}

// UploadURLResponse represents the presigned URLs response.
type UploadURLResponse struct {
	UploadURL      string `json:"upload_url"`
	VideoKey       string `json:"video_key"`
	ThumbUploadURL string `json:"thumb_upload_url"`
	ThumbKey       string `json:"thumb_key"`
}

// FinalizeVideoRequest represents the finalize payload.
type FinalizeVideoRequest struct {
	MachineID   string `json:"machine_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	VideoKey    string `json:"video_key"`
	ThumbKey    string `json:"thumb_key,omitempty"`
	DurationSec int    `json:"duration_sec"`
}

// GetUploadURL returns pre-signed URLs for uploading video and thumbnail content.
func (h *Handlers) GetUploadURL(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(r)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrorCodeUnauthorized, "unauthorized")
		return
	}

	var req UploadURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "invalid request body")
		return
	}

	if err := h.validateUploadRequest(&req); err != nil {
		httpx.WriteAPIError(w, err)
		return
	}

	if h.uploadLimiter != nil {
		decision, err := h.uploadLimiter.Allow(r.Context(), userID)
		if err != nil {
			httpx.WriteAPIError(w, httpx.WrapError(err, http.StatusInternalServerError, httpx.ErrorCodeInternal, "rate limit check failed"))
			return
		}
		if !decision.Allowed {
			retrySeconds := int(math.Ceil(decision.RetryAfter.Seconds()))
			if retrySeconds <= 0 {
				retrySeconds = 1
			}
			message := fmt.Sprintf("Rate limit exceeded. Try again in %d seconds.", retrySeconds)
			httpx.WriteError(w, http.StatusTooManyRequests, httpx.ErrorCodeTooManyRequests, message)
			return
		}
	}

	storageSvc := h.storageService()
	if storageSvc == nil {
		service, err := storage.NewS3Service(h.config)
		if err != nil {
			httpx.WriteAPIError(w, httpx.WrapError(err, http.StatusInternalServerError, httpx.ErrorCodeInternal, "failed to initialize storage service"))
			return
		}
		h.SetObjectStorage(service)
		storageSvc = service
	}

	ext := allowedVideoContentTypes[req.ContentType]
	videoKey := fmt.Sprintf("videos/%s%s", uuid.NewString(), ext)
	thumbKey := fmt.Sprintf("videos/thumbs/%s.jpg", uuid.NewString())

    videoURL, err := storageSvc.PresignPut(r.Context(), videoKey, req.ContentType, req.Bytes, uploadURLTTL)
	if err != nil {
		httpx.WriteAPIError(w, httpx.WrapError(err, http.StatusInternalServerError, httpx.ErrorCodeInternal, "failed to generate upload url"))
		return
	}

	thumbURL, err := storageSvc.PresignPut(r.Context(), thumbKey, thumbnailContentType, 0, uploadURLTTL)
	if err != nil {
		httpx.WriteAPIError(w, httpx.WrapError(err, http.StatusInternalServerError, httpx.ErrorCodeInternal, "failed to generate thumbnail upload url"))
		return
	}

    resp := UploadURLResponse{
        UploadURL:      videoURL,
        VideoKey:       videoKey,
        ThumbUploadURL: thumbURL,
        ThumbKey:       thumbKey,
    }

    if h.analytics != nil {
        h.analytics.EmitEvent(r.Context(), userID, "upload_started", map[string]any{
            "machine_id": req.MachineID,
        })
    }

    httpx.WriteJSON(w, http.StatusOK, resp)
}

// FinalizeVideo persists metadata for an uploaded video.
func (h *Handlers) FinalizeVideo(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(r)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrorCodeUnauthorized, "unauthorized")
		return
	}

	var req FinalizeVideoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "invalid request body")
		return
	}

    if err := h.validateFinalizeRequest(&req); err != nil {
        httpx.WriteAPIError(w, err)
        return
    }

	machineSvc := h.getMachineService()
	if machineSvc == nil {
		httpx.WriteAPIError(w, httpx.NewError(http.StatusInternalServerError, httpx.ErrorCodeInternal, "machine service unavailable"))
		return
	}

	if _, err := machineSvc.GetByID(req.MachineID); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "machine not found")
		return
	}

	videoSvc := h.getVideoService()
	if videoSvc == nil {
		httpx.WriteAPIError(w, httpx.NewError(http.StatusInternalServerError, httpx.ErrorCodeInternal, "video service unavailable"))
		return
	}

	title := strings.TrimSpace(req.Title)
	var description *string
	if trimmed := strings.TrimSpace(req.Description); trimmed != "" {
		description = &trimmed
	}

	var thumbKey *string
	if tk := strings.TrimSpace(req.ThumbKey); tk != "" {
		thumbKey = &tk
	}

	var duration *int
	if req.DurationSec > 0 {
		duration = &req.DurationSec
	}

    video, err := videoSvc.Create(req.MachineID, userID, title, description, req.VideoKey, thumbKey, duration)
    if err != nil {
        httpx.WriteAPIError(w, httpx.WrapError(err, http.StatusInternalServerError, httpx.ErrorCodeInternal, "failed to create video"))
        return
    }

    storageSvc := h.storageService()
    if storageSvc == nil {
        if service, err := storage.NewS3Service(h.config); err == nil {
            h.SetObjectStorage(service)
            storageSvc = service
        }
    }
    if storageSvc != nil {
        if playURL, err := storageSvc.SignedGet(r.Context(), video.VideoKey, h.cdnBaseURL, 5*time.Minute); err == nil {
            video.PlayURL = playURL
        }
    }

    if h.analytics != nil {
        h.analytics.EmitEvent(r.Context(), userID, "video_uploaded", map[string]any{
            "machine_id": req.MachineID,
        })
    }

    httpx.WriteJSON(w, http.StatusCreated, video)
}

// GetVideos returns a paginated list of videos for a machine.
func (h *Handlers) GetVideos(w http.ResponseWriter, r *http.Request) {
	machineID := strings.TrimSpace(r.URL.Query().Get("machine_id"))
	if machineID == "" {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "machine_id parameter is required")
		return
	}

	limit := defaultVideoLimit
	if limitParam := strings.TrimSpace(r.URL.Query().Get("limit")); limitParam != "" {
		value, err := strconv.Atoi(limitParam)
		if err != nil || value <= 0 {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "limit must be a positive integer")
			return
		}
		if value > maxVideoLimit {
			value = maxVideoLimit
		}
		limit = value
	}

	var cursorPtr *pagination.TimeDescCursor
	if cursorStr := strings.TrimSpace(r.URL.Query().Get("cursor")); cursorStr != "" {
		cursor, err := pagination.DecodeCursor[pagination.TimeDescCursor](cursorStr)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "invalid cursor")
			return
		}
		cursorPtr = &cursor
	}

	videoSvc := h.getVideoService()
	if videoSvc == nil {
		httpx.WriteAPIError(w, httpx.NewError(http.StatusInternalServerError, httpx.ErrorCodeInternal, "video service unavailable"))
		return
	}

    page, err := videoSvc.ListByMachine(machineID, limit, cursorPtr)
	if err != nil {
		if errors.Is(err, pagination.ErrInvalidLimit) {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "limit must be greater than zero")
			return
		}
		var apiErr *httpx.APIError
		if errors.As(err, &apiErr) {
			httpx.WriteAPIError(w, apiErr)
			return
		}
		httpx.WriteAPIError(w, httpx.WrapError(err, http.StatusInternalServerError, httpx.ErrorCodeInternal, "failed to list videos"))
		return
	}

    storageSvc := h.storageService()
    if storageSvc == nil {
        service, err := storage.NewS3Service(h.config)
        if err != nil {
            httpx.WriteAPIError(w, httpx.WrapError(err, http.StatusInternalServerError, httpx.ErrorCodeInternal, "storage unavailable"))
            return
        }
        h.SetObjectStorage(service)
        storageSvc = service
    }

    for i := range page.Items {
        playURL, err := storageSvc.SignedGet(r.Context(), page.Items[i].VideoKey, h.cdnBaseURL, 5*time.Minute)
        if err == nil {
            page.Items[i].PlayURL = playURL
        }
    }

    httpx.WriteJSONWithCache(w, http.StatusOK, page, 30*time.Second)
}

// GetVideo returns a single video by ID.
func (h *Handlers) GetVideo(w http.ResponseWriter, r *http.Request) {
	videoID := chi.URLParam(r, "id")

	videoSvc := h.getVideoService()
	if videoSvc == nil {
		httpx.WriteAPIError(w, httpx.NewError(http.StatusInternalServerError, httpx.ErrorCodeInternal, "video service unavailable"))
		return
	}

    video, err := videoSvc.GetByID(videoID)
    if err != nil {
        if errors.Is(err, videosstore.ErrVideoNotFound) {
            httpx.WriteError(w, http.StatusNotFound, httpx.ErrorCodeNotFound, "video not found")
            return
        }
        httpx.WriteAPIError(w, httpx.WrapError(err, http.StatusInternalServerError, httpx.ErrorCodeInternal, "failed to fetch video"))
        return
    }

    if storageSvc := h.storageService(); storageSvc != nil {
        if playURL, err := storageSvc.SignedGet(r.Context(), video.VideoKey, h.cdnBaseURL, 5*time.Minute); err == nil {
            video.PlayURL = playURL
        }
    }

    httpx.WriteJSON(w, http.StatusOK, video)
}

// LikeVideo records a like for the authenticated user.
func (h *Handlers) LikeVideo(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(r)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrorCodeUnauthorized, "unauthorized")
		return
	}

	videoID := chi.URLParam(r, "id")

	videoSvc := h.getVideoService()
	if videoSvc == nil {
		httpx.WriteAPIError(w, httpx.NewError(http.StatusInternalServerError, httpx.ErrorCodeInternal, "video service unavailable"))
		return
	}

	if err := videoSvc.LikeVideo(videoID, userID); err != nil {
		httpx.WriteAPIError(w, httpx.WrapError(err, http.StatusInternalServerError, httpx.ErrorCodeInternal, "failed to like video"))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UnlikeVideo removes a like for the authenticated user.
func (h *Handlers) UnlikeVideo(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(r)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrorCodeUnauthorized, "unauthorized")
		return
	}

	videoID := chi.URLParam(r, "id")

	videoSvc := h.getVideoService()
	if videoSvc == nil {
		httpx.WriteAPIError(w, httpx.NewError(http.StatusInternalServerError, httpx.ErrorCodeInternal, "video service unavailable"))
		return
	}

	if err := videoSvc.UnlikeVideo(videoID, userID); err != nil {
		httpx.WriteAPIError(w, httpx.WrapError(err, http.StatusInternalServerError, httpx.ErrorCodeInternal, "failed to unlike video"))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) validateUploadRequest(req *UploadURLRequest) *httpx.APIError {
	req.MachineID = strings.TrimSpace(req.MachineID)
	if req.MachineID == "" {
		return httpx.NewError(http.StatusBadRequest, httpx.ErrorCodeBadRequest, "machine_id is required")
	}

	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		return httpx.NewError(http.StatusBadRequest, httpx.ErrorCodeBadRequest, "title is required")
	}

	req.ContentType = strings.TrimSpace(req.ContentType)
	if req.ContentType == "" {
		return httpx.NewError(http.StatusBadRequest, httpx.ErrorCodeBadRequest, "content_type is required")
	}

	if _, ok := allowedVideoContentTypes[req.ContentType]; !ok {
		return httpx.NewError(http.StatusBadRequest, httpx.ErrorCodeBadRequest, "unsupported content_type")
	}

	if req.Bytes <= 0 {
		return httpx.NewError(http.StatusBadRequest, httpx.ErrorCodeBadRequest, "bytes must be greater than zero")
	}

	maxBytes := int64(h.config.MaxUploadMB) * 1024 * 1024
	if maxBytes > 0 && req.Bytes > maxBytes {
		message := fmt.Sprintf("file size exceeds limit of %dMB", h.config.MaxUploadMB)
		return httpx.NewError(http.StatusBadRequest, httpx.ErrorCodeBadRequest, message)
	}

	if h.moderationEnabled {
		if err := moderation.ValidateVideoMeta(req.Title, req.Description); err != nil {
			return httpx.NewError(http.StatusBadRequest, httpx.ErrorCodeBadRequest, err.Error())
		}
	}

	return nil
}

func (h *Handlers) validateFinalizeRequest(req *FinalizeVideoRequest) *httpx.APIError {
	req.MachineID = strings.TrimSpace(req.MachineID)
	if req.MachineID == "" {
		return httpx.NewError(http.StatusBadRequest, httpx.ErrorCodeBadRequest, "machine_id is required")
	}

	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		return httpx.NewError(http.StatusBadRequest, httpx.ErrorCodeBadRequest, "title is required")
	}

	req.VideoKey = strings.TrimSpace(req.VideoKey)
	if req.VideoKey == "" {
		return httpx.NewError(http.StatusBadRequest, httpx.ErrorCodeBadRequest, "video_key is required")
	}

	if req.DurationSec < 0 {
		return httpx.NewError(http.StatusBadRequest, httpx.ErrorCodeBadRequest, "duration_sec must be non-negative")
	}

	if h.moderationEnabled {
		if err := moderation.ValidateVideoMeta(req.Title, req.Description); err != nil {
			return httpx.NewError(http.StatusBadRequest, httpx.ErrorCodeBadRequest, err.Error())
		}
	}

	return nil
}
