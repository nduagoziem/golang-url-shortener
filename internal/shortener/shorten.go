package shortener

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/nduagoziem/golang-url-shortener/internal/repository"
)

type ShortenerService struct {
	shortenerRepository *repository.ShortenerRepository
}

func NewShortenerService(shortenerRepository *repository.ShortenerRepository) *ShortenerService {
	return &ShortenerService{
		shortenerRepository: shortenerRepository,
	}
}

// SaveUrls saves the original URL and generated code in the db,
// and returns the shortened URL for the current host.
func (s *ShortenerService) SaveUrls(hostUrl string, ctx context.Context, originalUrl string, userID uuid.UUID) (string, error) {

	code := generateShortenerCode()

	err := s.shortenerRepository.SaveUrls(ctx, originalUrl, code, userID)
	if err != nil {
		return "", err
	}

	return BuildShortenedUrl(hostUrl, code), err
}

func BuildShortenedUrl(hostUrl string, code string) string {
	return strings.TrimRight(hostUrl, "/") + "/" + code
}

func ExtractCode(shortenedUrlOrCode string) string {
	value := strings.TrimSpace(shortenedUrlOrCode)
	if value == "" {
		return ""
	}

	parsed, err := url.Parse(value)
	if err == nil && parsed.Host != "" {
		value = parsed.Path
	}

	return strings.Trim(strings.TrimRight(value, "/"), "/")
}

func generateShortenerCode() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b) // Ignore error for this basic standard crypto example
	return base64.RawURLEncoding.EncodeToString(b)[:6]
}

// FindShortenedUrl returns the shortened url.
//
// This can be used when a user forgets the shortened one.
func (s *ShortenerService) FindShortenedUrl(hostUrl string, ctx context.Context, originalUrl string, userID uuid.UUID) (string, error) {

	code, err := s.shortenerRepository.FindShortenedUrl(ctx, originalUrl, userID)
	if err != nil {
		return code, err
	}

	return BuildShortenedUrl(hostUrl, ExtractCode(code)), nil
}

func (s *ShortenerService) FindOriginalUrl(ctx context.Context, shortenedUrlOrCode string, userID uuid.UUID) (string, error) {

	url, err := s.shortenerRepository.FindOriginalUrl(ctx, ExtractCode(shortenedUrlOrCode), userID)
	if err != nil {
		return url, err
	}

	return url, nil
}

func (s *ShortenerService) FindOriginalUrlPublic(ctx context.Context, code string) (string, error) {
	url, err := s.shortenerRepository.FindOriginalUrlPublic(ctx, code)
	if err != nil {
		return url, err
	}

	return url, nil
}
