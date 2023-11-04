package grocybarcode

import (
	"fmt"
	"log"

	"roob.re/grocy-barcode/grocy"
)

type GrocyBarcode struct {
	Grocy  grocy.Client
	Reader BarcodeReader
}

type handlerFunc func(barcode string) error

func (g GrocyBarcode) Start() error {
	handler := g.Grocy.AddOrCreate

	for {
		log.Printf("Ready to read barcode")
		barcode, err := g.Reader.Read()
		if err != nil {
			return fmt.Errorf("error reading barcode: %w", err)
		}

		// See if this is a mode-changing barcode.
		if newHandler := g.handler(barcode); newHandler != nil {
			log.Printf("Switching to mode %s", barcode)
			handler = newHandler
			continue
		}

		err = handler(barcode)
		if err != nil {
			log.Printf("error processing %s: %v", barcode, err)
		}
	}
}

func (g GrocyBarcode) handler(barcode string) handlerFunc {
	switch barcode {
	case "addcreate":
		return g.Grocy.AddOrCreate
	case "consume":
		return g.Grocy.Consume
	case "open":
		return g.Grocy.Open
	default:
		return nil
	}
}

type BarcodeReader interface {
	Read() (string, error)
}
