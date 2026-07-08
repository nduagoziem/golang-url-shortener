package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/nduagoziem/golang-url-shortener/internal/models"
	"github.com/nduagoziem/golang-url-shortener/internal/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
	ErrExpiredToken       = errors.New("token has expired")
	ErrEmailInUse         = errors.New("email already in use")
)

// JWTService provides authentication functionality
type AuthService struct {
	user           *repository.UserRepository
	refreshToken   *repository.RefreshTokenRepository
	jwtSecret      []byte
	accessTokenTTL time.Duration
}

// NewJWTService creates a new authentication service
func NewAuthService(user *repository.UserRepository, refreshToken *repository.RefreshTokenRepository, jwtSecret string, accessTokenTTL time.Duration) *AuthService {
	return &AuthService{
		user:           user,
		refreshToken:   refreshToken,
		jwtSecret:      []byte(jwtSecret),
		accessTokenTTL: accessTokenTTL,
	}
}

// Register creates a new user with the provided credentials
func (s *AuthService) Register(ctx context.Context, email, fullName, password string) (*models.User, error) {
	// Check if user already exists
	_, err := s.user.FindUserByEmail(ctx, email)
	if err == nil {
		return nil, ErrEmailInUse
	}

	// Only proceed if the error was "user not found"
	if !errors.Is(err, sql.ErrNoRows) {
		fmt.Println(err)
		return nil, err
	}

	// Hash the password
	hashedPassword, err := HashPassword(password)
	if err != nil {
		fmt.Println(err)

		return nil, err
	}

	// Create the user
	user, err := s.user.CreateUser(ctx, email, fullName, hashedPassword)
	if err != nil {
		fmt.Println(err)

		return nil, err
	}

	return user, nil
}

// Login authenticates a user and returns an access token
func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	// Get the user from the database
	user, err := s.user.FindUserByEmail(ctx, email)
	if err != nil {
		return "", ErrInvalidCredentials
	}

	// Verify the password
	if err := VerifyPassword(user.PasswordHash, password); err != nil {
		return "", ErrInvalidCredentials
	}

	// Generate an access token
	token, err := s.generateAccessToken(user)
	if err != nil {
		return "", err
	}

	return token, nil
}

// generateAccessToken creates a new JWT access token
func (s *AuthService) generateAccessToken(user *models.User) (string, error) {
	// Set the expiration time
	expirationTime := time.Now().Add(s.accessTokenTTL)

	// Create the JWT claims
	claims := jwt.MapClaims{
		"sub":   user.ID.String(),      // subject (user ID)
		"email": user.Email,            // custom claim
		"exp":   expirationTime.Unix(), // expiration time
		"iat":   time.Now().Unix(),     // issued at time
	}

	// Create the token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with our secret key
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken verifies a JWT token and returns the claims
func (s *AuthService) ValidateToken(tokenString string) (jwt.MapClaims, error) {
	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	// Extract and validate claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

// LoginWithRefresh authenticates a user and returns both access and refresh tokens.
//
// Mostly for first time login after creating an account or when the refresh token has expired.
func (s *AuthService) LoginWithRefresh(ctx context.Context, email, password string, refreshTokenTTL time.Duration) (accessToken string, refreshToken string, err error) {
	// Get the user from the database
	user, err := s.user.FindUserByEmail(ctx, email)
	if err != nil {
		return "", "", ErrInvalidCredentials
	}

	// Verify the password
	if err := VerifyPassword(user.PasswordHash, password); err != nil {
		return "", "", ErrInvalidCredentials
	}

	// Generate an access token
	accessToken, err = s.generateAccessToken(user)
	if err != nil {
		return "", "", err
	}

	// Create a refresh token
	token, err := s.refreshToken.CreateRefreshToken(ctx, user.ID, refreshTokenTTL)
	if err != nil {
		return "", "", err
	}

	return accessToken, token.Token, nil
}

// RefreshAccessToken creates a new access token using a refresh token
func (s *AuthService) RefreshAccessToken(ctx context.Context, refreshTokenString string) (string, error) {
	// Retrieve the refresh token
	token, err := s.refreshToken.GetRefreshToken(ctx, refreshTokenString)
	if err != nil {
		return "", ErrInvalidToken
	}

	// Check if the token is valid
	if token.Revoked {
		return "", ErrInvalidToken
	}

	// Check if the token has expired
	if time.Now().After(token.ExpiresAt) {
		return "", ErrExpiredToken
	}

	// Get the user
	user, err := s.user.FindUserByID(ctx, token.UserID.String())
	if err != nil {
		return "", err
	}

	// Generate a new access token
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return "", err
	}

	return accessToken, nil
}
