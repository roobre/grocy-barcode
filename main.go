package main

import (
	"log"
	"os"

	"roob.re/grocy-barcode/barcodehid"
	"roob.re/grocy-barcode/grocy"
)

func main() {
	hid, _ := os.LookupEnv("GB_HID")
	if hid == "" {
		log.Fatalf("GB_HID must be set to the HID to use")
	}

	grocyServer, _ := os.LookupEnv("GB_GROCY_URL")
	if grocyServer == "" {
		log.Fatal("GB_GROCY_URL must be set to the URL for the grocy server")
	}

	grocyAPIKey, _ := os.LookupEnv("GB_GROCY_API_KEY")
	if grocyServer == "" {
		log.Fatal("GB_GROCY_API_KEY must be set to a grocy API key")
	}

	log.Printf("Opening HID barcode reader at %s", hid)
	dev, err := os.Open(hid)
	if err != nil {
		log.Fatal(err)
	}

	bs := barcodehid.New(dev)

	grocyClient := grocy.Client{
		Server: grocyServer,
		APIKey: grocyAPIKey,
		Defaults: grocy.Defaults{
			ProductUnitID:     2,
			ProductLocationID: 4,
		},
	}
	log.Printf("Created grocy client for %s", grocyServer)

	for {
		log.Printf("Ready to read barcode")
		barcode, err := bs.Read()
		if err != nil {
			log.Printf("error reading barcode: %v", err)
			return
		}

		err = grocyClient.AddOrCreate(barcode)
		if err != nil {
			log.Printf("error adding or creating %s: %v", barcode, err)
		}
	}
}
