package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"slices"
	"strconv"
	"sync"
	"time"
)

type Stock struct {
	Ticker       string
	Gap          float64
	OpeningPrice float64
}

type Position struct {
	EntryPrice      float64
	Shares          int
	TakeProfitPrice float64
	StopLossPrice   float64
	Profit          float64
}

type Selection struct {
	Ticker string
	Position
	Articles []Article
}

type attributes struct {
	PublishOn time.Time `json:"publishOn"`
	Title     string    `json:"title"`
}

type seekingAlphaNews struct {
	Attributes attributes `json:"attributeS"`
}

type SeekingAlphaResponse struct {
	Data []seekingAlphaNews `json:"data`
}

type Article struct {
	PublishOn time.Time
	Headline  string
}

func Load(path string) ([]Stock, error) {
	f, err := os.Open(path) //if is the opened csv file

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	defer f.Close()

	r := csv.NewReader(f)
	rows, err := r.ReadAll()

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	rows = slices.Delete(rows, 0, 1)

	var stocks []Stock

	for _, row := range rows {
		ticker := row[0]
		gap, err := strconv.ParseFloat(row[1], 64)

		if err != nil {
			continue
		}

		openingPrice, err := strconv.ParseFloat(row[2], 64)

		if err != nil {
			continue
		}

		stocks = append(stocks, Stock{
			Ticker:       ticker,
			Gap:          gap,
			OpeningPrice: openingPrice,
		})
	}
	return stocks, nil
}

func Calculate(gapPercent, openingPrice float64) Position {
	closingPrice := openingPrice / (1 + gapPercent)
	gapValue := closingPrice - openingPrice
	profitFromGap := profitPercent * gapValue
	stopLoss := openingPrice - profitFromGap
	takeProfit := openingPrice + profitFromGap
	shares := int(maxLossPerTrade / math.Abs(stopLoss-openingPrice))

	profit := math.Abs(openingPrice-takeProfit) * float64(shares)
	profit = math.Round(profit*100) / 100

	return Position{
		EntryPrice:      math.Round(openingPrice*100) / 100,
		Shares:          shares,
		TakeProfitPrice: math.Round(takeProfit*100) / 100,
		StopLossPrice:   math.Round(stopLoss*100) / 100,
		Profit:          math.Round(profit*100) / 100,
	}
}

const (
	url          = "https://seeking-alpha.p.rapidapi.com/news/v2/list-by-symbol?size=5&id="
	apiKeyHeader = "x-rapidapi-key"
	apiKey       = "f8bbeecee3mshd11af53a602b7fcp133f59jsn3b4ffb3f20df"
)

func FetchNews(ticker string) ([]Article, error) {
	req, err := http.NewRequest(http.MethodGet, url+ticker, nil)

	if err != nil {
		return nil, err
	}

	req.Header.Add(apiKeyHeader, apiKey)
	client := &http.Client{}
	resp, err := client.Do(req) //returns pointerto http res type and an error

	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("unsuccessful status code received", resp.StatusCode)
	}

	res := &SeekingAlphaResponse{}

	json.NewDecoder(resp.Body).Decode(res) //converts res into json type

	var articles []Article

	for _, item := range res.Data {
		art := Article{
			PublishOn: item.Attributes.PublishOn,
			Headline:  item.Attributes.Title,
		}
		articles = append(articles, art)
	}
	return articles, nil
}

func Deliver(filePath string, selections []Selection) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(selections)
	if err != nil {
		return fmt.Errorf("error encoding selections: %w", err)
	}
	return nil
}

var accountBalance = 10000.0
var lossTolerance = .02
var maxLossPerTrade = accountBalance * lossTolerance
var profitPercent = .8

func main() {
	stocks, err := Load("./opg.csv")
	if err != nil {
		fmt.Println(err)
		return
	}
	stocks = slices.DeleteFunc(stocks, func(stock Stock) bool {
		return math.Abs(stock.Gap) < .1
	})

	var selections []Selection

	var wg sync.WaitGroup
	for _, stock := range stocks {
		wg.Add(1)

		go func(s Stock) {
			defer wg.Done()
			position := Calculate(stock.Gap, stock.OpeningPrice)
			articles, err := FetchNews(stock.Ticker)

			if err != nil {
				log.Printf("error loading news %s, %v", stock.Ticker, err)
			} else {
				log.Printf("Found %d articles about %s", len(articles), stock.Ticker)
			}

			sel := Selection{
				Ticker:   stock.Ticker,
				Position: position,
				Articles: articles,
			}
			selections = append(selections, sel)
		}(stock)
	}
	wg.Wait()

	outputPath := "./opg.json"
	err = Deliver(outputPath, selections)
	if err != nil {
		log.Printf("Error writing output, %v", err)
		return
	}
	log.Printf("Finished writing output to %s\n", outputPath)
}
//339