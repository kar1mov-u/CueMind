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
