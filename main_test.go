package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestForecastHandler tests the forecast endpoint with mocked NWS API
func TestForecastHandler(t *testing.T) {
	tests := []struct {
		name               string
		latitude           string
		longitude          string
		pointsStatusCode   int
		forecastStatusCode int
		forecastResponse   string
		expectedStatus     int
		expectedForecast   string
		expectedTemp       string
	}{
		{
			name:               "successful forecast - moderate temperature",
			latitude:           "47.6062",
			longitude:          "-122.3321",
			pointsStatusCode:   200,
			forecastStatusCode: 200,
			forecastResponse: `{
				"properties": {
					"periods": [
						{
							"shortForecast": "Partly Cloudy",
							"temperature": 65
						}
					]
				}
			}`,
			expectedStatus:   200,
			expectedForecast: "Partly Cloudy",
			expectedTemp:     "moderate",
		},
		{
			name:               "successful forecast - cold temperature",
			latitude:           "47.6062",
			longitude:          "-122.3321",
			pointsStatusCode:   200,
			forecastStatusCode: 200,
			forecastResponse: `{
				"properties": {
					"periods": [
						{
							"shortForecast": "Snow",
							"temperature": 25
						}
					]
				}
			}`,
			expectedStatus:   200,
			expectedForecast: "Snow",
			expectedTemp:     "cold",
		},
		{
			name:               "successful forecast - hot temperature",
			latitude:           "34.0522",
			longitude:          "-118.2437",
			pointsStatusCode:   200,
			forecastStatusCode: 200,
			forecastResponse: `{
				"properties": {
					"periods": [
						{
							"shortForecast": "Sunny",
							"temperature": 95
						}
					]
				}
			}`,
			expectedStatus:   200,
			expectedForecast: "Sunny",
			expectedTemp:     "hot",
		},
		{
			name:               "successful forecast - temperature at cold boundary (30)",
			latitude:           "47.6062",
			longitude:          "-122.3321",
			pointsStatusCode:   200,
			forecastStatusCode: 200,
			forecastResponse: `{
				"properties": {
					"periods": [
						{
							"shortForecast": "Freezing",
							"temperature": 30
						}
					]
				}
			}`,
			expectedStatus:   200,
			expectedForecast: "Freezing",
			expectedTemp:     "cold",
		},
		{
			name:               "successful forecast - temperature at hot boundary (80)",
			latitude:           "34.0522",
			longitude:          "-118.2437",
			pointsStatusCode:   200,
			forecastStatusCode: 200,
			forecastResponse: `{
				"properties": {
					"periods": [
						{
							"shortForecast": "Hot",
							"temperature": 80
						}
					]
				}
			}`,
			expectedStatus:   200,
			expectedForecast: "Hot",
			expectedTemp:     "hot",
		},
		{
			name:             "points API returns 404",
			latitude:         "99.9999",
			longitude:        "-999.9999",
			pointsStatusCode: 404,
			expectedStatus:   404,
		},
		{
			name:             "points API returns 500",
			latitude:         "47.6062",
			longitude:        "-122.3321",
			pointsStatusCode: 500,
			expectedStatus:   500,
		},
		{
			name:               "forecast API returns 404",
			latitude:           "47.6062",
			longitude:          "-122.3321",
			pointsStatusCode:   200,
			forecastStatusCode: 404,
			forecastResponse:   `{"status": 404, "detail": "Forecast not found"}`,
			expectedStatus:     404,
		},
		{
			name:               "forecast API returns 503",
			latitude:           "47.6062",
			longitude:          "-122.3321",
			pointsStatusCode:   200,
			forecastStatusCode: 503,
			forecastResponse:   `{"status": 503, "detail": "Service unavailable"}`,
			expectedStatus:     503,
		},
		{
			name:               "empty periods array",
			latitude:           "47.6062",
			longitude:          "-122.3321",
			pointsStatusCode:   200,
			forecastStatusCode: 200,
			forecastResponse: `{
				"properties": {
					"periods": []
				}
			}`,
			expectedStatus: 404,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock NWS API server
			mockNWS := createMockNWSServer(tt.pointsStatusCode, tt.forecastStatusCode, tt.forecastResponse)
			defer mockNWS.Close()

			// Override nwsAPIHost for testing
			originalHost := nwsAPIHost
			nwsAPIHost = mockNWS.URL
			defer func() { nwsAPIHost = originalHost }()

			// Create test request
			url := fmt.Sprintf("/forecast?latitude=%s&longitude=%s", tt.latitude, tt.longitude)
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			// Execute handler
			forecastHandler(w, req)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// For successful cases, verify the response body
			if tt.expectedStatus == 200 {
				var response ForecastOutput
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if response.Forecast != tt.expectedForecast {
					t.Errorf("expected forecast %q, got %q", tt.expectedForecast, response.Forecast)
				}

				if response.Temperature != tt.expectedTemp {
					t.Errorf("expected temperature %q, got %q", tt.expectedTemp, response.Temperature)
				}
			}
		})
	}
}

