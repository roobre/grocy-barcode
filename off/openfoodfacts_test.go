package off_test

import (
	"testing"

	"roob.re/grocy-barcode/off"
)

func TestQuery(t *testing.T) {
	product, err := off.Query("8411525020169")
	if err != nil {
		t.Fatalf("querying product: %v", err)
	}

	if product.Name == "" {
		t.Fatal("Got empty product")
	}

	t.Log(product)
}
