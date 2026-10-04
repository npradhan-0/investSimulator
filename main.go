package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

type Quote struct {
	CurrentPrice  float64 `json:"c"`
	Change        float64 `json:"d"`
	PercentChange float64 `json:"dp"`
	High          float64 `json:"h"`
	Low           float64 `json:"l"`
	Open          float64 `json:"o"`
	PreviousClose float64 `json:"pc"`
	Timestamp     int64   `json:"t"`
}

func main() {
	// Load the .env file
	err := godotenv.Load()

	if err != nil {
		fmt.Println("Error loading .env file")
		return
	}

	// Get the Finnhub API key
	apiKey := os.Getenv("FINNHUB_API_KEY")

	if apiKey == "" {
		fmt.Println("Finnhub API key is not set")
		return
	}

	// Stock we want to collect data for, need to hookup with the search symbol
	symbol := "NVDA"

	// Build the Finnhub API URL
	url := "https://finnhub.io/api/v1/quote?symbol=" + symbol + "&token=" + apiKey

	// Make the API request
	response, err := http.Get(url)

	if err != nil {
		fmt.Println("Error making API request:", err)
		return
	}

	defer response.Body.Close()

	// Check the HTTP response
	if response.StatusCode != http.StatusOK {
		fmt.Println("API request failed with status:", response.Status)
		return
	}

	// Convert JSON response into our Quote struct
	var quote Quote

	err = json.NewDecoder(response.Body).Decode(&quote)

	if err != nil {
		fmt.Println("Error decoding API response:", err)
		return
	}

	// results display
	fmt.Println("INVESTSIMULATOR")
	fmt.Println("-----------------")
	fmt.Println("Symbol:", symbol)
	fmt.Printf("Current Price: $%.2f\n", quote.CurrentPrice)
	fmt.Printf("Change: $%.2f\n", quote.Change)
	fmt.Printf("Percent Change: %.2f%%\n", quote.PercentChange)
	fmt.Printf("Day High: $%.2f\n", quote.High)
	fmt.Printf("Day Low: $%.2f\n", quote.Low)
	fmt.Printf("Open: $%.2f\n", quote.Open)
	fmt.Printf("Previous Close: $%.2f\n", quote.PreviousClose)

}
