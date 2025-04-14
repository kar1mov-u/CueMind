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
	for i := range dbCards {
		var card Card
		card.Back = dbCards[i].Back
		card.Front = dbCards[i].Front
		cards = append(cards, card)
	}
	collection := Collection{Name: dbCollection.Name, ID: dbCollection.ID}
	return &CollectionFull{Collection: collection, Cards: cards}, nil

}

func (s *Server) CheckUserOwnership(ctx context.Context, collectionID, userID uuid.UUID) error {
	v, err := s.DB.CheckUserCollectionOwnership(ctx, database.CheckUserCollectionOwnershipParams{ID: collectionID, UserID: userID})
	if err != nil {
		return fmt.Errorf("user doenst own the collection: %v", err)
	}
	if v != 1 {
		return fmt.Errorf("error ownership res doesnt equals 1")
	}
	return nil
}

func (s *Server) ListCollections(ctx context.Context, userID uuid.UUID) ([]Collection, error) {
	dbCollections, err := s.DB.ListCollections(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error on listing collections: %v", err)
	}

	var collections []Collection
	for i := range dbCollections {
		var collection Collection
		collection.ID = dbCollections[i].ID
		collection.Name = dbCollections[i].Name

		collections = append(collections, collection)
	}

	return collections, nil
}

func (s *Server) GetCard(ctx context.Context, userID, cardID uuid.UUID) (*Card, error) {
	dbCard, err := s.DB.GetCard(ctx, database.GetCardParams{ID: cardID, UserID: userID})
	if err != nil {
		return nil, fmt.Errorf("error on getting card: %v", err)
	}
	return &Card{Front: dbCard.Front, Back: dbCard.Back, ID: cardID}, nil
}

func (s *Server) CreateCard(ctx context.Context, collectionID uuid.UUID, card *Card) error {
	cardID, err := s.DB.CreateCard(ctx, database.CreateCardParams{Front: card.Front, Back: card.Back, CollectionID: collectionID})
	if err != nil {
		return fmt.Errorf("error on creating card: %v", err)
	}
	card.ID = cardID
	return nil
}

func (s *Server) DeleteCard(ctx context.Context, cardID, collectionID uuid.UUID) error {
	err := s.DB.DeleteCard(ctx, database.DeleteCardParams{ID: cardID, CollectionID: collectionID})
	if err != nil {
		return fmt.Errorf("error on deleting card: %v", err)
	}
	return nil
}
