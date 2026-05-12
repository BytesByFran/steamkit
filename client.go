package steamkit

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const (
	defaultSearchURL  = "https://store.steampowered.com/api/storesearch/?term=%v&cc=es&l=spanish"
	defaultDetailsURL = "https://store.steampowered.com/api/appdetails?appids=%d&cc=es&l=spanish"
)

type Price struct {
	Currency        string `json:"currency"`
	Initial         uint32 `json:"initial"`
	Final           uint32 `json:"final"`
	DiscountPercent uint8  `json:"discount_percent"`
	FinalFormatted  string `json:"final_formatted"`
}

type Game struct {
	ID           uint32 `json:"id"`
	Name         string `json:"name"`
	Metascore    uint32 `json:"metascore"`
	TinyImageURL string `json:"tiny_image"`
	Price        Price  `json:"price"`
}

type MetascoreDetails struct {
	Metascore uint32 `json:"score"`
	Url       string `json:"url"`
}

type GameDetails struct {
	ID        uint32           `json:"id"`
	Name      string           `json:"name"`
	Metascore MetascoreDetails `json:"metascore"`
	Price     Price            `json:"price"`
}

type searchResponse struct {
	Total int    `json:"total"`
	Items []Game `json:"items"`
}

type detailsResponse struct {
	Success bool        `json:"success"`
	Data    GameDetails `json:"data"`
}

type Client struct {
	httpClient *http.Client
	searchURL  string
	detailsURL string
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{},
		searchURL:  defaultSearchURL,
		detailsURL: defaultDetailsURL,
	}
}

func (c *Client) Search(name string) ([]Game, error) {
	if name == "" {
		return nil, fmt.Errorf("search term cannot be empty")
	}

	resp, err := c.httpClient.Get(fmt.Sprintf(c.searchURL, url.QueryEscape(name)))
	if err != nil {
		return nil, fmt.Errorf("error making search request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search request failed with status %d", resp.StatusCode)
	}

	var result searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding search response: %w", err)
	}

	return result.Items, nil
}

func (c *Client) GetGameDetails(id uint32) (GameDetails, error) {
	resp, err := c.httpClient.Get(fmt.Sprintf(c.detailsURL, id))
	var details GameDetails
	if err != nil {
		return details, fmt.Errorf("error making details request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return details, fmt.Errorf("details request failed with status %d", resp.StatusCode)
	}

	var result map[string]detailsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return details, fmt.Errorf("error decoding details response: %w", err)
	}

	entry, ok := result[fmt.Sprintf("%d", id)]
	if !ok || !entry.Success {
		return details, fmt.Errorf("game with ID %d not found", id)
	}

	return entry.Data, nil
}
