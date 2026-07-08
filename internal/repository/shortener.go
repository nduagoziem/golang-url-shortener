package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nduagoziem/golang-url-shortener/internal/models"
)

type ShortenerRepository struct {
	db *pgxpool.Pool
}

func NewShortenerRepository(db *pgxpool.Pool) *ShortenerRepository {
	return &ShortenerRepository{
		db: db,
	}
}

// SaveUrls inserts a new short code together with the original URL record into the database for a given user.
// It takes the original URL, short code, and user ID as parameters.
// The method automatically generates a unique UUID for the URL record and timestamps it with the current time.
// Returns an error if the database insertion fails, such as when the code already exists (UNIQUE constraint violation)
// or if there are database connectivity issues.
func (shortener *ShortenerRepository) SaveUrls(ctx context.Context, originalUrl, code string, userID uuid.UUID) error {
	url := models.ShortenerModel{
		ID:          uuid.New(),
		UserID:      userID,
		Code:        code,
		OriginalUrl: originalUrl,
		CreatedAt:   time.Now(),
	}

	query := `INSERT INTO urls (id, user_id, shortened_url_code, original_url, created_at) VALUES ($1, $2, $3, $4, $5)`

	_, err := shortener.db.Exec(ctx, query, url.ID, url.UserID, url.Code, url.OriginalUrl, url.CreatedAt)

	if err != nil {
		return err
	}

	return nil
}

// FindShortenedUrl retrieves an existing short code for a given original URL and user.
// It queries the database to find a short code that matches both the original URL and user ID.
// Returns the code string if found, or an error if the query fails or no matching record exists.
// Parameters:
//   - ctx: context.Context for request cancellation and timeouts
//   - originalUrl: the original long URL to search for
//   - userID: the UUID of the user who owns the shortened URL
//
// Returns:
//   - string: the short code if found
//   - error: database error or sql.ErrNoRows if no matching record exists
func (shortener *ShortenerRepository) FindShortenedUrl(ctx context.Context, originalUrl string, userID uuid.UUID) (string, error) {
	var code string

	query := `SELECT shortened_url_code FROM urls WHERE original_url = $1 AND user_id = $2`
	err := shortener.db.QueryRow(ctx, query, originalUrl, userID).Scan(&code)

	if err != nil {
		return "", err
	}

	return code, nil
}

// FindOriginalUrl retrieves the original URL.
// It queries the database to find the original URL that matches both the short code and user ID.
// Returns the original URL string if found, or an error if the query fails or no matching record exists.
// Parameters:
//   - ctx: context.Context for request cancellation and timeouts
//   - code: the short code to search for
//   - userID: the UUID of the user who owns the original URL
//
// Returns:
//   - string: the original URL if found
//   - error: database error or sql.ErrNoRows if no matching record exists
func (shortener *ShortenerRepository) FindOriginalUrl(ctx context.Context, code string, userID uuid.UUID) (string, error) {
	var originalUrl string

	query := `SELECT original_url FROM urls WHERE shortened_url_code = $1 AND user_id = $2`
	err := shortener.db.QueryRow(ctx, query, code, userID).Scan(&originalUrl)

	if err != nil {
		return "", err
	}

	return originalUrl, nil
}

func (shortener *ShortenerRepository) FindOriginalUrlPublic(ctx context.Context, code string) (string, error) {
	var originalUrl string

	query := `SELECT original_url FROM urls WHERE shortened_url_code = $1`
	err := shortener.db.QueryRow(ctx, query, code).Scan(&originalUrl)

	if err != nil {
		return "", err
	}

	return originalUrl, nil
}
