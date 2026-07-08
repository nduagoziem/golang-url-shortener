package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/nduagoziem/golang-url-shortener/internal/auth"
	"github.com/nduagoziem/golang-url-shortener/internal/db"
	"github.com/nduagoziem/golang-url-shortener/internal/handlers"
	"github.com/nduagoziem/golang-url-shortener/internal/middleware"
	"github.com/nduagoziem/golang-url-shortener/internal/repository"
	"github.com/nduagoziem/golang-url-shortener/internal/shortener"
)

func GenerateJWTSecret(length int) (string, error) {
	bytes := make([]byte, length)

	// Read cryptographically secure random numbers into the slice
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	// Convert the raw bytes to a human-readable hex string
	return hex.EncodeToString(bytes), nil
}

func main() {

	// LOad env variables
	_ = godotenv.Load()

	ctx := context.Background()

	jwtSecret, err := GenerateJWTSecret(32)
	if err != nil {
		log.Fatalf("Failed to generate secret: %v", err)
	}

	r := chi.NewRouter()
	dbPool := db.NewPostgresPool(ctx, os.Getenv("DATABASE_URL"))
	// defer dbPool.Close()

	// Create repos
	userRepo := repository.NewUserRepository(dbPool)
	refreshTokenRepo := repository.NewRefreshTokenRepository(dbPool)
	shortenerRepo := repository.NewShortenerRepository(dbPool)

	// Service
	authService := auth.NewAuthService(userRepo, refreshTokenRepo, jwtSecret, 7*24*time.Hour)
	shortenerService := shortener.NewShortenerService(shortenerRepo)

	// Middleware
	authMiddleWare := middleware.AuthMiddleware(authService)

	// Handlers
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userRepo)
	shortenerHandler := handlers.NewShortenerHandler(shortenerService)

	// Public Routes
	r.Post("/api/v1/auth/register", func(w http.ResponseWriter, r *http.Request) {
		authHandler.Register(r.Context(), w, r)
	})
	r.Post("/api/v1/auth/login", func(w http.ResponseWriter, r *http.Request) {
		authHandler.Login(r.Context(), w, r)
	})
	r.Post("/api/v1/auth/refresh", func(w http.ResponseWriter, r *http.Request) {
		authHandler.RefreshToken(r.Context(), w, r)
	})
	r.Get("/api/v1/url/{code}", func(w http.ResponseWriter, r *http.Request) {
		shortenerHandler.RedirectToOriginalUrl(r.Context(), w, r)
	})

	// Protected Routes
	r.Group(func(pr chi.Router) {
		pr.Use(authMiddleWare)
		pr.Get("/api/v1/user", func(w http.ResponseWriter, r *http.Request) {
			userHandler.Profile(r.Context(), w, r)
		})
		pr.Post("/api/v1/shorten/create", func(w http.ResponseWriter, r *http.Request) {
			shortenerHandler.SaveAndShortenUrl(os.Getenv("HOST_URL"), r.Context(), w, r)
		})
		pr.Get("/api/v1/shorten/original-url", func(w http.ResponseWriter, r *http.Request) {
			shortenerHandler.RetrieveOriginalUrl(r.Context(), w, r)
		})
		pr.Get("/api/v1/shorten/shortened-url", func(w http.ResponseWriter, r *http.Request) {
			shortenerHandler.RetrieveShortenedUrl(os.Getenv("HOST_URL"), r.Context(), w, r)
		})
	})

	// rdb := cache.NewRedisCache(ctx, os.Getenv("REDIS_URL"))
	// if err := rdb.Set(ctx, "username", "nduagoziem", 10*time.Minute); err != nil {
	// 	log.Printf("Redis set failed: %v", err)
	// }

	// username, err := rdb.Get(ctx, "username")
	// if err != nil {
	// 	log.Printf("Redis get failed: %v", err)
	// } else {
	// 	fmt.Println(username)
	// }

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := fmt.Sprintf(":%s", port)

	log.Printf("Starting server on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
