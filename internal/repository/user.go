package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/nduagoziem/golang-url-shortener/internal/models"
)

type UserRepositoryInterface interface {
	CreateUser(ctx context.Context, email, fullName, passwordHash string) (*models.User, error)
	FindUserByID(ctx context.Context, id string) (*models.User, error)
	FindUserByEmail(ctx context.Context, email string) (*models.User, error)
}

type UserRepository struct {
	database *pgx.Conn
}

// Enforcing interface compliance at compile-time
var _ UserRepositoryInterface = (*UserRepository)(nil)

// Database init
func NewUserRepository(database *pgx.Conn) *UserRepository {
	return &UserRepository{
		database: database,
	}
}

// Satisfy the interface - Create User
func (pg *UserRepository) CreateUser(ctx context.Context, email, fullName, passwordHash string) (*models.User, error) {

	user := &models.User{
		ID:           uuid.New(),
		Email:        email,
		FullName:     fullName,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if user.IsEmailValid(email) == false {
		return nil, errors.New("Email must contain @")
	}

	q := `INSERT INTO users (id, email, full_name, password_hash, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := pg.database.Exec(ctx, q, user.ID, user.Email, user.FullName, user.PasswordHash, user.CreatedAt, user.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return user, err
}

func (pg *UserRepository) FindUserByEmail(ctx context.Context, email string) (*models.User, error) {

	var user models.User
	var lastLogin sql.NullTime

	q := `SELECT id, email, full_name, password_hash, created_at, updated_at, last_login FROM users WHERE email VALUES $1`

	err := pg.database.QueryRow(ctx, q, email).Scan(
		&user.ID,
		&user.Email,
		&user.FullName,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
		&lastLogin.Valid,
	)

	if err != nil {
		return nil, err
	}

	if lastLogin.Valid {
		user.LastLogin = &lastLogin.Time
	}

	return &user, err
}

func (pg *UserRepository) FindUserByID(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	var lastLogin sql.NullTime

	q := `SELECT id, email, full_name, password_hash, created_at, updated_at, last_login FROM users WHERE id VALUES $1`

	err := pg.database.QueryRow(ctx, q, id).Scan(
		&user.ID,
		&user.Email,
		&user.FullName,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
		&lastLogin.Valid,
	)

	if err != nil {
		return nil, err
	}

	if lastLogin.Valid {
		user.LastLogin = &lastLogin.Time
	}

	return &user, err
}
