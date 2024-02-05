// The functions in this file originated at the Google quickstart for the Go API.
// See https://developers.google.com/sheets/api/quickstart/go for details.

package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"log"
	"net/http"
	"os"
)

// GetClient Retrieve a token, saves the token, then returns the generated client.
// This function originated at the Google quickstart for the Go sheets API.
func GetClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := TokenFromFile(tokFile)

	if err != nil {
		tok = GetTokenFromWeb(config)
		SaveToken(tokFile, tok)
	}

	return config.Client(context.Background(), tok)
}

// SaveToken Saves a token to a file path.
// This function originated at the Google quickstart for the Go sheets API.
func SaveToken(path string, token *oauth2.Token) {
	log.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)

	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	err = json.NewEncoder(f).Encode(token)

	if err != nil {
		log.Fatalf("Failure encoding token: %v", err)
	}
}

// GetTokenFromWeb Request a token from the web, then returns the retrieved token.
// This function originated at the Google quickstart for the Go sheets API.
func GetTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	log.Printf("Go to the following link in your browser then type the "+
		"authorization code ('code=' contained within the generated URL): \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}

	return tok
}

// TokenFromFile Retrieves a token from a local file.
// This function originated at the Google quickstart for the Go sheets API.
func TokenFromFile(file string) (*oauth2.Token, error) {
	fileHandle, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer fileHandle.Close()

	tok := &oauth2.Token{}
	err = json.NewDecoder(fileHandle).Decode(tok)

	return tok, err
}

func AccessGoogleClient() (context.Context, *http.Client, error) {
	ctx := context.Background()
	fileHandle, err := os.ReadFile("credentials.json")

	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(
		fileHandle,
		"https://www.googleapis.com/auth/spreadsheets.readonly",
		"https://www.googleapis.com/auth/calendar.events")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	client := GetClient(config)

	return ctx, client, err
}
