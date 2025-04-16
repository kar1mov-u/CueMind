// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0

package database

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Card struct {
	ID           uuid.UUID
	Front        string
	Back         string
	CreatedAt    time.Time
	CollectionID uuid.UUID
}

type Collection struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	Name      string
	UserID    uuid.UUID
}

type File struct {
	ID           uuid.UUID
	CollectionID uuid.UUID
	UserID       uuid.UUID
	FileName     string
	FilePath     string
	UploadedAt   time.Time
	Processed    sql.NullBool
}

type User struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	Username  string
	Email     string
	Password  string
}
