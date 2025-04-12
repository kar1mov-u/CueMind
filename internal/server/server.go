package server

import (
	"CueMind/internal/database"
	"CueMind/internal/llm"
	"context"
	"fmt"

	"github.com/google/uuid"
)

type Server struct {
	LLM *llm.LLMService
	DB  *database.Queries
}

func NewServer(llm *llm.LLMService, db *database.Queries) *Server {
	return &Server{LLM: llm, DB: db}
}

func (s *Server) CreateCollection(ctx context.Context, userId uuid.UUID, collec *Collection) error {
	id, err := s.DB.CreateCollection(ctx, database.CreateCollectionParams{Name: collec.Name, UserID: userId})
	if err != nil {
		return fmt.Errorf("error on creating collection: %v", err)
	}
	collec.ID = id
	return nil
}

func (s *Server) GetCollection(ctx context.Context, userId uuid.UUID, collectId uuid.UUID) (*CollectionFull, error) {
	// var collectionFull CollectionFull
	var cards []Card

	dbCollection, err := s.DB.GetCollectionById(ctx, database.GetCollectionByIdParams{ID: collectId, UserID: userId})
	if err != nil {
		return nil, fmt.Errorf("error on gettig collection: %v", err)
	}

	dbCards, err := s.DB.GetCardsFomCollection(ctx, collectId)
	if err != nil {
		return nil, fmt.Errorf("error on getting cards: %v", err)
	}
	for i, _ := range dbCards {
		var card Card
		card.Back = dbCards[i].Back
		card.Front = dbCards[i].Front
		cards = append(cards, card)
	}
	collection := Collection{Name: dbCollection.Name, ID: dbCollection.ID}
	return &CollectionFull{Collection: collection, Cards: cards}, nil

}

func (s *Server) ListCollections(ctx context.Context, userID uuid.UUID) ([]Collection, error) {
	dbCollections, err := s.DB.ListCollections(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error on listing collections: %v", err)
	}

	var collections []Collection
	for i, _ := range dbCollections {
		var collection Collection
		collection.ID = dbCollections[i].ID
		collection.Name = dbCollections[i].Name

		collections = append(collections, collection)
	}

	return collections, nil
}
