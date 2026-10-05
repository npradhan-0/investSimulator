package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	loggly "github.com/JamesPEarly/loggly"
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

func fetchQuote(symbol string, apiKey string, client *loggly.ClientType) {
	url := "https://finnhub.io/api/v1/quote?symbol=" + symbol + "&token=" + apiKey

	response, err := http.Get(url)

	if err != nil {
		fmt.Println("Error making API request:", err)

		logErr := client.Send("error", "Finnhub request failed for "+symbol+": "+err.Error())

		if logErr != nil {
			fmt.Println("Loggly error:", logErr)
		}

		return
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		fmt.Println("API request failed status:", response.Status)

		err = client.Send("error", "Finnhub request failed for "+symbol+": "+response.Status)

		if err != nil {
			fmt.Println("Loggly error:", err)
		}

		return
	}

	var quote Quote

	err = json.NewDecoder(response.Body).Decode(&quote)

	if err != nil {
		fmt.Println("Error decoding API response:", err)

		logErr := client.Send("error", "Finnhub response decoding failed for "+symbol+": "+err.Error())

		if logErr != nil {
			fmt.Println("Loggly error:", logErr)
		}

		return
	}

	if quote.CurrentPrice == 0 {
		fmt.Println("API request returned no valid market data")

		logErr := client.Send("error", "Finnhub returned no valid market data for "+symbol)

		if logErr != nil {
			fmt.Println("Loggly error:", logErr)
		}

		return
	}

	// results display
	fmt.Println("INVESTSIMULATOR")
	fmt.Println("-----------------")
	fmt.Println("Time:", time.Now())
	fmt.Println("Symbol:", symbol)
	fmt.Printf("Current Price: $%.2f\n", quote.CurrentPrice)
	fmt.Printf("Change: $%.2f\n", quote.Change)
	fmt.Printf("Percent Change: %.2f%%\n", quote.PercentChange)
	fmt.Printf("Day High: $%.2f\n", quote.High)
	fmt.Printf("Day Low: $%.2f\n", quote.Low)
	fmt.Printf("Open: $%.2f\n", quote.Open)
	fmt.Printf("Previous Close: $%.2f\n", quote.PreviousClose)

	logErr := client.Send(
		"info",
		"Finnhub request successful: collected 1 quote for "+symbol,
	)

	if logErr != nil {
		fmt.Println("Loggly error:", logErr)
	}

	fmt.Println("")
}

func main() {
	err := godotenv.Load()

	if err != nil {
		fmt.Println("Error loading .env file")
		return
	}

	apiKey := os.Getenv("FINNHUB_API_KEY")

	if apiKey == "" {
		fmt.Println("Finnhub API key is not set")
		return
	}

	client := loggly.New("InvestSimulator-MarketData")

	// Stock to collect data for, need to hookup with the search symbol API later
	symbol := "NVDA"

	// Polling every 5 minutes
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	fmt.Println("InvestSimulator Market Data Agent started.")
	fmt.Println("Polling Finnhub every 5 minutes...")
	fmt.Println()

	for {
		fetchQuote(symbol, apiKey, client)

		fmt.Println("Waiting 5 minutes for next request...")
		fmt.Println()

		<-ticker.C
	}
}
