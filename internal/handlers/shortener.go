package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nduagoziem/golang-url-shortener/internal/middleware"
	"github.com/nduagoziem/golang-url-shortener/internal/shortener"
)

type ShortenerHandler struct {
	shortenerService *shortener.ShortenerService
}

func NewShortenerHandler(shortenerService *shortener.ShortenerService) *ShortenerHandler {
	return &ShortenerHandler{
		shortenerService: shortenerService,
	}
}

type ShortenUrlRequest struct {
	OriginalUrl string `json:"originalUrl"`
}

type ShortenUrlResponse struct {
	ShortenedUrl string `json:"shortenedUrl"`
}

func (h *ShortenerHandler) SaveAndShortenUrl(hostUrl string, ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req ShortenUrlRequest

	if r.Method != http.MethodPost {
		http.Error(w, `Invalid Request Type - Expected POST, gave `+r.Method, http.StatusMethodNotAllowed)
		return
	}

	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid Request Payload", http.StatusBadRequest)
		return
	}

	if req.OriginalUrl == "" {
		http.Error(w, "Original Url missing", http.StatusBadRequest)
		return
	}

	url, err := h.shortenerService.SaveUrls(hostUrl, ctx, req.OriginalUrl, userID)
	if err != nil {
		log.Printf("Error occurred while saving URL: %v", err)
		http.Error(w, `Something went wrong - could not shorten url`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ShortenUrlResponse{
		ShortenedUrl: url,
	})

}

type RetrieveOriginalUrlResponse struct {
	OriginalUrl string `json:"originalUrl"`
}

func (h *ShortenerHandler) RetrieveOriginalUrl(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `Invalid Request Type - Expected GET, gave `+r.Method, http.StatusMethodNotAllowed)
		return
	}

	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		code = r.URL.Query().Get("shortenedUrl")
	}
	if code == "" {
		http.Error(w, "Short code missing", http.StatusBadRequest)
		return
	}

	url, err := h.shortenerService.FindOriginalUrl(ctx, code, userID)
	if err != nil {
		log.Printf("Error occurred while finding original URL: %v", err)
		http.Error(w, `Something went wrong - could not find original url`, http.StatusNotFound)
		return

	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(RetrieveOriginalUrlResponse{
		OriginalUrl: url,
	})

}

type RetrieveShortenedUrlResponse struct {
	ShortenedUrl string `json:"shortenedUrl"`
}

func (h *ShortenerHandler) RetrieveShortenedUrl(hostUrl string, ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `Invalid Request Type - Expected GET, gave `+r.Method, http.StatusMethodNotAllowed)
		return
	}

	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	originalUrl := r.URL.Query().Get("originalUrl")
	if originalUrl == "" {
		http.Error(w, "Original Url missing", http.StatusBadRequest)
		return
	}

	url, err := h.shortenerService.FindShortenedUrl(hostUrl, ctx, originalUrl, userID)
	if err != nil {
		log.Printf("Error occurred while finding shortened URL: %v", err)
		http.Error(w, `Something went wrong - could not find shortened url`, http.StatusNotFound)
		return

	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(RetrieveShortenedUrlResponse{
		ShortenedUrl: url,
	})
}

func (h *ShortenerHandler) RedirectToOriginalUrl(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `Invalid Request Type - Expected GET, gave `+r.Method, http.StatusMethodNotAllowed)
		return
	}

	code := chi.URLParam(r, "code")
	if code == "" {
		http.NotFound(w, r)
		return
	}

	originalUrl, err := h.shortenerService.FindOriginalUrlPublic(ctx, code)
	if err != nil {
		log.Printf("Error occurred while finding public original URL: %v", err)
		http.NotFound(w, r)
		return
	}

	http.Redirect(w, r, originalUrl, http.StatusFound)
}
