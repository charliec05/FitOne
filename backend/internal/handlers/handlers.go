package handlers

import (
	"context"
	"time"

	"fitonex/backend/internal/analytics"
	"fitonex/backend/internal/cache"
	"fitonex/backend/internal/config"
	"fitonex/backend/internal/flags"
	"fitonex/backend/internal/models"
	"fitonex/backend/internal/notifications"
	"fitonex/backend/internal/oauth"
	"fitonex/backend/internal/payments"
	"fitonex/backend/internal/pagination"
	"fitonex/backend/internal/ratelimit"
	"fitonex/backend/internal/store"
)

type nearbyGymsService interface {
	GetNearby(lat, lng, radiusKm float64, limit int, cursor *pagination.DistanceAscCursor) (pagination.Paginated[models.NearbyGym], error)
}

type machineService interface {
	GetByID(id string) (*models.Machine, error)
}

type videoService interface {
	Create(machineID, uploaderID, title string, description *string, videoKey string, thumbKey *string, durationSec *int, premiumOnly bool) (*models.InstructionVideo, error)
	ListByMachine(machineID string, limit int, cursor *pagination.TimeDescCursor) (pagination.Paginated[models.InstructionVideo], error)
	GetByID(id string) (*models.InstructionVideo, error)
	LikeVideo(videoID, userID string) error
	UnlikeVideo(videoID, userID string) error
	ExportByUser(userID string) ([]models.InstructionVideo, error)
	AnonymizeByUser(userID string) error
	DeleteLikesByUser(userID string) error
}

type objectStorage interface {
	PresignPut(ctx context.Context, key, contentType string, sizeBytes int64, ttl time.Duration) (string, error)
	SignedGet(ctx context.Context, key string, cdnBase string, ttl time.Duration) (string, error)
	Ping(ctx context.Context) error
}

// Handlers contains all HTTP handlers
type Handlers struct {
	store       *store.Store
	config      *config.Config
	gymsService nearbyGymsService
	machineService machineService
	videoService   videoService
	storage     objectStorage
	uploadLimiter ratelimit.Limiter
	reportLimiter ratelimit.Limiter
	cache       *cache.Cache
	analytics   *analytics.Emitter
	flags       *flags.Manager
	moderationEnabled bool
	cdnBaseURL string
	payments   payments.Provider
	emails     notifications.EmailSender
	oauth      *oauth.GoogleVerifier
}

// New creates a new handlers instance
func New(store *store.Store, config *config.Config) *Handlers {
	return &Handlers{
		store:  store,
		config: config,
	}
}

func (h *Handlers) getGymsService() nearbyGymsService {
	if h.gymsService != nil {
		return h.gymsService
	}
	return h.store.Gyms
}

func (h *Handlers) getMachineService() machineService {
	if h.machineService != nil {
		return h.machineService
	}
	return h.store.Machines
}

func (h *Handlers) getVideoService() videoService {
	if h.videoService != nil {
		return h.videoService
	}
	return h.store.Videos
}

func (h *Handlers) storageService() objectStorage {
	return h.storage
}

// SetObjectStorage configures the object storage provider.
func (h *Handlers) SetObjectStorage(storage objectStorage) {
	h.storage = storage
}

// SetUploadLimiter configures the upload rate limiter.
func (h *Handlers) SetUploadLimiter(limiter ratelimit.Limiter) {
	h.uploadLimiter = limiter
}

func (h *Handlers) SetReportLimiter(limiter ratelimit.Limiter) {
	h.reportLimiter = limiter
}

func (h *Handlers) SetCache(cache *cache.Cache) {
	h.cache = cache
}

func (h *Handlers) SetAnalytics(emitter *analytics.Emitter) {
	h.analytics = emitter
}

func (h *Handlers) SetFlags(manager *flags.Manager) {
	h.flags = manager
}

func (h *Handlers) SetModerationEnabled(enabled bool) {
	h.moderationEnabled = enabled
}

func (h *Handlers) SetCDNBaseURL(base string) {
	h.cdnBaseURL = base
}

func (h *Handlers) SetPaymentProvider(provider payments.Provider) {
	h.payments = provider
}

func (h *Handlers) SetEmailSender(sender notifications.EmailSender) {
	h.emails = sender
}

func (h *Handlers) SetOAuthVerifier(verifier *oauth.GoogleVerifier) {
	h.oauth = verifier
}

// SetMachineService overrides the machine service implementation (useful for tests).
func (h *Handlers) SetMachineService(service machineService) {
	h.machineService = service
}

// SetVideoService overrides the video service implementation (useful for tests).
func (h *Handlers) SetVideoService(service videoService) {
	h.videoService = service
}
