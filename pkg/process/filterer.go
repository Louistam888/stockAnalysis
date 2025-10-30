package process

import (
	"math"
	"slices"

	"github.com/louistam888/stockAnalysis/internal/raw"
)

type filterer struct {
	minGap float64
}

func (n *filterer) Filter(candidates []raw.Stock) (filtered []raw.Stock) {
	filtered = slices.DeleteFunc(candidates, func(s raw.Stock) bool {
		return math.Abs(s.Gap) < n.minGap
	})
	return
}

func NewFilterer(minGap float64) raw.Filterer {
	return &filterer{
		minGap: minGap,
	}
}
