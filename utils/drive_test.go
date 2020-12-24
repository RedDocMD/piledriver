package utils

import (
	"sync"
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

func TestUploadWithoutChannel(t *testing.T) {
	service := RetrieveDriveService()
	backupDir := getBackupDirID(service, t)
	times := 10
	filename := "./test_data/upload5M.pdf"
	for i := 1; i <= times; i++ {
		err := CreateFile(service, filename, backupDir, nil)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestUploadWithChannel(t *testing.T) {
	service := RetrieveDriveService()
	backupDir := getBackupDirID(service, t)
	times := 10
	filename := "./test_data/upload5M.pdf"
	var wg sync.WaitGroup
	for i := 1; i <= times; i++ {
		wg.Add(1)
		go func() {
			err := CreateFile(service, filename, backupDir, nil)
			if err != nil {
				t.Error(err)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}
