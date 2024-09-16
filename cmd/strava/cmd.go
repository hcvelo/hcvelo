package strava

import (
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	lib "github.com/hcvelo/hcvelo/pkg/strava"
)

type MarkdownEvent struct {
	lib.SimpleClubEvent
	URL  string
	Date string
	Time string
}

var Cmd = &cobra.Command{
	Use:   "strava",
	Short: "get hcvelo information from strava",
	Long:  "get hcvelo information from strava",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			return
		}
	},
}

var (
	clientID     string
	clientSecret string
	clubID       string
	format       string
	outputDir    string
)

func init() {
	Cmd.PersistentFlags().StringVar(&clientID, "clientID", "", "clientID for accessing strava")
	Cmd.PersistentFlags().StringVar(&clientSecret, "clientSecret", "", "client secret for accessing strava")
	Cmd.PersistentFlags().StringVar(&clubID, "clubID", "", "club ID for accessing strava")

	events := &cobra.Command{
		Use:   "events",
		Short: "get hcvelo events information from strava",
		Long:  "get hcvelo events information from strava",
		Run: func(cmd *cobra.Command, args []string) {
			f := cmd.Flags()
			configPath, _ := f.GetString("configDir")
			getEvents(configPath)
		},
	}
	events.Flags().StringVar(&format, "format", "json", "output format, can be json or md")
	events.Flags().StringVar(&outputDir, "output", "", "if given, output will be written to the specified directory")
	Cmd.AddCommand(events)

	tokens := &cobra.Command{
		Use:   "tokens",
		Short: "get current strava tokens",
		Long:  "get current strava tokens",
		Run: func(cmd *cobra.Command, args []string) {
			f := cmd.Flags()
			configPath, _ := f.GetString("configDir")
			getTokens(configPath)

			fmt.Println(clubID)
		},
	}
	Cmd.AddCommand(tokens)
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

	// ClubID is the ID of the club to get events for
	ClubID string `json:"club_id"`
}

func getEvents(configDir string) {
	// load application settings
	appSettings, err := loadStravaSettings(configDir)
	if err != nil {
		fmt.Printf("failed to load application settings: %v\n", err)
		return

	}
	client := http.DefaultClient

	// load tokens from file
	appTokens, err := readTokensFromFile(configDir)
	if err != nil {
		fmt.Printf("failed to read tokens from file: %v\n", err)
		return
	}

	expiryTime := time.Unix(int64(appTokens.ExpiresAt), 0)
	if expiryTime.Before(time.Now().Add(10 * time.Minute)) {
		// get new tokens from refresh token
		input := lib.RefreshTokensInput{
			ClientID:     appSettings.ClientID,
			ClientSecret: appSettings.ClientSecret,
			RefreshToken: appTokens.RefreshToken,
			Client:       client,
		}
		appTokens, err = lib.RefreshTokens(input)
		if err != nil {
			fmt.Printf("failed to refresh tokens: %v\n", err)
			return
		}
	}

	// get club activities
	getActivitiesInput := lib.GetActivitiesInput{
		ClubID:      appSettings.ClubID,
		AccessToken: appTokens.AccessToken,
		Client:      client,
	}

	upcoming, err := lib.GetClubActivities(getActivitiesInput)
	if err != nil {
		fmt.Printf("failed to get upcoming Strava events: %v\n", err)
	}

	if format == "md" {
		ukLoc, err := time.LoadLocation("Europe/London")
		if err != nil {
			fmt.Printf("failed to load location: %v\n", err)
			return
		}
		for _, event := range upcoming {
			markdownEvent := MarkdownEvent{
				SimpleClubEvent: event,
				URL:             fmt.Sprintf("https://www.strava.com/clubs/%s/group_events/%d", appSettings.ClubID, event.ID),
			}
			date, err := time.Parse(time.RFC3339, event.UpcomingOccurences[0])
			if err != nil {
				fmt.Printf("failed to parse time: %v\n", err)
				return
			}
			markdownEvent.Date = date.Format("01-01-2006")
			markdownEvent.Time = date.In(ukLoc).Format("15:04")
			markdownEvent.Description = html.UnescapeString(event.Description)

			// load template from a file
			tmpl, err := template.ParseFiles("event.md")
			if err != nil {
				fmt.Printf("failed to parse template: %v\n", err)
				return
			}

			// execute template
			if outputDir != "" {
				outputPath := filepath.Join(outputDir, fmt.Sprintf("%d.md", event.ID))
				file, err := os.Create(outputPath)
				if err != nil {
					fmt.Printf("failed to create output file: %v\n", err)
					return
				}
				defer file.Close()

				err = tmpl.Execute(file, markdownEvent)
				if err != nil {
					fmt.Printf("failed to execute template: %v\n", err)
					return
				}
			} else {
				err = tmpl.Execute(os.Stdout, markdownEvent)
				if err != nil {
					fmt.Printf("failed to execute template: %v\n", err)
					return
				}
			}
		}
	} else {
		jsonUpcoming, err := json.MarshalIndent(upcoming, "", "  ")
		if err != nil {
			fmt.Printf("failed to marshal upcoming events: %v\n", err)
			return
		}
		if outputDir != "" {
			outputPath := filepath.Join(outputDir, "upcoming.json")
			file, err := os.Create(outputPath)
			if err != nil {
				fmt.Printf("failed to create output file: %v\n", err)
				return
			}
			defer file.Close()

			_, err = file.Write(jsonUpcoming)
			if err != nil {
				fmt.Printf("failed to write output file: %v\n", err)
				return
			}
		} else {
			fmt.Println(string(jsonUpcoming))
		}
	}

	// store tokens to file
	err = storeTokensToFile(configDir, appTokens)
	if err != nil {
		fmt.Printf("failed to store tokens to file: %v\n", err)
		return
	}
}

