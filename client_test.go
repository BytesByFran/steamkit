package steamkit

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestSearchClient(t *testing.T, response any) (*Client, func()) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(response)
	}))
	c := NewClient()
	c.searchURL = server.URL + "?term=%v"
	return c, server.Close
}

func newTestDetailsClient(t *testing.T, response any) (*Client, func()) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(response)
	}))
	c := NewClient()
	c.detailsURL = server.URL + "?appids=%d"
	return c, server.Close
}

func TestSearch_ReturnsGames(t *testing.T) {
	expected := []Game{
		{ID: 70, Name: "Half-Life", Metascore: 96, Price: Price{Currency: "EUR", Initial: 999, Final: 799}},
		{ID: 730, Name: "Counter-Strike 2", Price: Price{Currency: "EUR", Initial: 0, Final: 0}},
	}
	c, cleanup := newTestSearchClient(t, searchResponse{Total: 2, Items: expected})
	defer cleanup()

	games, err := c.Search("half-life")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(games) != len(expected) {
		t.Fatalf("expected %d games, got %d", len(expected), len(games))
	}
	for i, g := range games {
		if g.ID != expected[i].ID {
			t.Errorf("game[%d]: expected ID %d, got %d", i, expected[i].ID, g.ID)
		}
		if g.Name != expected[i].Name {
			t.Errorf("game[%d]: expected name %q, got %q", i, expected[i].Name, g.Name)
		}
	}
}

func TestSearch_EmptyResults(t *testing.T) {
	c, cleanup := newTestSearchClient(t, searchResponse{Total: 0, Items: []Game{}})
	defer cleanup()

	games, err := c.Search("xyznonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(games) != 0 {
		t.Errorf("expected 0 games, got %d", len(games))
	}
}

func TestSearch_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not valid json"))
	}))
	defer server.Close()

	c := NewClient()
	c.searchURL = server.URL + "?term=%v"

	_, err := c.Search("test")
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestSearch_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	c := NewClient()
	c.searchURL = server.URL + "?term=%v"

	_, err := c.Search("test")
	if err == nil {
		t.Fatal("expected error for server error response, got nil")
	}
}

func TestGetGameDetails_ReturnsGame(t *testing.T) {
	const appID uint32 = 1245620
	c, cleanup := newTestDetailsClient(t, map[string]detailsResponse{
		"1245620": {Success: true, Data: GameDetails{ID: appID, Name: "ELDEN RING", Metascore: 96}},
	})
	defer cleanup()

	game, err := c.GetGameDetails(appID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if game.ID != appID {
		t.Errorf("expected ID %d, got %d", appID, game.ID)
	}
	if game.Name != "ELDEN RING" {
		t.Errorf("expected name %q, got %q", "ELDEN RING", game.Name)
	}
	if game.Metascore != 96 {
		t.Errorf("expected metascore 96, got %d", game.Metascore)
	}
}

func TestGetGameDetails_SuccessFalse(t *testing.T) {
	const appID uint32 = 9999
	c, cleanup := newTestDetailsClient(t, map[string]detailsResponse{
		"9999": {Success: false},
	})
	defer cleanup()

	_, err := c.GetGameDetails(appID)
	if err == nil {
		t.Fatal("expected error when success is false, got nil")
	}
}

func TestGetGameDetails_IDNotInResponse(t *testing.T) {
	const appID uint32 = 1245620
	c, cleanup := newTestDetailsClient(t, map[string]detailsResponse{
		"9999": {Success: true, Data: GameDetails{ID: 9999, Name: "Other Game"}},
	})
	defer cleanup()

	_, err := c.GetGameDetails(appID)
	if err == nil {
		t.Fatal("expected error when ID not in response, got nil")
	}
}

func TestGetGameDetails_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not valid json"))
	}))
	defer server.Close()

	c := NewClient()
	c.detailsURL = server.URL + "?appids=%d"

	_, err := c.GetGameDetails(1245620)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestGetGameDetails_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	c := NewClient()
	c.detailsURL = server.URL + "?appids=%d"

	_, err := c.GetGameDetails(1245620)
	if err == nil {
		t.Fatal("expected error for server error response, got nil")
	}
}

func TestGetGameDetails_NetworkError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	c := NewClient()
	c.detailsURL = server.URL + "?appids=%d"
	server.Close()

	_, err := c.GetGameDetails(1245620)
	if err == nil {
		t.Fatal("expected error for network error, got nil")
	}
}
