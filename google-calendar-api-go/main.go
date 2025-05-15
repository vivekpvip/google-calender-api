package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// Checking that a token file already exists
	tokenFile := "token.json"
	tok, err := tokenFromFile(tokenFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokenFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then return the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	url := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Visit the following URL to obtain a code: \n%v\n", url)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.Background(), code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
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

// Saves a token to a file.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.Create(path)
	if err != nil {
		log.Fatalf("Unable to cache OAuth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func main() {
	ctx := context.Background()

	
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// Configure OAuth2
	config, err := google.ConfigFromJSON(b, calendar.CalendarScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	client := getClient(config)

	// Create Calendar Service
	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
	}

	// Creating an event
	event := &calendar.Event{
		Summary:     "Test Event",
		Location:    "Online",
		Description: "A test event created using the Google Calendar API",
		Start: &calendar.EventDateTime{
			DateTime: time.Now().Add(24 * time.Hour).Format(time.RFC3339),
			TimeZone: "Asia/Kolkata",
		},
		End: &calendar.EventDateTime{
			DateTime: time.Now().Add(25 * time.Hour).Format(time.RFC3339),
			TimeZone: "Asia/Kolkata",
		},
	}

	createdEvent, err := srv.Events.Insert("primary", event).Do()
	if err != nil {
		log.Fatalf("Unable to create event: %v", err)
	}
	fmt.Printf("Event created: %s\n", createdEvent.HtmlLink)

	// Update the event
	createdEvent.Summary = "Updated Test Event"
	updatedEvent, err := srv.Events.Update("primary", createdEvent.Id, createdEvent).Do()
	if err != nil {
		log.Fatalf("Unable to update event: %v", err)
	}
	fmt.Printf("Event updated: %s\n", updatedEvent.HtmlLink)

	// Get the event
	retrievedEvent, err := srv.Events.Get("primary", updatedEvent.Id).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve event: %v", err)
	}
	fmt.Printf("Retrieved event: %s\n", retrievedEvent.Summary)

	// Delete the event
	err = srv.Events.Delete("primary", retrievedEvent.Id).Do()
	if err != nil {
		log.Fatalf("Unable to delete event: %v", err)
	}
	fmt.Println("Event deleted successfully.")
}

