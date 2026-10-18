package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	loggly "github.com/jamespearly/loggly"
	"github.com/joho/godotenv"
)

const tableName = "npradhan_StockQuotes"

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

// QuoteItem is the structure stored in DynamoDB.
type QuoteItem struct {
	Symbol         string  `dynamodbav:"Symbol"`
	Timestamp      string  `dynamodbav:"TimeStamp"`
	QuoteTimestamp int64   `dynamodbav:"QuoteTimestamp"`
	CurrentPrice   float64 `dynamodbav:"CurrentPrice"`
	Change         float64 `dynamodbav:"Change"`
	PercentChange  float64 `dynamodbav:"PercentChange"`
	High           float64 `dynamodbav:"High"`
	Low            float64 `dynamodbav:"Low"`
	Open           float64 `dynamodbav:"Open"`
	PreviousClose  float64 `dynamodbav:"PreviousClose"`
}

func fetchQuote(
	symbol string,
	apiKey string,
	client *loggly.ClientType,
	db *dynamodb.Client,
) {
	url := "https://finnhub.io/api/v1/quote?symbol=" + symbol + "&token=" + apiKey

	response, err := http.Get(url)
	if err != nil {
		fmt.Println("Error making API request:", err)
		if logErr := client.Send("error", "Finnhub request failed for "+symbol+": "+err.Error()); logErr != nil {
			fmt.Println("Loggly error:", logErr)
		}
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		fmt.Println("API request failed status:", response.Status)
		if logErr := client.Send("error", "Finnhub request failed for "+symbol+": "+response.Status); logErr != nil {
			fmt.Println("Loggly error:", logErr)
		}
		return
	}

	var quote Quote
	if err := json.NewDecoder(response.Body).Decode(&quote); err != nil {
		fmt.Println("Error decoding API response:", err)
		if logErr := client.Send("error", "Finnhub response decoding failed for "+symbol+": "+err.Error()); logErr != nil {
			fmt.Println("Loggly error:", logErr)
		}
		return
	}

	if quote.CurrentPrice == 0 {
		fmt.Println("API request returned no valid market data")
		if logErr := client.Send("error", "Finnhub returned no valid market data for "+symbol); logErr != nil {
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

	// send the successful quote to Loggly.
	if err := client.Send(
		"info",
		"Finnhub request successful: collected 1 quote for "+symbol,
	); err != nil {
		fmt.Println("Loggly error:", err)
	}

	collectedAt := time.Now().UTC().Format("2006-01-02T15:04:05.000000000Z")

	item := QuoteItem{
		Symbol:         symbol,
		Timestamp:      collectedAt,
		QuoteTimestamp: quote.Timestamp,
		CurrentPrice:   quote.CurrentPrice,
		Change:         quote.Change,
		PercentChange:  quote.PercentChange,
		High:           quote.High,
		Low:            quote.Low,
		Open:           quote.Open,
		PreviousClose:  quote.PreviousClose,
	}

	ddbItem, err := attributevalue.MarshalMap(item)
	if err != nil {
		fmt.Println("Error preparing DynamoDB item:", err)
		_ = client.Send("error", "Failed to prepare DynamoDB item for "+symbol+": "+err.Error())
		return
	}

	_, err = db.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      ddbItem,
	})
	if err != nil {
		fmt.Println("Error writing to DynamoDB:", err)
		if logErr := client.Send("error", "DynamoDB write failed for "+symbol+": "+err.Error()); logErr != nil {
			fmt.Println("Loggly error:", logErr)
		}
		return
	}

	fmt.Println("Successfully saved quote to DynamoDB.")
	fmt.Println("DynamoDB table:", tableName)
	fmt.Println("Record timestamp:", collectedAt)
	fmt.Println()
}

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found; using environment variables")
	}

	apiKey := os.Getenv("FINNHUB_API_KEY")
	if apiKey == "" {
		fmt.Println("Finnhub API key is not set")
		return
	}

	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-east-1"
	}

	// Load AWS credentials from the standard AWS credential.
	awsConfig, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion(region),
	)
	if err != nil {
		fmt.Println("Error loading AWS configuration:", err)
		return
	}

	db := dynamodb.NewFromConfig(awsConfig)

	client := loggly.New("InvestSimulator-MarketData")

	// Stock to collect data for, need to hookup with the search symbol API later
	symbol := "NVDA"

	// Polling every 5 minutes
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	fmt.Println("InvestSimulator Market Data Agent started.")
	fmt.Println("Polling Finnhub every 5 minutes...")
	fmt.Println("DynamoDB table:", tableName)
	fmt.Println("AWS Region:", region)
	fmt.Println()

	for {
		fetchQuote(symbol, apiKey, client, db)

		fmt.Println("Waiting 5 minutes for next request...")
		fmt.Println()

		<-ticker.C
	}
}
