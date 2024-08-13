package couchdbUtils

import (
	"os"
	"strconv"
	"time"
)

// FileReq request struct
type FileReq struct {
	FileName string `json:"filename"`
	URI      string `json:"data_uri" quad:"data_uri"`
	FileType string `json:"filetype" quad:"filetype,opt"`
}

// File struct saved in db
type File struct {
	OrginalFileName string `json:"original_filename" quad:"original_filename"`
	FileName        string `json:"filename" quad:"filename"`
	FileType        string `json:"filetype" quad:"filetype,opt"`
}

func uploadFiles(files []FileReq) ([]File, error) {
	methodName := "uploadFiles"
	printStart(methodName)
	var filesIDList []File

	for _, upload := range files {

		URI := []byte(upload.URI)
		// filename := file.Filename
		newName, _ := CreateID()

		today := time.Now()
		month := (today.Month()).String()
		year := strconv.Itoa(today.Year())
		directory := "./uploads/" + year + "/" + month

		_ = os.MkdirAll(directory, 0777)
		path := directory + "/" + newName
		file, _ := os.Create(path)
		defer file.Close()

		_, err := file.Write(URI)
		if err != nil {
			return nil, err
		}

		idFile := year + "/" + month + "/" + newName

		filesIDList = append(filesIDList, File{
			FileName: idFile,
			FileType: upload.FileType})
	}
	return filesIDList, nil
}
