package models

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	Email        string     `json:"email" db:"email"`
	FullName     string     `json:"fullName" db:"full_name"`
	PasswordHash string     `json:"_" db:"password_hash"`
	CreatedAt    time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time  `json:"updatedAt" db:"updated_at"`
	LastLogin    *time.Time `json:"lastLogin" db:"last_login"`
}

func (u *User) IsEmailValid(email string) bool {
	return strings.Contains(email, "@")
}
