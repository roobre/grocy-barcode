package off_test

import (
	"testing"

	"roob.re/grocy-barcode/off"
)

func TestQuery(t *testing.T) {
	off := off.OpenFoodFacts{}
	product, err := off.Product("8411525020169")
	if err != nil {
		t.Fatalf("querying product: %v", err)
	}

	if product.Name == "" {
		t.Fatal("Got empty product")
	}

	t.Log(product)
}
