package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"fitonex/backend/internal/config"
	"fitonex/backend/internal/handlers"
	"fitonex/backend/internal/ratelimit"
	"fitonex/backend/internal/redisclient"
	"fitonex/backend/internal/storage"
	"fitonex/backend/internal/store"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/redis/go-redis/v9"
)

// Server represents the HTTP server
type Server struct {
	config      *config.Config
	store       *store.Store
	router      *chi.Mux
	handlers    *handlers.Handlers
	redisClient *redis.Client
	httpServer  *http.Server
}

// New creates a new server instance
func New(cfg *config.Config) *Server {
	// Initialize store
	store := store.New(cfg)

	// Create router
	router := chi.NewRouter()

	// Add middleware
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(60 * time.Second))

	// CORS configuration
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"}, // Configure appropriately for production
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Create handlers
	handlerSet := handlers.New(store, cfg)

	// Setup routes
	setupRoutes(router, handlerSet)

	return &Server{
		config:   cfg,
		store:    store,
		router:   router,
		handlers: handlerSet,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	// Connect to database
	if err := s.store.Connect(); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Run migrations
	if err := s.store.MigrateUp(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Initialize Redis
	redisClient, err := redisclient.New(s.config)
	if err != nil {
		return fmt.Errorf("failed to initialize redis: %w", err)
	}
	s.redisClient = redisClient

	uploadLimiter := ratelimit.NewTokenBucket(redisClient, "videos:upload", 5, time.Minute)
	s.handlers.SetUploadLimiter(uploadLimiter)

	// Initialize object storage
	storageService, err := storage.NewS3Service(s.config)
	if err != nil {
		return fmt.Errorf("failed to initialize storage service: %w", err)
	}
	s.handlers.SetObjectStorage(storageService)

	// Start server
	s.httpServer = &http.Server{
		Addr:    ":" + s.config.Port,
		Handler: s.router,
	}

	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer != nil {
		shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("http server shutdown: %w", err)
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

// setupRoutes configures all application routes
func setupRoutes(r *chi.Mux, h *handlers.Handlers) {
	// Health check endpoint
	r.Get("/healthz", h.HealthCheck)

	// API v1 routes
	r.Route("/v1", func(r chi.Router) {
		// Public routes
		r.Post("/auth/register", h.Register)
		r.Post("/auth/login", h.Login)
		
		// Public gym routes
		r.Get("/gyms/nearby", h.GetNearbyGyms)
		r.Get("/gyms/{id}", h.GetGym)
		r.Get("/gyms/{id}/machines", h.GetGymMachines)
		r.Get("/gyms/{id}/prices", h.GetGymPrices)
		r.Get("/gyms/{id}/reviews", h.GetGymReviews)
		
		// Public machine routes
		r.Get("/machines", h.SearchMachines)
		r.Get("/machines/body-parts", h.GetBodyParts)
		r.Get("/machines/{id}", h.GetMachine)
		
		// Public video routes
		r.Get("/videos", h.GetVideos)
		r.Get("/videos/{id}", h.GetVideo)

		// Protected routes
		r.Route("/", func(r chi.Router) {
			r.Use(h.AuthMiddleware)
			
			// User routes
			r.Get("/profile", h.GetProfile)
			r.Put("/profile", h.UpdateProfile)
			
			// Workout routes
			r.Get("/workouts", h.GetWorkouts)
			r.Post("/workouts", h.CreateWorkout)
			r.Get("/workouts/{id}", h.GetWorkout)
			r.Put("/workouts/{id}", h.UpdateWorkout)
			r.Delete("/workouts/{id}", h.DeleteWorkout)
			
			// Gym review routes
			r.Post("/gyms/{id}/reviews", h.CreateGymReview)
			
			// Video routes
			r.Post("/videos/upload-url", h.GetUploadURL)
			r.Post("/videos/finalize", h.FinalizeVideo)
			r.Post("/videos/{id}/like", h.LikeVideo)
			r.Delete("/videos/{id}/like", h.UnlikeVideo)
			
			// Check-in routes
			r.Post("/checkins/today", h.CheckinToday)
			r.Get("/checkins/me", h.GetCheckinStats)
			
			// Exercise routes
			r.Post("/exercises", h.CreateExercise)
			r.Get("/exercises", h.GetExercises)
			r.Get("/exercises/{id}", h.GetExercise)
		})
	})
}
