package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"fitonex/backend/internal/analytics"
	"fitonex/backend/internal/cache"
	"fitonex/backend/internal/config"
	"fitonex/backend/internal/flags"
	"fitonex/backend/internal/handlers"
	"fitonex/backend/internal/observability"
	"fitonex/backend/internal/ratelimit"
	"fitonex/backend/internal/redisclient"
	"fitonex/backend/internal/storage"
	"fitonex/backend/internal/store"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/redis/go-redis/v9"
)

type Server struct {
	config    *config.Config
	store     *store.Store
	router    *chi.Mux
	handlers  *handlers.Handlers

	logger    *slog.Logger
	cache     *cache.Cache
	analytics *analytics.Emitter
	flags     *flags.Manager
	tracker   *observability.Tracker
	alerter   *observability.Alerter

	redisClient *redis.Client
	httpServer  *http.Server
}

func New(cfg *config.Config) *Server {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	store := store.New(cfg)
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(requestLogger(logger))
	router.Use(recovery(logger))
	router.Use(middleware.Timeout(60 * time.Second))

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"ETag"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	handlerSet := handlers.New(store, cfg)
	setupRoutes(router, handlerSet)

	return &Server{
		config:   cfg,
		store:    store,
		router:   router,
		handlers: handlerSet,
		logger:   logger,
		tracker:  observability.NewTracker(logger),
		alerter:  observability.NewAlerter(cfg.AlertWebhookURL),
		flags:    flags.New(cfg.FeatureFlags),
	}
}

func (s *Server) Start() error {
	if err := s.store.Connect(); err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	if err := s.store.MigrateUp(); err != nil {
		return fmt.Errorf("migrate up: %w", err)
	}

	redisClient, err := redisclient.New(s.config)
	if err != nil {
		return fmt.Errorf("init redis: %w", err)
	}
	s.redisClient = redisClient
	s.cache = cache.New(redisClient)
	s.analytics = analytics.NewEmitter(s.config.AnalyticsSink, s.logger)

	uploadLimiter := ratelimit.NewTokenBucket(redisClient, "videos:upload", 5, time.Minute)
	reportLimiter := ratelimit.NewTokenBucket(redisClient, "reports", 5, time.Minute)

	storageService, err := storage.NewS3Service(s.config)
	if err != nil {
		return fmt.Errorf("init storage: %w", err)
	}

	s.handlers.SetObjectStorage(storageService)
	s.handlers.SetUploadLimiter(uploadLimiter)
	s.handlers.SetReportLimiter(reportLimiter)
	s.handlers.SetCache(s.cache)
	s.handlers.SetAnalytics(s.analytics)
	s.handlers.SetFlags(s.flags)
	s.handlers.SetModerationEnabled(s.config.ModerationEnabled)
	s.handlers.SetCDNBaseURL(s.config.CDNBaseURL)

	s.httpServer = &http.Server{
		Addr:    ":" + s.config.Port,
		Handler: s.router,
	}

	go s.watchHealth()

	s.logger.Info("server starting", slog.String("port", s.config.Port))
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer != nil {
		shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("http shutdown: %w", err)
		}
	}
	if s.redisClient != nil {
		if err := s.redisClient.Close(); err != nil {
			return fmt.Errorf("redis close: %w", err)
		}
	}
	if err := s.store.Close(); err != nil {
		return fmt.Errorf("store close: %w", err)
	}
	return nil
}

func (s *Server) watchHealth() {
	if s.alerter == nil {
		return
	}
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	failures := 0
	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err := s.handlers.HealthDetailsCheck(ctx)
		cancel()
		if err != nil {
			failures++
			if failures >= 3 {
				payload := fmt.Sprintf(`{"alert":"health_check_failed","error":"%s"}`, err.Error())
				_ = s.alerter.Notify(context.Background(), []byte(payload))
				failures = 0
			}
			continue
		}
		failures = 0
	}
}

func setupRoutes(r *chi.Mux, h *handlers.Handlers) {
	r.Get("/healthz", h.HealthCheck)

	r.Route("/v1", func(r chi.Router) {
		r.Get("/flags", h.GetFeatureFlags)
		r.Get("/search", h.Search)

		r.Get("/gyms/nearby", h.GetNearbyGyms)
		r.Get("/gyms/{id}", h.GetGym)
		r.Get("/gyms/{id}/machines", h.GetGymMachines)
		r.Get("/gyms/{id}/prices", h.GetGymPrices)
		r.Get("/gyms/{id}/reviews", h.GetGymReviews)

		r.Get("/machines", h.SearchMachines)
		r.Get("/machines/body-parts", h.GetBodyParts)
		r.Get("/machines/{id}", h.GetMachine)

		r.Get("/videos", h.GetVideos)
		r.Get("/videos/{id}", h.GetVideo)

		r.Post("/auth/register", h.Register)
		r.Post("/auth/login", h.Login)

		r.Post("/reports", h.CreateReport)

		r.Route("/", func(r chi.Router) {
			r.Use(h.AuthMiddleware)

			r.Get("/profile", h.GetProfile)
			r.Put("/profile", h.UpdateProfile)

			r.Get("/workouts", h.GetWorkouts)
			r.Post("/workouts", h.CreateWorkout)
			r.Get("/workouts/{id}", h.GetWorkout)
			r.Put("/workouts/{id}", h.UpdateWorkout)
			r.Delete("/workouts/{id}", h.DeleteWorkout)

			r.Post("/gyms/{id}/reviews", h.CreateGymReview)

			r.Post("/videos/upload-url", h.GetUploadURL)
			r.Post("/videos/finalize", h.FinalizeVideo)
			r.Post("/videos/{id}/like", h.LikeVideo)
			r.Delete("/videos/{id}/like", h.UnlikeVideo)

			r.Post("/checkins/today", h.CheckinToday)
			r.Get("/checkins/me", h.GetCheckinStats)

			r.Post("/exercises", h.CreateExercise)
			r.Get("/exercises", h.GetExercises)
			r.Get("/exercises/{id}", h.GetExercise)
		})
	})

	r.Route("/v1/_admin", func(r chi.Router) {
		r.Get("/health/details", h.HealthDetails)
	})
}

func requestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := &responseWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(ww, r)
			duration := time.Since(start)
			logger.Info("http_request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", ww.status),
				slog.String("duration", duration.String()),
			)
		})
	}
}

func recovery(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.Error("panic", slog.Any("error", rec))
					http.Error(w, "internal server error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (w *responseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}
