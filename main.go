package main

import (
	"flag"
	"fmt"
	"os"

	root "github.com/louistam888/stockAnalysis/cmd"
	"github.com/louistam888/stockAnalysis/internal/news"
	"github.com/louistam888/stockAnalysis/internal/pos"
	"github.com/louistam888/stockAnalysis/internal/raw"
	"github.com/louistam888/stockAnalysis/internal/trade"
	"github.com/louistam888/stockAnalysis/pkg/csv"
	"github.com/louistam888/stockAnalysis/pkg/json"
	"github.com/louistam888/stockAnalysis/pkg/process"
	"github.com/louistam888/stockAnalysis/pkg/salpha"
)

// const (
// 	url          = "https://seeking-alpha.p.rapidapi.com/news/v2/list-by-symbol?size=5&id="
// 	apiKeyHeader = "x-rapidapi-key"
// 	apiKey       = "f8bbeecee3mshd11af53a602b7fcp133f59jsn3b4ffb3f20df"
// )

func main() {
	var seekingAlphaURL = os.Getenv("SEEKING_ALPHA_URL")
	var seekingAlphaAPIKey = os.Getenv("SEEKING_ALPHA_API_KEY")

	//validate env varibales
	if seekingAlphaURL == "" {
		fmt.Println("Missing SEEKING_ALPHA_URL environment variable")
		os.Exit(1)
	}
	if seekingAlphaAPIKey == "" {
		fmt.Println("Missing SEEKING_ALPHA_API_KEY environment variable")
		os.Exit(1)
	}

	//define commandline flags

	inputPath := flag.String("i", "", "path to input file (required)")
	accountBalance := flag.Float64("b", 0.0, "Account balance (required)")
	outputPath := flag.String("o", "./opg/json", "Path to output file")
	lossTolerance := flag.Float64("l", 0.02, "Loss tolerance percentage")
	profitPercent := flag.Float64("p", 0.8, "Percentage of the gap to take as a profit")
	minGap := flag.Float64("m", 0.1, "Minimum gap value to consider")

	//parse command lineflags
	flag.Parse()

	//check if requried flags are provided

	if *inputPath == "" || *accountBalance == 0.0 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	var ldr raw.Loader = csv.NewLoader(*inputPath)
	var f raw.Filterer = process.NewFilterer(*minGap)
	var c pos.Calculator = process.NewCalculator(*accountBalance, *lossTolerance, *profitPercent)
	var fet news.Fetcher = salpha.NewClient(seekingAlphaURL, seekingAlphaAPIKey)
	var del trade.Deliverer = json.NewDeliverer(*outputPath)

	err := root.Run(ldr, f, c, fet, del)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
//356