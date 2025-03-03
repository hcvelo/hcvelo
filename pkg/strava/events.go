package strava

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"time"
)

type SimpleClubEvent struct {
	ID                 int      `json:"id"`
	Title              string   `json:"title"`
	Description        string   `json:"description"`
	Address            string   `json:"address"`
	Date               string   `json:"date"`
	UpcomingOccurences []string `json:"upcoming_occurrences"`
}

type SimpleClubEvents []SimpleClubEvent

type GetActivitiesInput struct {
	ClubID      string
	AccessToken string
	Client      *http.Client
}

func GetClubActivities(input GetActivitiesInput) (SimpleClubEvents, error) {
	query := url.Values{
		"per_page": {"10"},
		"page":     {"1"},
	}
	request := http.Request{
		Method: http.MethodGet,
		URL: &url.URL{
			Path:   fmt.Sprintf("/api/v3/clubs/%s/group_events", input.ClubID),
			Scheme: "https",
			Host:   "www.strava.com",
		},
		Header: http.Header{
			"Authorization": {"Bearer " + input.AccessToken},
		},
	}
	request.URL.RawQuery = query.Encode()

	res, err := input.Client.Do(&request)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	events := SimpleClubEvents{}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	err = json.Unmarshal(body, &events)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	futureEvents := SimpleClubEvents{}
	futureEventIDs := make(map[int]struct{})
	fourHoursAgo := time.Now().Add(-4 * time.Hour)
	for _, event := range events {

		// bit of a hack to ignore chaingang event which is summer only
		// comment out in may!
		if event.ID == 1611651 {
			continue
		}

		if len(event.UpcomingOccurences) > 0 {
			for _, occurrence := range event.UpcomingOccurences {
				eventDate, err := time.Parse(time.RFC3339, occurrence)
				if err != nil {
					return nil, fmt.Errorf("failed to parse event date: %w", err)
				}
				if eventDate.After(fourHoursAgo) {
					event.Title = html.UnescapeString(event.Title)
					event.Description = html.UnescapeString(event.Description)
					if _, ok := futureEventIDs[event.ID]; !ok {
						futureEvents = append(futureEvents, event)
						futureEventIDs[event.ID] = struct{}{}
					}
				}
			}
		}
	}

	return futureEvents, nil
}
