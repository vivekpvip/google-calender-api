package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

func getClient(config *oauth2.Config) *http.Client {
	//local server for OAuth2 redirect
	ln, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		log.Fatalf("Unable to start local server: %v", err)
	}
	defer ln.Close()

	// Generates OAuth2 URL
	state := "state-token"
	url := config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	fmt.Printf("Open the following URL in the browser:\n%v\n", url)

	// Channel to capture the authorization code
	codeCh := make(chan string)
	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("state") != state {
				http.Error(w, "State does not match", http.StatusBadRequest)
				return
			}
			code := r.URL.Query().Get("code")
			fmt.Fprint(w, "Authorization successful! You can close this tab.")
			codeCh <- code
		})
		http.Serve(ln, nil)
	}()

	code := <-codeCh

	// Exchange code for token
	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		log.Fatalf("Unable to exchange code for token: %v", err)
	}

	saveToken(token)
	return config.Client(context.Background(), token)
}

func tokenCacheFile() string {
	usr, _ := user.Current()
	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir, "calendar-token.json")
}

// Saves the token to a secure location
func saveToken(token *oauth2.Token) {
	file := tokenCacheFile()
	f, err := os.Create(file)
	if err != nil {
		log.Fatalf("Unable to cache OAuth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
	fmt.Printf("Token saved to %s\n", file)
}

func createEvent(srv *calendar.Service) string {
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
	return createdEvent.Id
}

func updateEvent(srv *calendar.Service, eventId string) {
	event, err := srv.Events.Get("primary", eventId).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve event: %v", err)
	}

	// Modifications in event fields
	event.Summary = "Updated Test Event"
	event.Location = "Updated Location"
	event.Description = "Updated Description"
	event.Start.DateTime = time.Now().Add(48 * time.Hour).Format(time.RFC3339)
	event.End.DateTime = time.Now().Add(49 * time.Hour).Format(time.RFC3339)

	updatedEvent, err := srv.Events.Update("primary", event.Id, event).Do()
	if err != nil {
		log.Fatalf("Unable to update event: %v", err)
	}
	fmt.Printf("Event updated: %s\n", updatedEvent.HtmlLink)
}

func deleteEvent(srv *calendar.Service, eventId string) {
	err := srv.Events.Delete("primary", eventId).Do()
	if err != nil {
		log.Fatalf("Unable to delete event: %v", err)
	}
	fmt.Println("Event deleted successfully.")
}

func main() {
	ctx := context.Background()

	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read credentials.json: %v", err)
	}

	// Set up OAuth2 config
	config, err := google.ConfigFromJSON(b, calendar.CalendarScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	client := getClient(config)
	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
	}

	// Create, update, and delete event for demonstration
	eventId := createEvent(srv)
	updateEvent(srv, eventId)
	deleteEvent(srv, eventId)
}
