package trade

import (
	"github.com/louistam888/stockAnalysis/internal/news"
	"github.com/louistam888/stockAnalysis/internal/pos"
)

type Selection struct {
	Ticker string
	pos.Position
	Articles []news.Article
}

type Deliverer interface {
	Deliver(selections []Selection) error
}
