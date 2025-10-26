package json

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/louistam888/stockAnalysis/internal/trade"
)

type deliverer struct {
	filePath string
}

func (d *deliverer) Deliver(selections []trade.Selection) error {
	file, err := os.Create(d.filePath)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(selections)
	if err != nil {
		return fmt.Errorf("error encoding selections: %w", err)
	}

	log.Printf("Finished writing output to %s\n", d.filePath)
	return nil
}

func NewDeliverer(filePath string) trade.Deliverer {
	return &deliverer{filePath: filePath}
}

//cmd folder at 352
