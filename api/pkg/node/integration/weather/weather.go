package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// WeatherData represents the parsed weather API response
type WeatherData struct {
	Temperature float64 `json:"temperature"`
	Location    string  `json:"location"`
	RawResponse map[string]any `json:"rawResponse"`
}

// Client is a weather API client
type Client struct {
	httpClient *http.Client
	timeout    time.Duration
}

// NewClient creates a new weather API client
func NewClient(timeout time.Duration) *Client {
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	
	return &Client{
		httpClient: &http.Client{},
		timeout:    timeout,
	}
}

// GetWeather fetches weather data for the specified location
func (c *Client) GetWeather(ctx context.Context, endpoint string, lat, lon float64, cityName string) (*WeatherData, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	
	// Format URL with coordinates
	url := strings.ReplaceAll(endpoint, "{lat}", fmt.Sprintf("%f", lat))
	url = strings.ReplaceAll(url, "{lon}", fmt.Sprintf("%f", lon))
	
	// Create and execute request
	req, err := http.NewRequestWithContext(ctxWithTimeout, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call weather API: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("weather API returned status %d", resp.StatusCode)
	}
	
	// Parse response
	var weatherData map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&weatherData); err != nil {
		return nil, fmt.Errorf("failed to parse weather API response: %w", err)
	}
	
	currentWeather, ok := weatherData["current_weather"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid weather API response format")
	}
	
	temperature, ok := currentWeather["temperature"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid temperature value in API response")
	}
	
	return &WeatherData{
		Temperature: temperature,
		Location:    cityName,
		RawResponse: weatherData,
	}, nil
}
