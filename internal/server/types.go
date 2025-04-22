package server

import "github.com/google/uuid"

type Collection struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type CollectionFull struct {
	Collection
	Cards []Card `json:"cards"`
}

type Card struct {
	ID    uuid.UUID `json:"id"`
	Front string    `json:"front"`
	Back  string    `json:"back"`
}

type RegisterData struct {
	Email    string `json:"email"`
	UserName string `json:"username"`
	Password string `json:"password"`
}

type User struct {
	ID       uuid.UUID `json:"id"`
	Email    string    `json:"email"`
	UserName string    `json:"username"`
	password string
}

func (u *User) Password() string {
	return u.password
}

type File struct {
	ID           uuid.UUID `json:"id"`
	Filename     string    `json:"filename"`
	CollectionID uuid.UUID `json:"collection_id"`
	UserID       uuid.UUID `json:"user_id"`
	// FileKey      string    `json:"file_key"`
}
