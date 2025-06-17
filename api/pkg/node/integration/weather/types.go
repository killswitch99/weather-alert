package weather

import (
	"encoding/json"
)

// WeatherOption represents a location for weather data
type WeatherOption struct {
	City string  `json:"city"`
	Lat  float64 `json:"lat"`
	Lon  float64 `json:"lon"`
}

// IntegrationNodeMeta holds configuration for weather integration nodes
type IntegrationNodeMeta struct {
	APIEndpoint string         `json:"apiEndpoint"`
	Options     []WeatherOption `json:"options"`
}

// ParseMetadata converts a generic metadata map to IntegrationNodeMeta
func ParseMetadata(meta map[string]any) (IntegrationNodeMeta, error) {
	var im IntegrationNodeMeta
	b, err := json.Marshal(meta)
	if err != nil {
		return im, err
	}
	err = json.Unmarshal(b, &im)
	return im, err
}
