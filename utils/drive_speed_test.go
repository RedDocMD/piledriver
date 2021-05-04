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
	id, err := CreateFolder(service, "piledriver")
	if err != nil {
		log.Fatalln(err)
	}
	_, err = CreateFile(service, "test_data/speed", id)
	if err != nil {
		log.Fatalln(err)
	}
}

func createService() {
	homedir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to find homedir: %s\n", err)
	}
	homedirParts := afs.SplitPathPlatform(homedir)
	tokenPathParts := append(homedirParts, []string{".config", ".piledriver.token"}...)
	tokenPath := afs.JoinPathPlatform(tokenPathParts, true)
	service = GetDriveService(tokenPath)
}

func BenchmarkListSpeed(b *testing.B) {
	createService()
	createTestFile()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := QueryFileID(service, "piledriver/speed")
		if err != nil {
			b.Error(err)
		}
	}
}
