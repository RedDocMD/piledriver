package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

var clientSecret = [430]byte{123, 34, 105, 110, 115, 116, 97, 108, 108, 101, 100, 34, 58, 123, 34, 99, 108, 105, 101, 110, 116, 95, 105, 100, 34, 58, 34, 55, 48, 54, 49, 55, 48, 54, 54, 56, 56, 53, 53, 45, 101, 51, 100, 108, 112, 102, 56, 111, 102, 48, 116, 103, 108, 111, 97, 118, 101, 52, 104, 118, 98, 103, 54, 116, 108, 111, 49, 55, 117, 114, 117, 116, 46, 97, 112, 112, 115, 46, 103, 111, 111, 103, 108, 101, 117, 115, 101, 114, 99, 111, 110, 116, 101, 110, 116, 46, 99, 111, 109, 34, 44, 34, 112, 114, 111, 106, 101, 99, 116, 95, 105, 100, 34, 58, 34, 112, 105, 108, 101, 100, 114, 105, 118, 101, 114, 45, 49, 54, 48, 56, 51, 49, 53, 53, 51, 50, 53, 56, 51, 34, 44, 34, 97, 117, 116, 104, 95, 117, 114, 105, 34, 58, 34, 104, 116, 116, 112, 115, 58, 47, 47, 97, 99, 99, 111, 117, 110, 116, 115, 46, 103, 111, 111, 103, 108, 101, 46, 99, 111, 109, 47, 111, 47, 111, 97, 117, 116, 104, 50, 47, 97, 117, 116, 104, 34, 44, 34, 116, 111, 107, 101, 110, 95, 117, 114, 105, 34, 58, 34, 104, 116, 116, 112, 115, 58, 47, 47, 111, 97, 117, 116, 104, 50, 46, 103, 111, 111, 103, 108, 101, 97, 112, 105, 115, 46, 99, 111, 109, 47, 116, 111, 107, 101, 110, 34, 44, 34, 97, 117, 116, 104, 95, 112, 114, 111, 118, 105, 100, 101, 114, 95, 120, 53, 48, 57, 95, 99, 101, 114, 116, 95, 117, 114, 108, 34, 58, 34, 104, 116, 116, 112, 115, 58, 47, 47, 119, 119, 119, 46, 103, 111, 111, 103, 108, 101, 97, 112, 105, 115, 46, 99, 111, 109, 47, 111, 97, 117, 116, 104, 50, 47, 118, 49, 47, 99, 101, 114, 116, 115, 34, 44, 34, 99, 108, 105, 101, 110, 116, 95, 115, 101, 99, 114, 101, 116, 34, 58, 34, 122, 115, 76, 119, 56, 89, 55, 107, 56, 109, 118, 82, 81, 117, 118, 88, 70, 109, 53, 107, 98, 78, 104, 85, 34, 44, 34, 114, 101, 100, 105, 114, 101, 99, 116, 95, 117, 114, 105, 115, 34, 58, 91, 34, 117, 114, 110, 58, 105, 101, 116, 102, 58, 119, 103, 58, 111, 97, 117, 116, 104, 58, 50, 46, 48, 58, 111, 111, 98, 34, 44, 34, 104, 116, 116, 112, 58, 47, 47, 108, 111, 99, 97, 108, 104, 111, 115, 116, 34, 93, 125, 125}

func getClient(config *oauth2.Config) (context.Context, *http.Client) {
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	ctx := context.Background()
	return ctx, config.Client(ctx, tok)
}

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Printf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

// RetrieveDriveService gets the drive service via an HTTP client
func RetrieveDriveService() *drive.Service {
	clientConfig, err := google.ConfigFromJSON(clientSecret[:], drive.DriveFileScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	context, client := getClient(clientConfig)
	service, err := drive.NewService(context, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Drive client: %v", err)
	}
	return service
}

// CreateFile creates the file in drive, with the parent directory specified by
// parentID and filename same as input file.
// It does NOT check for the validity of parentID.
// If queue is nil, the Do() will be executed in this method itself
func CreateFile(
	service *drive.Service,
	local string,
	parentID string,
	queue chan interface{}) error {

	filename := path.Base(local)
	driveFile := &drive.File{
		Name:    filename,
		Parents: []string{parentID},
	}
	localfile, err := os.Open(local)
	if err != nil {
		return err
	}
	defer localfile.Close()
	call := service.Files.Create(driveFile).Media(localfile)
	if queue != nil {
		queue <- call
	} else {
		_, err := call.Do()
		if err != nil {
			return err
		}
	}
	return nil
}

// CreateFolder creates a folder in drive, with a parent directory specified by parentID
// If no parent directories are specified, then it is not set
func CreateFolder(service *drive.Service, name string, parentID ...string) (*drive.File, error) {
	dir := &drive.File{
		Name:     name,
		MimeType: "application/vnd.google-apps.folder",
		Parents:  parentID,
	}
	return service.Files.Create(dir).Do()
}
