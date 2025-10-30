package root

import (
	"fmt"
	"log"

	"github.com/louistam888/stockAnalysis/internal/news"
	"github.com/louistam888/stockAnalysis/internal/pos"
	"github.com/louistam888/stockAnalysis/internal/raw"
	"github.com/louistam888/stockAnalysis/internal/trade"
)

func Run(ldr raw.Loader, f raw.Filterer, c pos.Calculator, fet news.Fetcher, del trade.Deliverer) error {
	stocks, err := ldr.Load()

	if err != nil {
		return fmt.Errorf("error loading stocks: %w", err)
	}

	stocks = f.Filter(stocks)

	selectionsChan := make(chan trade.Selection, len(stocks))

	for _, stock := range stocks {
		go func(s raw.Stock, selected chan<- trade.Selection) {
			position := c.Calculate(s.Gap, s.OpeningPrice)
			articles, err := fet.Fetch(s.Ticker)
			if err != nil {
				log.Printf("error loading news about %s, %v", s.Ticker, err)
				selected <- trade.Selection{}
				return
			} else {
				log.Printf("Found %d articles about %s", len(articles), s.Ticker)
			}

			sel := trade.Selection{
				Ticker:   s.Ticker,
				Articles: articles,
				Position: position,
			}
			selected <- sel
		}(stock, selectionsChan)

		var selections []trade.Selection
		for sel := range selectionsChan {
			selections = append(selections, sel)
			if len(selections) == len(stocks) {
				close(selectionsChan)
			}
		}

		err = del.Deliver(selections)
		if err != nil {
			fmt.Errorf("error delivering selections: %w", err)
		}
	}
	return nil
}
