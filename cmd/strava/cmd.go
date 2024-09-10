package strava

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "strava",
	Short: "get hcvelo information from strava",
	Long:  "get hcvelo information from strava",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("hello from strava")
	},
}

var Foo = &cobra.Command{
	Use:   "foo",
	Short: "get foo information from strava",
	Long:  "get foo information from strava",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("foo")
		wibble()
	},
}

var (
	configPath string
	tokensPath string
)

func init() {
	Cmd.Flags().StringVarP(&configPath, "config", "c", "./config.json", "path to strava config file")
	Cmd.Flags().StringVarP(&tokensPath, "tokens", "t", "./data/tokens.json", "path to strava tokwns file")
	Cmd.AddCommand(Foo)
}

type StravaSettings struct {
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

func wibble() {
	// load application settings
	appSettings, err := loadStravaSettings()
	if err != nil {
		fmt.Printf("failed to load application settings: %v\n", err)
		return

	}
	client := http.DefaultClient

	// load tokens from file
	appTokens, err := ReadTokensFromFile(appSettings.LocalStore)
	if err != nil {
		fmt.Printf("failed to read tokens from file: %v\n", err)
		return
	}

	expiryTime := time.Unix(int64(appTokens.ExpiresAt), 0)
	if expiryTime.Before(time.Now().Add(10 * time.Minute)) {
		// get new tokens from refresh token
		input := RefreshTokensInput{
			ClientID:     appSettings.ClientID,
			ClientSecret: appSettings.ClientSecret,
			RefreshToken: appTokens.RefreshToken,
			Client:       client,
		}
		appTokens, err = RefreshTokens(input)
		if err != nil {
			fmt.Printf("failed to refresh tokens: %v\n", err)
			return
		}
	}

	// get club activities
	getActivitiesInput := GetActivitiesInput{
		ClubID:      appSettings.ClubID,
		AccessToken: appTokens.AccessToken,
		Client:      client,
	}

	upcoming, err := GetClubActivities(getActivitiesInput)
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
	err = StoreTokensToFile(appSettings.LocalStore, appTokens)
	if err != nil {
		fmt.Printf("failed to store tokens to file: %v\n", err)
		return
	}
}

func loadStravaSettings() (StravaSettings, error) {
	var settings StravaSettings

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
