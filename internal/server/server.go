package server

import (
	"CueMind/internal/database"
	"CueMind/internal/llm"
	"CueMind/internal/storage"
	"context"
	"fmt"
	"io"

	"github.com/google/uuid"
)

type Server struct {
	lLM     *llm.LLMService
	dB      *database.Queries
	storage *storage.Storage
}

func NewServer(llm *llm.LLMService, db *database.Queries, storage *storage.Storage) *Server {
	return &Server{lLM: llm, dB: db, storage: storage}
}

func (s *Server) CraeteUser(ctx context.Context, regData RegisterData) (*User, error) {
	dbUser, err := s.dB.CreateUser(ctx, database.CreateUserParams{Username: regData.UserName, Email: regData.Email, Password: regData.Password})
	if err != nil {
		return nil, fmt.Errorf("error on creating user: %v", err)
	}
	return &User{UserName: dbUser.Username, Email: dbUser.Email, ID: dbUser.ID}, nil
}

func (s *Server) GetUser(ctx context.Context, email string) (*User, error) {
	dbUser, err := s.dB.GetUser(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("error on getting user: %v", err)
	}
	return &User{ID: dbUser.ID, Email: dbUser.Email, password: dbUser.Password, UserName: dbUser.Username}, nil
}

func (s *Server) CreateCollection(ctx context.Context, userId uuid.UUID, collec *Collection) error {
	id, err := s.dB.CreateCollection(ctx, database.CreateCollectionParams{Name: collec.Name, UserID: userId})
	if err != nil {
		return fmt.Errorf("error on creating collection: %v", err)
	}
	collec.ID = id
	return nil
}

func (s *Server) GetCollection(ctx context.Context, userId uuid.UUID, collectId uuid.UUID) (*CollectionFull, error) {
	// var collectionFull CollectionFull
	var cards []Card

	dbCollection, err := s.dB.GetCollectionById(ctx, database.GetCollectionByIdParams{ID: collectId, UserID: userId})
	if err != nil {
		return nil, fmt.Errorf("error on gettig collection: %v", err)
	}

	dbCards, err := s.dB.GetCardsFomCollection(ctx, collectId)
	if err != nil {
		return nil, fmt.Errorf("error on getting cards: %v", err)
	}
	for i := range dbCards {
		var card Card
		card.Back = dbCards[i].Back
		card.Front = dbCards[i].Front
		card.ID = dbCards[i].ID
		cards = append(cards, card)
	}
	collection := Collection{Name: dbCollection.Name, ID: dbCollection.ID}
	return &CollectionFull{Collection: collection, Cards: cards}, nil

}

func (s *Server) CheckUserOwnership(ctx context.Context, collectionID, userID uuid.UUID) error {
	v, err := s.dB.CheckUserCollectionOwnership(ctx, database.CheckUserCollectionOwnershipParams{ID: collectionID, UserID: userID})
	if err != nil {
		return fmt.Errorf("user doenst own the collection: %v", err)
	}
	if v != 1 {
		return fmt.Errorf("error ownership res doesnt equals 1")
	}
	return nil
}

func (s *Server) ListCollections(ctx context.Context, userID uuid.UUID) ([]Collection, error) {
	dbCollections, err := s.dB.ListCollections(ctx, userID)
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
	dbCard, err := s.dB.GetCard(ctx, database.GetCardParams{ID: cardID, UserID: userID})
	if err != nil {
		return nil, fmt.Errorf("error on getting card: %v", err)
	}
	return &Card{Front: dbCard.Front, Back: dbCard.Back, ID: cardID}, nil
}

func (s *Server) CreateCard(ctx context.Context, collectionID uuid.UUID, card *Card) error {
	cardID, err := s.dB.CreateCard(ctx, database.CreateCardParams{Front: card.Front, Back: card.Back, CollectionID: collectionID})
	if err != nil {
		return fmt.Errorf("error on creating card: %v", err)
	}
	card.ID = cardID
	return nil
}

func (s *Server) DeleteCard(ctx context.Context, cardID, collectionID uuid.UUID) error {
	err := s.dB.DeleteCard(ctx, database.DeleteCardParams{ID: cardID, CollectionID: collectionID})
	if err != nil {
		return fmt.Errorf("error on deleting card: %v", err)
	}
	return nil
}

func (s *Server) UploadFile(file io.Reader, objectKey string) error {
	ctx := context.Background()
	return s.storage.UploadFile(ctx, objectKey, file)
}

func (s *Server) CreateFile(ctx context.Context, file *File) error {
	fileID, err := s.dB.CreateFile(ctx, database.CreateFileParams{CollectionID: file.CollectionID, UserID: file.UserID, FileName: file.Filename, FilePath: file.filepath})
	if err != nil {
		return fmt.Errorf("error on creating file in DB: %v", err)
	}
	file.ID = fileID
	return nil
}

func (s *Server) GetFilesForCollection(ctx context.Context, collectionID uuid.UUID, userID uuid.UUID) ([]File, error) {
	var files []File
	dbFiles, err := s.dB.GetFilesForCollection(ctx, database.GetFilesForCollectionParams{CollectionID: collectionID, UserID: userID})
	if err != nil {
		return files, fmt.Errorf("error on gettig files from DB: %v", err)
	}
	for i := range dbFiles {
		var file File
		file.Filename = dbFiles[i].FileName
		file.CollectionID = dbFiles[i].CollectionID
		file.UserID = dbFiles[i].UserID
		file.ID = dbFiles[i].ID
		files = append(files, file)
	}
	return files, nil
}
