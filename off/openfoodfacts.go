package off

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Product struct {
	Name     string   `json:"product_name"`
	Keywords []string `json:"_keywords"`
}

func (p Product) String() string {
	return fmt.Sprintf("%s (%v)", p.Name, p.Keywords)
}

func Query(barcode string) (Product, error) {
	resp, err := http.Get(fmt.Sprintf("https://world.openfoodfacts.org/api/v0/product/%s.json", barcode))
	if err != nil {
		return Product{}, fmt.Errorf("making request to off: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return Product{}, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	var out struct {
		Product Product `json:"product"`
	}

	err = json.NewDecoder(resp.Body).Decode(&out)
	if err != nil {
		return Product{}, fmt.Errorf("reading product json: %w", err)
	}

	return out.Product, nil
}
