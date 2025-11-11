# Forecast API

A Go HTTP server that provides weather forecasts using the National Weather Service (NWS) API. The service accepts latitude and longitude coordinates and returns a simplified weather forecast with temperature categorization.

## Features

- RESTful API endpoint for weather forecasts
- Integration with National Weather Service API
- Temperature categorization (cold/moderate/hot)
- Comprehensive error handling
- 79% test coverage with mocked API tests

## Prerequisites

- Go 1.16 or higher
- Make (optional, for using Makefile commands)

## Building the Forecast Server

### Using Make (Recommended)

Build the server:
```bash
make build
```

This will create a binary named `forecast` in the current directory.

### Using Go Directly

```bash
go build -o forecast .
```

### Build and Test

Run a full build and test cycle:
```bash
make all
```

## Running the Server

### Using Make

```bash
make run
```

### Running the Binary Directly

After building:
```bash
./forecast
```

The server will start on port 8080 and display:
```
Server starting on :8080
```

## API Usage

### Endpoint

```
GET /forecast
```

### Query Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| latitude | string | Yes | Latitude coordinate (e.g., "47.6062") |
| longitude | string | Yes | Longitude coordinate (e.g., "-122.3321") |

### Response Format

**Success Response (200 OK):**
```json
{
  "forecast": "Partly Cloudy",
  "temperature": "moderate"
}
```

**Temperature Categories:**
- `cold` - Temperature ≤ 30°F
- `moderate` - Temperature between 31°F and 79°F
- `hot` - Temperature ≥ 80°F

**Error Responses:**
- `400 Bad Request` - Missing latitude or longitude parameter
- `404 Not Found` - Forecast not available for the given coordinates
- `405 Method Not Allowed` - HTTP method other than GET
- `500 Internal Server Error` - Server or API error
- `503 Service Unavailable` - NWS API unavailable

## Examples

### Example 1: Get forecast for Seattle, WA

```bash
curl "http://localhost:8080/forecast?latitude=47.6062&longitude=-122.3321"
```

**Response:**
```json
{
  "forecast": "Partly Cloudy",
  "temperature": "moderate"
}
```

### Example 2: Get forecast for Phoenix, AZ (typically hot)

```bash
curl "http://localhost:8080/forecast?latitude=33.4484&longitude=-112.0740"
```

**Response:**
```json
{
  "forecast": "Sunny",
  "temperature": "hot"
}
```

### Example 3: Get forecast for Anchorage, AK (typically cold)

```bash
curl "http://localhost:8080/forecast?latitude=61.2181&longitude=-149.9003"
```

**Response:**
```json
{
  "forecast": "Snow Showers",
  "temperature": "cold"
}
```

### Example 4: Missing parameters (error case)

```bash
curl "http://localhost:8080/forecast?latitude=47.6062"
```

**Response:**
```
Missing latitude or longitude parameter
```
**Status Code:** 400

### Example 5: Invalid coordinates (error case)

```bash
curl "http://localhost:8080/forecast?latitude=999&longitude=999"
```

**Response:** NWS API error message
**Status Code:** 404 (or error code from NWS API)

## Testing

### Run all tests

```bash
make test
```

### Run tests with coverage

```bash
make coverage
```

This generates a detailed HTML coverage report in `coverage.html`.

### Using Go directly

```bash
# Run tests
go test -v ./...

# Run with coverage
go test -cover ./...
```

## Development

### Format code

```bash
make fmt
```

### Clean build artifacts

```bash
make clean
```

### Install dependencies

```bash
make deps
```

### View all available commands

```bash
make help
```

## Project Structure

```
.
├── main.go           # Server implementation
├── main_test.go      # Unit tests with mocked NWS API
├── Makefile          # Build and test automation
├── go.mod            # Go module definition
└── README.md         # This file
```

## How It Works

1. Client sends GET request to `/forecast` with latitude and longitude
2. Server calls NWS API `/points/{lat},{lon}` endpoint
3. Server extracts forecast URL from the points response
4. Server calls the forecast endpoint to get detailed weather data
5. Server extracts the first period's forecast and temperature
6. Server categorizes temperature as cold/moderate/hot
7. Server returns simplified JSON response to client

## API Integration

This service integrates with the National Weather Service API:
- **Points Endpoint:** `https://api.weather.gov/points/{latitude},{longitude}`
- **Forecast Endpoint:** Retrieved dynamically from points response
- **User-Agent:** All requests include `(murphybytes.com murphybytes@gmail.com)`

## License

This project is provided as-is for demonstration purposes.