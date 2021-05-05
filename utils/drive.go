package utils

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"path"
	"sync"

	"github.com/RedDocMD/piledriver/afs"
	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

const clientID = "706170668855-5j1vgust696v8cuj1ei8fs0r12vruo1r.apps.googleusercontent.com"

// Yeah its not really a secret ;)
const clientSecret = "RYnJ8ATUBnY9qI9WrnRMw4o1"

// The following is a hack to not have to open files
// I just don't know any better :(
const successResponse = `
<!DOCTYPE html>
<html lang="en">

<head>
    <meta charset="UTF-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.0.0-beta3/dist/css/bootstrap.min.css" rel="stylesheet"
        integrity="sha384-eOJMYsd53ii+scO/bJGFsiCZc+5NDVN2yr8+0RDqr0Ql0h+rP48ckxlpbzKgwra6" crossorigin="anonymous">
    <title>Success</title>
</head>

<body>
    <div class="container-fluid mt-3">
        <p class="display-3 text-center">Piledriver</p>
        <div class="text-center">
            <div class="fw-bold text-success fs-5">Auhentication succeeded!</div>You may now return to the application.
        </div>
    </div>
</body>

</html>
`

const failureResponse = `
<!DOCTYPE html>
<html lang="en">

<head>
    <meta charset="UTF-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.0.0-beta3/dist/css/bootstrap.min.css" rel="stylesheet"
        integrity="sha384-eOJMYsd53ii+scO/bJGFsiCZc+5NDVN2yr8+0RDqr0Ql0h+rP48ckxlpbzKgwra6" crossorigin="anonymous">
    <title>Failure</title>
</head>

<body>
    <div class="container-fluid mt-3">
        <p class="display-3 text-center">Piledriver</p>
        <div class="text-center">
            <div class="fw-bold text-danger fs-5">Authentication failed!</div>Please return to the application for more information
        </div>
    </div>
</body>

</html>
`

// GetDriveService reads the token from the file denoted by tokenLocation
// and then returns the Google Drive service. If it cannot find the token file,
// it errors out and stops the program.
func GetDriveService(tokenLocation string) *drive.Service {
	tok, err := tokenFromFile(tokenLocation)
	if err != nil {
		log.Fatalf("Piledriver has not been authenticated: please run \"piledriver auth\"\n")
	}

	ctx := context.Background()

	redirectPath := "http://127.0.0.1"
	redirectPort := 4598

	conf := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{drive.DriveFileScope},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
		},
		RedirectURL: fmt.Sprintf("%s:%d", redirectPath, redirectPort),
	}

	driveService, err := drive.NewService(ctx, option.WithTokenSource(conf.TokenSource(ctx, tok)))
	if err != nil {
		log.Fatalf("Failed to create drive client: %s\n", err)
	}
	return driveService
}

// AuthorizeApp triggers the OAuth 2.0 web authentication
// for Piledriver
func AuthorizeApp(tokenLocation string) {
	ctx := context.Background()

	redirectPath := "http://127.0.0.1"
	redirectPort := 4598

	conf := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{drive.DriveFileScope},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
		},
		RedirectURL: fmt.Sprintf("%s:%d", redirectPath, redirectPort),
	}

	randLim := big.NewInt(1)
	randLim.Lsh(randLim, 200)
	csrfVal, err := rand.Int(rand.Reader, randLim)
	if err != nil {
		log.Fatalf("Error while generating CSRF token: %s\n", err)
	}

	url := conf.AuthCodeURL(fmt.Sprint(csrfVal), oauth2.AccessTypeOffline)
	fmt.Printf("Open the following URL in your browser:\n%s\n\n", url)

	var wg sync.WaitGroup
	authChan := make(chan map[string]string)
	go func() {
		wg.Add(1)
		handleOAuthRedirect(redirectPort, &wg, &authChan)
	}()
	wg.Wait()

	ans := <-authChan

	var code string
	if val, ok := ans["code"]; ok {
		code = val
	} else if val, ok := ans["error"]; ok {
		fmt.Println("Failed to authenticate!")
		fmt.Println("Reason:", val)
		os.Exit(1)
	} else {
		fmt.Println("Something unexpected happened while authenticating")
		os.Exit(1)
	}

	tok, err := conf.Exchange(ctx, code)
	if err != nil {
		log.Fatalf("Failed to get token\n")
	}
	saveToken(tokenLocation, tok)
}

func handleOAuthRedirect(port int, wg *sync.WaitGroup, ans *chan map[string]string) {
	successBytes := []byte(successResponse)
	failureBytes := []byte(failureResponse)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		queries := r.URL.Query()
		resp := make(map[string]string)

		if code, ok := queries["code"]; ok {
			resp["code"] = code[0]
			w.Write(successBytes)
		} else if err, ok := queries["error"]; ok {
			resp["error"] = err[0]
			w.Write(failureBytes)
		} else {
			w.Write([]byte("Unknown problem occured"))
		}

		*ans <- resp
		wg.Done()
	})
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
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

type PathID struct {
	path string
	id   string
}

// CreateFile creates the file in drive, with the parent directory specified by
// parentID and filename same as input file.
// It does NOT check for the validity of parentID.
// If queue is nil, the Do() will be executed in this method itself
func CreateFile(service *drive.Service, local string, parentID string) (string, error) {
	filename := path.Base(local)
	driveFile := &drive.File{
		Name:    filename,
		Parents: []string{parentID},
	}
	localfile, err := os.Open(local)
	if err != nil {
		return "", err
	}
	defer localfile.Close()
	driveFile, err = service.Files.Create(driveFile).Media(localfile).Do()
	return driveFile.Id, err
}

// CreateFolder creates a folder in drive, with a parent directory specified by parentID
// If no parent directories are specified, then it is not set
func CreateFolder(service *drive.Service, remote string, parentID ...string) (string, error) {
	parts := afs.SplitPathPlatform(remote)
	dir := &drive.File{
		Name:     parts[len(parts)-1],
		MimeType: "application/vnd.google-apps.folder",
		Parents:  parentID,
	}
	file, err := service.Files.Create(dir).Do()
	return file.Id, err
}

// DeleteFileOrFolder deletes the file (or folder) in the drive with the givwn ID
func DeleteFileOrFolder(service *drive.Service, id string) error {
	return service.Files.Delete(id).Do()
}

func QueryFileID(service *drive.Service, local string) (string, error) {
	parts := afs.SplitPathPlatform(local)
	name := parts[len(parts)-1]

	nextPageToken := ""

	for {
		listCall := service.Files.List().
			Fields("nextPageToken, files(name, id)")
		if nextPageToken != "" {
			listCall = listCall.PageToken(nextPageToken)
		}
		list, err := listCall.Do()
		if err != nil {
			return "", err
		}
		for _, file := range list.Files {
			if file.Name == name {
				return file.Id, nil
			}
		}
		if nextPageToken == "" {
			break
		}
	}
	return "", fmt.Errorf("Didn't find %s in you Drive", local)
}

func QueryAllContents(service *drive.Service) ([]*drive.File, error) {
	nextPageToken := ""
	var nonTrashFiles []*drive.File

	for {
		listCall := service.Files.List().
			Fields("nextPageToken, files(name, id, trashed, parents, mimeType)").
			PageToken(nextPageToken)
		list, err := listCall.Do()
		if err != nil {
			return nil, err
		}
		for _, file := range list.Files {
			if !file.Trashed {
				nonTrashFiles = append(nonTrashFiles, file)
			}
		}
		if nextPageToken == "" {
			break
		}
	}
	return nonTrashFiles, nil
}
