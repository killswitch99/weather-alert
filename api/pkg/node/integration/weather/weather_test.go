package weather

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWeatherEmoji(t *testing.T) {
	emoji := WeatherEmoji{}
	
	testCases := []struct {
		temp     float64
		expected string
	}{
		{-10, "🥶"}, // cold
		{0, "🥶"},   // cold
		{5, "🧥"},   // cool
		{10, "🧥"},  // cool
		{15, "🙂"},  // mild
		{20, "🙂"},  // mild
		{25, "😎"},  // warm
		{30, "😎"},  // warm
		{35, "🥵"},  // very hot
		{40, "🥵"},  // very hot
	}
	
	for _, tc := range testCases {
		t.Run(string(tc.expected), func(t *testing.T) {
			result := emoji.Emoji(tc.temp)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestParseMetadata(t *testing.T) {
	testCases := []struct {
		name          string
		input         map[string]any
		expectError   bool
		expectedMeta  IntegrationNodeMeta
	}{
		{
			name: "Valid metadata",
			input: map[string]any{
				"apiEndpoint": "https://api.example.com/weather",
				"options": []any{
					map[string]any{
						"city": "New York",
						"lat":  40.7128,
						"lon":  -74.0060,
					},
					map[string]any{
						"city": "London",
						"lat":  51.5072,
						"lon":  -0.1276,
					},
				},
			},
			expectError: false,
			expectedMeta: IntegrationNodeMeta{
				APIEndpoint: "https://api.example.com/weather",
				Options: []WeatherOption{
					{
						City: "New York",
						Lat:  40.7128,
						Lon:  -74.0060,
					},
					{
						City: "London",
						Lat:  51.5072,
						Lon:  -0.1276,
					},
				},
			},
		},
		{
			name: "Empty metadata",
			input: map[string]any{
				"apiEndpoint": "",
				"options":     []any{},
			},
			expectError: false,
			expectedMeta: IntegrationNodeMeta{
				APIEndpoint: "",
				Options:     []WeatherOption{},
			},
		},
		{
			name: "Invalid options format",
			input: map[string]any{
				"apiEndpoint": "https://api.example.com/weather",
				"options":     "not-an-array",
			},
			expectError: true,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			meta, err := ParseMetadata(tc.input)
			
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedMeta.APIEndpoint, meta.APIEndpoint)
				
				// Compare each option separately for better error messages
				assert.Equal(t, len(tc.expectedMeta.Options), len(meta.Options))
				
				for i, expectedOpt := range tc.expectedMeta.Options {
					if i < len(meta.Options) {
						assert.Equal(t, expectedOpt.City, meta.Options[i].City)
						assert.Equal(t, expectedOpt.Lat, meta.Options[i].Lat)
						assert.Equal(t, expectedOpt.Lon, meta.Options[i].Lon)
					}
				}
			}
		})
	}
}

func TestWeatherOptionMarshaling(t *testing.T) {
	// Test JSON marshaling/unmarshaling
	original := WeatherOption{
		City: "Tokyo",
		Lat:  35.6762,
		Lon:  139.6503,
	}
	
	// Marshal to JSON
	jsonBytes, err := json.Marshal(original)
	assert.NoError(t, err)
	
	// Unmarshal back to struct
	var unmarshaled WeatherOption
	err = json.Unmarshal(jsonBytes, &unmarshaled)
	assert.NoError(t, err)
	
	// Compare
	assert.Equal(t, original.City, unmarshaled.City)
	assert.Equal(t, original.Lat, unmarshaled.Lat)
	assert.Equal(t, original.Lon, unmarshaled.Lon)
}
