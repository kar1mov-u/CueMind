package api

import (
	"CueMind/internal/server"
	"fmt"
	"net/http"
	"path/filepath"
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
	r.ParseMultipartForm(10 << 20)

	//Get File
	rFile, handler, err := r.FormFile("file")
	if err != nil {
		RespondWithErr(w, 400, "file missing")
		return
	}
	defer rFile.Close()

	//Sanitize filename
	filename := filepath.Base(handler.Filename)

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

	//create os file
	// osFile, err := os.Create("./tmp/" + filename)
	// if err != nil {
	// 	RespondWithErr(w, 500, "error on creating file")
	// 	return
	// }
	// defer osFile.Close()

	// //copy file
	// _, err = io.Copy(osFile, rFile)
	// if err != nil {
	// 	RespondWithErr(w, 500, "error on copying file")
	// 	return
	// }

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
