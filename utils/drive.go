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

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

const clientId = "706170668855-5j1vgust696v8cuj1ei8fs0r12vruo1r.apps.googleusercontent.com"

// Yeah its not really a secret ;)
const clientSecret = "RYnJ8ATUBnY9qI9WrnRMw4o1"

func Authorize() {
	// ctx := context.Background()
	redirectPath := "http://127.0.0.1"
	redirectPort := 4000

	conf := &oauth2.Config{
		ClientID:     clientId,
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
	wg.Add(1)
	go codeListener(redirectPort, &wg)
	wg.Wait()
}

func codeListener(port int, wg *sync.WaitGroup) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		queries := r.URL.Query()
		code, ok := queries["code"]
		if ok {
			fmt.Println("Code", code[0])
			w.Write([]byte("You can now go back to your application!"))
		} else {
			err, ok := queries["error"]
			if ok {
				fmt.Println("Error:", err[0])
				w.Write([]byte("Failed to authorize!"))
			}
		}

		wg.Done()
	})
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

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
	clientConfig, err := google.ConfigFromJSON([]byte(clientSecret), drive.DriveFileScope)
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
