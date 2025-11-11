package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

const (
	userAgent = "(murphybytes.com murphybytes@gmail.com)"
)

var (
	// nwsAPIHost can be overridden for testing
	nwsAPIHost = "https://api.weather.gov"
)

// PointResponse represents the NWS points API response
type PointResponse struct {
	Properties struct {
		Forecast string `json:"forecast"`
	} `json:"properties"`
}

// ForecastResponse represents the NWS forecast API response
type ForecastResponse struct {
	Properties struct {
		Periods []struct {
			ShortForecast string `json:"shortForecast"`
			Temperature   int    `json:"temperature"`
		} `json:"periods"`
	} `json:"properties"`
}

// ForecastOutput represents our API response
type ForecastOutput struct {
	Forecast    string `json:"forecast"`
	Temperature string `json:"temperature"`
}

func main() {
	http.HandleFunc("/forecast", forecastHandler)

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func forecastHandler(w http.ResponseWriter, r *http.Request) {
	// Only accept GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get query parameters
	lat := r.URL.Query().Get("latitude")
	lon := r.URL.Query().Get("longitude")

	if lat == "" || lon == "" {
		http.Error(w, "Missing latitude or longitude parameter", http.StatusBadRequest)
		return
	}

	// Step 1: Call the points endpoint
	pointsURL := fmt.Sprintf("%s/points/%s,%s", nwsAPIHost, lat, lon)
	pointResp, statusCode, err := makeNWSRequest(pointsURL)
	if err != nil {
		http.Error(w, err.Error(), statusCode)
		return
	}

	var pointData PointResponse
	if err := json.Unmarshal(pointResp, &pointData); err != nil {
		http.Error(w, "Failed to parse points response", http.StatusInternalServerError)
		return
	}

	// Step 2: Get the forecast URL from the response
	forecastURL := pointData.Properties.Forecast
	if forecastURL == "" {
		http.Error(w, "Forecast URL not found", http.StatusNotFound)
		return
	}

	// Step 3: Call the forecast endpoint
	forecastResp, statusCode, err := makeNWSRequest(forecastURL)
	if err != nil {
		http.Error(w, err.Error(), statusCode)
		return
	}

	var forecastData ForecastResponse
	if err := json.Unmarshal(forecastResp, &forecastData); err != nil {
		http.Error(w, "Failed to parse forecast response", http.StatusInternalServerError)
		return
	}

	// Step 4: Extract the first period's data
	if len(forecastData.Properties.Periods) == 0 {
		http.Error(w, "No forecast periods found", http.StatusNotFound)
		return
	}

	firstPeriod := forecastData.Properties.Periods[0]

	// Step 5: Map temperature to cold/moderate/hot
	tempCategory := mapTemperature(firstPeriod.Temperature)

	// Step 6: Build and return the response
	output := ForecastOutput{
		Forecast:    firstPeriod.ShortForecast,
		Temperature: tempCategory,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(output)
}

// makeNWSRequest makes an HTTP request to the NWS API with the required User-Agent header
func makeNWSRequest(url string) ([]byte, int, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("User-Agent", userAgent)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// If the status is not 2xx, return the status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, resp.StatusCode, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to read response: %v", err)
	}

	return body, resp.StatusCode, nil
}

// mapTemperature maps a temperature value to cold/moderate/hot
func mapTemperature(temp int) string {
	if temp <= 30 {
		return "cold"
	}
	if temp >= 80 {
		return "hot"
	}
	return "moderate"
}
