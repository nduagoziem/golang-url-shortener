package models

import (
	"time"

	"github.com/google/uuid"
)

type ShortenerModel struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Code        string
	OriginalUrl string
	CreatedAt   time.Time
}
