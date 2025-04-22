package api

import (
	"CueMind/internal/server"
	queue "CueMind/internal/worker-queue"
	"encoding/json"
	"log"

	"fmt"
	"net/http"

	"github.com/google/uuid"
)

// func (cfg *Config) UploadFile(w http.ResponseWriter, r *http.Request) {

// 	userId, err := getIdFromContext(r.Context(), "userID")
// 	if err != nil {
// 		RespondWithErr(w, 403, "cannot access userID")
// 		return
// 	}

// 	collectionId, err := getIdFromPath(r, "collectionID")
// 	if err != nil {
// 		RespondWithErr(w, 403, "cannot access collectionID")
// 		return
// 	}

// 	//parse the data
// 	if err := r.ParseMultipartForm(10 << 20); err != nil {
// 		RespondWithErr(w, 400, "unable to parse form")
// 		return
// 	}

// 	//Get File
// 	rFile, handler, err := r.FormFile("file")
// 	if err != nil {
// 		RespondWithErr(w, 400, "file missing")
// 		return
// 	}
// 	defer rFile.Close()

// 	//Sanitize filename
// 	uuid := uuid.New()
// 	filename := fmt.Sprintf("%s-%s", uuid.String(), filepath.Base(handler.Filename))

// 	//upload it to the s3
// 	err = cfg.Server.UploadFile(rFile, filename)
// 	if err != nil {
// 		RespondWithErr(w, 500, fmt.Sprintf("cannot upload to storage : %v", err))
// 		return
// 	}

// 	file := server.File{
// 		Filename:     filename,
// 		CollectionID: collectionId,
// 		UserID:       userId,
// 	}
// 	// file_path := createPath(filename)
// 	// file.SetFilepath(file_path)

// 	//create entry in the DB
// 	err = cfg.Server.CreateFile(r.Context(), &file)
// 	if err != nil {
// 		//TO-DO --- DO some logic if cannot create DB entry for file
// 		RespondWithErr(w, 500, err.Error())
// 		return
// 	}

// 	//send file to the worker

// 	RespondWithJson(w, 200, map[string]string{
// 		"status":   "success",
// 		"filename": filename,
// 	})

// }

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

func (cfg *Config) GeneratePresignedUrl(w http.ResponseWriter, r *http.Request) {
	collectionID, err := getIdFromPath(r, "collectionID")
	if err != nil {
		RespondWithErr(w, http.StatusBadRequest, err.Error())
		return
	}

	userID, err := getIdFromContext(r.Context(), "userID")
	if err != nil {
		RespondWithErr(w, http.StatusBadRequest, err.Error())
		return
	}

	file := server.File{UserID: userID, CollectionID: collectionID}

	//Save into DB
	err = cfg.Server.CreateFileEntry(r.Context(), &file)
	if err != nil {
		RespondWithErr(w, 500, err.Error())
		return
	}
	//use fileID as the object key in S3
	presignURL, err := cfg.Server.GeneratePresignUrl(r.Context(), file.ID.String())

	RespondWithJson(w, 200, map[string]string{"presignedurl": presignURL, "object_key": file.ID.String()})
}

func (cfg *Config) VerifyUpload(w http.ResponseWriter, r *http.Request) {
	collectionID, err := getIdFromPath(r, "collectionID")
	if err != nil {
		RespondWithErr(w, http.StatusBadRequest, err.Error())
		return
	}

	userID, err := getIdFromContext(r.Context(), "userID")
	if err != nil {
		RespondWithErr(w, http.StatusBadRequest, err.Error())
		return
	}
	//get data from request body
	type Verify struct {
		Status    string `json:"status"`
		ObjectKey string `json:"object_key"`
		FileName  string `json:"file_name"`
		Error     string `json:"error"`
	}
	var verify Verify
	err = json.NewDecoder(r.Body).Decode(&verify)
	if err != nil {
		RespondWithErr(w, http.StatusBadRequest, fmt.Sprintf("Cannot Decode Json :%v", err))
		return
	}

	//convert objetKey to valid UUID
	fileID, err := uuid.Parse(verify.ObjectKey)
	if err != nil {
		RespondWithErr(w, http.StatusBadGateway, err.Error())
		return
	}

	if verify.Status != "success" {
		err = cfg.Server.DeleteFile(r.Context(), fileID)
		if err != nil {
			log.Printf("ERROR: CANNOT DELETE FILE FROM THE DB: %v", err)
		}
		RespondWithJson(w, 200, "ok")
		return
	}

	//update file details to DB
	file := server.File{Filename: verify.FileName, UserID: userID, CollectionID: collectionID, ID: fileID}
	err = cfg.Server.AddFileName(r.Context(), file)
	if err != nil {
		RespondWithErr(w, 404, err.Error())
		return
	}

	//check if the file is not processed
	processed, err := cfg.Server.Processed(r.Context(), fileID)
	if err != nil {
		RespondWithErr(w, http.StatusInternalServerError, fmt.Sprintf("Error checking if file is processed: %v", err))
		return
	} // if processed return success
	if processed {
		RespondWithJson(w, 200, map[string]string{"status": "success", "filename": file.Filename})
		return
	}

	//send it to the queue
	queueMsg := queue.Message{
		UserID:       userID,
		CollectionID: collectionID,
		FileKey:      verify.ObjectKey,
	}

	err = cfg.Queue.PublishTask(queueMsg)
	if err != nil {
		log.Println(err)
		RespondWithErr(w, 500, "cannot publish to queue")
	}

	RespondWithJson(w, 200, map[string]string{"status": "success", "filename": file.Filename})

}
