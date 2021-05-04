package utils

import (
	"log"
	"os"
	"testing"

	"github.com/RedDocMD/piledriver/afs"
	"google.golang.org/api/drive/v3"
)

var service *drive.Service

func createTestFile() {
	outChan := make(chan PathID, 10)
	id, err := CreateFolder(service, "piledriver", outChan)
	if err != nil {
		log.Fatalln(err)
	}
	_, err = CreateFile(service, "test_data/speed", outChan, id)
	if err != nil {
		log.Fatalln(err)
	}
}

func TestMain(t *testing.M) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to find homedir: %s\n", err)
	}
	homedirParts := afs.SplitPathPlatform(homedir)
	tokenPathParts := append(homedirParts, []string{".config", ".piledriver.token"}...)
	tokenPath := afs.JoinPathPlatform(tokenPathParts, true)
	log.Printf("Using %s as token path", tokenPath)
	service = GetDriveService(tokenPath)
	createTestFile()
	os.Exit(t.Run())
}

func BenchmarkListSpeed(t *testing.B) {
	for i := 0; i < t.N; i++ {
		_, err := QueryFileID(service, "piledriver/speed")
		if err != nil {
			t.Error(err)
		}
	}
}
