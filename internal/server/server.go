package server

import (
	"CueMind/internal/database"
	"CueMind/internal/storage"
	"context"
	"database/sql"
	"fmt"
	"io"

	"github.com/google/uuid"
)

type Server struct {
	rawDB   *sql.DB
	dB      *database.Queries
	storage *storage.Storage
}

func New(db *database.Queries, storage *storage.Storage, rawSql *sql.DB) *Server {
	return &Server{dB: db, storage: storage, rawDB: rawSql}
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
	dbCollection, err := s.dB.GetCollectionById(ctx, database.GetCollectionByIdParams{ID: collectId, UserID: userId})
	if err != nil {
		return nil, fmt.Errorf("error on gettig collection: %v", err)
	}

	dbCards, err := s.dB.GetCardsFomCollection(ctx, collectId)
	if err != nil {
		return nil, fmt.Errorf("error on getting cards: %v", err)
	}

	cards := make([]Card, len(dbCards))

	for i := range dbCards {
		var card Card
		card.Back = dbCards[i].Back
		card.Front = dbCards[i].Front
		card.ID = dbCards[i].ID
		cards[i] = card
	}
	collection := Collection{Name: dbCollection.Name, ID: dbCollection.ID}
	return &CollectionFull{Collection: collection, Cards: cards}, nil

}

func (s *Server) DeleteCollection(ctx context.Context, collectionID, userID uuid.UUID) error {
	//create new transaction
	tx, err := s.rawDB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	qtx := s.dB.WithTx(tx)

	//perofrm queires
	err = qtx.DeleteAllCards(ctx, collectionID)
	if err != nil {
		// tx.Rollback()
		return fmt.Errorf("Error on deleteing cards: %v", err)
	}

	err = qtx.DeleteAllFiles(ctx, database.DeleteAllFilesParams{CollectionID: collectionID, UserID: userID})
	if err != nil {
		return fmt.Errorf("Error on deleteing files: %v", err)

	}

	err = qtx.DeleteCollection(ctx, database.DeleteCollectionParams{ID: collectionID, UserID: userID})
	if err != nil {
		// tx.Rollback()
		return fmt.Errorf("Error on deleteing collection: %v", err)

	}
	return tx.Commit()
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

	collections := make([]Collection, len(dbCollections))
	for i := range dbCollections {
		var collection Collection
		collection.ID = dbCollections[i].ID
		collection.Name = dbCollections[i].Name

		collections[i] = collection
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

func (s *Server) GeneratePresignUrl(ctx context.Context, objectKey string) (string, error) {
	url, err := s.storage.GeneratePresignedUrl(ctx, objectKey, 5)
	if err != nil {
		return "", err
	}
	return url, nil
}

func (s *Server) DeleteFile(ctx context.Context, id uuid.UUID) error {
	err := s.dB.DeleteFile(ctx, id)
	if err != nil {
		return fmt.Errorf("Error : cannot delete file :%v", err)
	}
	return nil
}

func (s *Server) AddFileName(ctx context.Context, file File) error {
	err := s.dB.AddFileName(ctx, database.AddFileNameParams{FileName: sql.NullString{String: file.Filename, Valid: true}, CollectionID: file.CollectionID, UserID: file.UserID, ID: file.ID})
	if err != nil {
		return fmt.Errorf("error on deleting file :%v", err)
	}
	return nil
}

func (s *Server) CreateFileEntry(ctx context.Context, file *File) error {
	fileID, err := s.dB.CraeteFileEntry(ctx, database.CraeteFileEntryParams{CollectionID: file.CollectionID, UserID: file.UserID})
	if err != nil {
		return fmt.Errorf("Error on creting file entry in DB : %v", err)
	}
	file.ID = fileID
	return nil
}

func (s *Server) Processed(ctx context.Context, id uuid.UUID) (bool, error) {
	val, err := s.dB.ProcessedCheck(ctx, id)
	if err != nil {
		return false, err
	}
	return val, nil
}

func (s *Server) GetFilesForCollection(ctx context.Context, collectionID uuid.UUID, userID uuid.UUID) ([]File, error) {
	var files []File
	dbFiles, err := s.dB.GetFilesForCollection(ctx, database.GetFilesForCollectionParams{CollectionID: collectionID, UserID: userID})
	if err != nil {
		return files, fmt.Errorf("error on gettig files from DB: %v", err)
	}
	for i := range dbFiles {
		var file File
		file.Filename = dbFiles[i].FileName.String
		file.CollectionID = dbFiles[i].CollectionID
		file.UserID = dbFiles[i].UserID
		file.ID = dbFiles[i].ID
		files = append(files, file)
	}
	return files, nil
}
