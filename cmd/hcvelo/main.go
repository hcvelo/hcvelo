package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/hcvelo/hcvelo/pkg/strava/events"
	"github.com/hcvelo/hcvelo/pkg/strava/tokens"
)

type ApplicationSettings struct {
	// ClientID is the client ID of the application
	ClientID string `json:"client_id"`

	// ClientSecret is the client secret of the application
	ClientSecret string `json:"client_secret"`

	// RedirectURL is the URL to redirect to after authorization
	RedirectURL string `json:"redirect_url"`

	// Scopes is a list of scopes to request
	Scopes []string `json:"scopes"`

	// AuthURL is the URL to redirect to for authorization
	AuthURL string `json:"auth_url"`

	// TokenURL is the URL to request tokens from
	TokenURL string `json:"token_url"`

	// LocalStore is the location to store application data
	LocalStore string `json:"local_store"`

	// ClubID is the ID of the club to get events for
	ClubID string `json:"club_id"`
}

func main() {
	// load application settings
	appSettings, err := loadApplicationSettings()
	if err != nil {
		fmt.Printf("failed to load application settings: %v\n", err)
		return

	}
	client := http.DefaultClient

	// load tokens from file
	appTokens, err := tokens.ReadTokensFromFile(appSettings.LocalStore)
	if err != nil {
		fmt.Printf("failed to read tokens from file: %v\n", err)
		return
	}

	expiryTime := time.Unix(int64(appTokens.ExpiresAt), 0)
	if expiryTime.Before(time.Now().Add(10 * time.Minute)) {
		// get new tokens from refresh token
		input := tokens.RefreshTokensInput{
			ClientID:     appSettings.ClientID,
			ClientSecret: appSettings.ClientSecret,
			RefreshToken: appTokens.RefreshToken,
			Client:       client,
		}
		appTokens, err = tokens.RefreshTokens(input)
		if err != nil {
			fmt.Printf("failed to refresh tokens: %v\n", err)
			return
		}
	}

	// get club activities
	getActivitiesInput := events.GetActivitiesInput{
		ClubID:      appSettings.ClubID,
		AccessToken: appTokens.AccessToken,
		Client:      client,
	}

	upcoming, err := events.GetClubActivities(getActivitiesInput)
	if err != nil {
		fmt.Printf("failed to get upcoming Strava events: %v\n", err)
	}

	// print upcoming events
	jsonUpcoming, err := json.MarshalIndent(upcoming, "", "  ")
	if err != nil {
		fmt.Printf("failed to marshal upcoming events: %v\n", err)
		return
	}

	fmt.Println(string(jsonUpcoming))

	// store tokens to file
	err = tokens.StoreTokensToFile(appSettings.LocalStore, appTokens)
	if err != nil {
		fmt.Printf("failed to store tokens to file: %v\n", err)
		return
	}
}

func loadApplicationSettings() (ApplicationSettings, error) {
	var settings ApplicationSettings

	configPath := "config.json"
	file, err := os.Open(configPath)
	if err != nil {
		return settings, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&settings)
	if err != nil {
		return settings, fmt.Errorf("failed to read settings from config file: %w", err)
	}

	return settings, nil
}
