package main

import (
	"log"
	"os"

	grocybarcode "roob.re/grocy-barcode"
	"roob.re/grocy-barcode/barcodehid"
	"roob.re/grocy-barcode/barcodetext"
	"roob.re/grocy-barcode/grocy"
	"roob.re/grocy-barcode/off"
)

func main() {
	hid, _ := os.LookupEnv("GB_HID")
	tty, _ := os.LookupEnv("GB_TTY")
	if hid == "" && tty == "" {
		log.Fatalf("GB_HID or GB_TTY must be set to the HID or TTY device to use")
	}

	grocyServer, _ := os.LookupEnv("GB_GROCY_URL")
	if grocyServer == "" {
		log.Fatal("GB_GROCY_URL must be set to the URL for the grocy server")
	}

	grocyAPIKey, _ := os.LookupEnv("GB_GROCY_API_KEY")
	if grocyServer == "" {
		log.Fatal("GB_GROCY_API_KEY must be set to a grocy API key")
	}

	var br grocybarcode.BarcodeReader
	if hid != "" {
		log.Printf("Opening HID barcode reader at %s", hid)
		dev, err := os.Open(hid)
		if err != nil {
			log.Fatal(err)
		}

		br = barcodehid.New(dev)
	}

	if tty != "" {
		log.Printf("Opening TTY barcode reader at %s", hid)
		dev, err := os.Open(tty)
		if err != nil {
			log.Fatal(err)
		}

		br = barcodetext.New(dev)
	}

	grocyClient := grocy.Client{
		Server:   grocyServer,
		APIKey:   grocyAPIKey,
		Provider: off.OpenFoodFacts{},
		Defaults: grocy.Defaults{
			ProductUnitID:     2,
			ProductLocationID: 4,
		},
	}

	log.Printf("Created grocy client for %s", grocyServer)

	gb := grocybarcode.GrocyBarcode{
		Grocy:  grocyClient,
		Reader: br,
	}

	err := gb.Start()
	if err != nil {
		log.Fatal(err)
	}
}
