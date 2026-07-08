package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nduagoziem/golang-url-shortener/internal/models"
)

type UserRepositoryInterface interface {
	CreateUser(ctx context.Context, email, fullName, passwordHash string) (*models.User, error)
	FindUserByID(ctx context.Context, id string) (*models.User, error)
	FindUserByEmail(ctx context.Context, email string) (*models.User, error)
}

type UserRepository struct {
	database *pgxpool.Pool
}

// Enforcing interface compliance at compile-time
var _ UserRepositoryInterface = (*UserRepository)(nil)

// Database init
func NewUserRepository(database *pgxpool.Pool) *UserRepository {
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

	q := `INSERT INTO users (id, email, full_name, password_hash, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := pg.database.Exec(ctx, q, user.ID, user.Email, user.FullName, user.PasswordHash, user.CreatedAt, user.UpdatedAt)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return user, err
}

func (pg *UserRepository) FindUserByEmail(ctx context.Context, email string) (*models.User, error) {

	var user models.User
	var lastLogin sql.NullTime

	q := `SELECT id, email, full_name, password_hash, created_at, updated_at, last_login FROM users WHERE email = $1`

	err := pg.database.QueryRow(ctx, q, email).Scan(
		&user.ID,
		&user.Email,
		&user.FullName,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
		&lastLogin,
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

	q := `SELECT id, email, full_name, password_hash, created_at, updated_at, last_login FROM users WHERE id = $1`

	err := pg.database.QueryRow(ctx, q, id).Scan(
		&user.ID,
		&user.Email,
		&user.FullName,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
		&lastLogin,
	)

	if err != nil {
		return nil, err
	}

	if lastLogin.Valid {
		user.LastLogin = &lastLogin.Time
	}

	return &user, err
}
