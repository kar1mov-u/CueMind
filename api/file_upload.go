package api

import (
	"CueMind/internal/server"
	queue "CueMind/internal/worker-queue"
	"log"

	"fmt"
	"net/http"
	"path/filepath"

	"github.com/google/uuid"
)

func (cfg *Config) UploadFile(w http.ResponseWriter, r *http.Request) {

	userId, err := getIdFromContext(r.Context(), "userID")
	if err != nil {
		RespondWithErr(w, 403, "cannot access userID")
		return
	}

	collectionId, err := getIdFromPath(r, "collectionID")
	if err != nil {
		RespondWithErr(w, 403, "cannot access collectionID")
		return
	}

	//parse the data
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		RespondWithErr(w, 400, "unable to parse form")
		return
	}

	//Get File
	rFile, handler, err := r.FormFile("file")
	if err != nil {
		RespondWithErr(w, 400, "file missing")
		return
	}
	defer rFile.Close()

	//Sanitize filename
	uuid := uuid.New()
	filename := fmt.Sprintf("%s-%s", uuid.String(), filepath.Base(handler.Filename))

	//upload it to the s3
	err = cfg.Server.UploadFile(rFile, filename)
	if err != nil {
		RespondWithErr(w, 500, fmt.Sprintf("cannot upload to storage : %v", err))
		return
	}

	file := server.File{
		Filename:     filename,
		CollectionID: collectionId,
		UserID:       userId,
	}
	file_path := createPath(filename)
	file.SetFilepath(file_path)

	//create entry in the DB
	err = cfg.Server.CreateFile(r.Context(), &file)
	if err != nil {
		//TO-DO --- DO some logic if cannot create DB entry for file
		RespondWithErr(w, 500, err.Error())
		return
	}

	//send it to the queue
	queueMsg := queue.Message{
		UserID:       userId,
		CollectionID: collectionId,
		FileKey:      filename,
	}

	err = cfg.Queue.PublishTask(queueMsg)
	if err != nil {
		log.Println(err)
		RespondWithErr(w, 500, "cannot publish to queue")
	}

	//send file to the worker

	RespondWithJson(w, 200, map[string]string{
		"status":   "success",
		"filename": filename,
	})

}

func (cfg *Config) GetFilesForCollection(w http.ResponseWriter, r *http.Request) {
	userID, err := getIdFromContext(r.Context(), "userID")
	if err != nil {
		RespondWithErr(w, 403, "cannot access userID")
		return
	}

	collectionID, err := getIdFromPath(r, "collectionID")
	if err != nil {
		RespondWithErr(w, 403, "cannot access collectionID")
		return
	}

	files, err := cfg.Server.GetFilesForCollection(r.Context(), collectionID, userID)
	if err != nil {
		RespondWithErr(w, 500, err.Error())
		return
	}
	RespondWithJson(w, 200, files)
}
