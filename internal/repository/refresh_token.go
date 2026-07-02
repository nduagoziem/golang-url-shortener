package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/nduagoziem/golang-url-shortener/internal/models"
)

type RefreshTokenRepositoryInterface interface {
	CreateRefreshToken(ctx context.Context, userID uuid.UUID, ttl time.Duration) (*models.RefreshToken, error)
	GetRefreshToken(ctx context.Context, tokenString string) (*models.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, tokenString string) error
}

// RefreshTokenRepository handles database operations for refresh tokens
type RefreshTokenRepository struct {
	db *pgx.Conn
}

// Enforcing interface compliance at compile-time
var _ RefreshTokenRepositoryInterface = (*RefreshTokenRepository)(nil)

// NewRefreshTokenRepository creates a new refresh token repository
func NewRefreshTokenRepository(db *pgx.Conn) *RefreshTokenRepository {
	return &RefreshTokenRepository{
		db: db,
	}
}

// CreateRefreshToken creates a new refresh token for a user
func (r *RefreshTokenRepository) CreateRefreshToken(ctx context.Context, userID uuid.UUID, ttl time.Duration) (*models.RefreshToken, error) {
	tokenID := uuid.New()
	expiresAt := time.Now().Add(ttl)

	token := &models.RefreshToken{
		ID:        tokenID,
		UserID:    userID,
		Token:     tokenID.String(),
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
		Revoked:   false,
	}

	q := `INSERT INTO refresh_tokens (id, user_id, token, expires_at, created_at, revoked) VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.db.Exec(ctx, q, token.ID, token.UserID, token.Token, token.ExpiresAt, token.CreatedAt, token.Revoked)

	if err != nil {
		return nil, err
	}

	return token, nil
}

// GetRefreshToken retrieves a refresh token by its token string
func (r *RefreshTokenRepository) GetRefreshToken(ctx context.Context, tokenString string) (*models.RefreshToken, error) {
	q := `SELECT id, user_id, token, expires_at, created_at, revoked FROM refresh_token WHERE token $1`

	var token models.RefreshToken

	err := r.db.QueryRow(ctx, q, tokenString).Scan(
		&token.ID,
		&token.UserID,
		&token.Token,
		&token.ExpiresAt,
		&token.CreatedAt,
		&token.Revoked,
	)

	if err != nil {
		return nil, err
	}

	return &token, nil
}

// RevokeRefreshToken marks a refresh token as revoked
func (r *RefreshTokenRepository) RevokeRefreshToken(ctx context.Context, tokenString string) error {
	q := `
        UPDATE refresh_tokens
        SET revoked = true
        WHERE token = $1
    `
	_, err := r.db.Exec(ctx, q, tokenString)
	return err
}