func getTokens(configDir string) {
	// load tokens from file
	appTokens, err := readTokensFromFile(configDir)
	if err != nil {
		fmt.Printf("failed to read tokens from file: %v\n", err)
		return
	}

	// print tokens
	jsonTokens, err := json.MarshalIndent(appTokens, "", "  ")
	if err != nil {
		fmt.Printf("failed to marshal tokens: %v\n", err)
		return
	}

	fmt.Println(string(jsonTokens))
}

func loadStravaSettings(configDir string) (StravaSettings, error) {
	var settings StravaSettings

	configPath := filepath.Join(configDir, "strava.json")
	file, err := os.Open(configPath)
	if err != nil {
		return settings, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&settings)
	if err != nil {
		return settings, fmt.Errorf("failed to read settings from config file: %w", err)
	}

	if clientID != "" {
		settings.ClientID = clientID
	}

	if clientSecret != "" {
		settings.ClientSecret = clientSecret
	}

	if clubID != "" {
		settings.ClubID = clubID
	}

	return settings, nil
}

func storeTokensToFile(configDir string, tokens lib.Tokens) error {
	// open file for writing
	tokensFilePath := filepath.Join(configDir, "stravaTokens.json")
	file, err := os.Create(tokensFilePath)
	if err != nil {
		return fmt.Errorf("failed to create tokens file: %w", err)
	}
	defer file.Close()

	// write tokens to file
	err = json.NewEncoder(file).Encode(tokens)
	if err != nil {
		return fmt.Errorf("failed to write tokens to file: %w", err)
	}

	return nil
}

func readTokensFromFile(configDir string) (lib.Tokens, error) {
	// open file for reading
	tokensFilePath := filepath.Join(configDir, "stravaTokens.json")
	file, err := os.Open(tokensFilePath)
	if err != nil {
		return lib.Tokens{}, fmt.Errorf("failed to open tokens file: %w", err)
	}
	defer file.Close()

	// read tokens from file
	var tokens lib.Tokens
	err = json.NewDecoder(file).Decode(&tokens)
	if err != nil {
		return lib.Tokens{}, fmt.Errorf("failed to read tokens from file: %w", err)
	}

	return tokens, nil
}
