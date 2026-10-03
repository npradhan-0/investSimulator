package main

import "fmt"

func main() {
	fmt.Println("InvestSimulator - Market Data Agent")
	if [ -n "$FINNHUB_API_KEY" ]; then echo "Finnhub key is set"; else echo "Finnhub key is NOT set"; fi
}
