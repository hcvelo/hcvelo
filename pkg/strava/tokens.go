package strava

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int    `json:"expires_at"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresAt    int    `json:"expires_at"`
	ExpiresIn    int    `json:"expires_in"`
}

type RefreshTokensInput struct {
	ClientID     string
	ClientSecret string
	RefreshToken string
	Client       *http.Client
}

func RefreshTokens(input RefreshTokensInput) (Tokens, error) {
	log.Println("refreshing tokens")

	query := url.Values{
		"client_id":     {input.ClientID},
		"client_secret": {input.ClientSecret},
		"grant_type":    {"refresh_token"},
		"refresh_token": {input.RefreshToken},
	}
	request := http.Request{
		Method: http.MethodPost,
		URL: &url.URL{
			Path:   "/api/v3/oauth/token",
			Scheme: "https",
			Host:   "www.strava.com",
		},
		Header: http.Header{
			"Content-Type": {"application/x-www-form-urlencoded"},
		},
	}
	request.Body = io.NopCloser(strings.NewReader(query.Encode()))
	res, err := input.Client.Do(&request)
	if err != nil {
		return Tokens{}, fmt.Errorf("failed to make request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return Tokens{}, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return Tokens{}, fmt.Errorf("failed to read response body: %w", err)
	}

	var tokenResponse TokenResponse
	err = json.Unmarshal(body, &tokenResponse)
	if err != nil {
		return Tokens{}, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return Tokens{
		AccessToken:  tokenResponse.AccessToken,
		RefreshToken: tokenResponse.RefreshToken,
		ExpiresAt:    tokenResponse.ExpiresAt,
	}, nil
}
