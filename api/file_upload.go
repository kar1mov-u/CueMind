package api

import (
	"fmt"
	"net/http"
	"path/filepath"
)

func (cfg *Config) UploadFile(w http.ResponseWriter, r *http.Request) {
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
