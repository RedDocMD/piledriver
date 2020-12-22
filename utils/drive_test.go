package utils

import (
	"testing"

	"google.golang.org/api/drive/v3"
)

const backupDir = "backup"

func getBackupDirID(service *drive.Service, t *testing.T) string {
	list, err := service.Files.List().Fields("files(name, id, mimeType)").Do()
	if err != nil {
		t.Fatal("error while retrieving", backupDir, err)
	}
	for _, f := range list.Files {
		if f.Name == backupDir && f.MimeType == "application/vnd.google-apps.folder" {
			return f.Id
		}
	}
	// So folder has not been found
	return createBackupDir(service, t)
}

func createBackupDir(service *drive.Service, t *testing.T) string {
	file, err := CreateFolder(service, backupDir)
	if err != nil {
		t.Fatal("error while creating", backupDir, err)
	}
	return file.Id
}