// TestForecastHandlerMissingParameters tests missing query parameters
func TestForecastHandlerMissingParameters(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		expectedStatus int
	}{
		{
			name:           "missing latitude",
			url:            "/forecast?longitude=-122.3321",
			expectedStatus: 400,
		},
		{
			name:           "missing longitude",
			url:            "/forecast?latitude=47.6062",
			expectedStatus: 400,
		},
		{
			name:           "missing both parameters",
			url:            "/forecast",
			expectedStatus: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.url, nil)
			w := httptest.NewRecorder()

			forecastHandler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

// TestForecastHandlerInvalidMethod tests non-GET methods
func TestForecastHandlerInvalidMethod(t *testing.T) {
	methods := []string{"POST", "PUT", "DELETE", "PATCH"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/forecast?latitude=47.6062&longitude=-122.3321", nil)
			w := httptest.NewRecorder()

			forecastHandler(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
			}
		})
	}
}

// TestMapTemperature tests the temperature mapping function
func TestMapTemperature(t *testing.T) {
	tests := []struct {
		temperature int
		expected    string
	}{
		{temperature: -10, expected: "cold"},
		{temperature: 0, expected: "cold"},
		{temperature: 30, expected: "cold"},
		{temperature: 31, expected: "moderate"},
		{temperature: 50, expected: "moderate"},
		{temperature: 79, expected: "moderate"},
		{temperature: 80, expected: "hot"},
		{temperature: 100, expected: "hot"},
		{temperature: 120, expected: "hot"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("temp_%d", tt.temperature), func(t *testing.T) {
			result := mapTemperature(tt.temperature)
			if result != tt.expected {
				t.Errorf("mapTemperature(%d) = %q, expected %q", tt.temperature, result, tt.expected)
			}
		})
	}
}

// createMockNWSServer creates a mock NWS API server for testing
func createMockNWSServer(pointsStatus int, forecastStatus int, forecastResp string) *httptest.Server {
	handler := http.NewServeMux()

	var server *httptest.Server

	// Mock points endpoint
	handler.HandleFunc("/points/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(pointsStatus)

		if pointsStatus == 200 {
			// Build points response with forecast URL pointing to mock server
			pointsResp := fmt.Sprintf(`{
				"properties": {
					"forecast": "%s/forecast-url"
				}
			}`, server.URL)
			w.Write([]byte(pointsResp))
		} else {
			// For error cases, return error response
			w.Write([]byte(fmt.Sprintf(`{"status": %d, "detail": "Error"}`, pointsStatus)))
		}
	})

	// Mock forecast endpoint
	handler.HandleFunc("/forecast-url", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(forecastStatus)
		w.Write([]byte(forecastResp))
	})

	server = httptest.NewServer(handler)
	return server
}
