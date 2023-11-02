package off

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"roob.re/grocy-barcode/provider"
)

type product struct {
	Name     string   `json:"product_name"`
	Keywords []string `json:"_keywords"`
}

type OpenFoodFacts struct{}

func (OpenFoodFacts) Product(barcode string) (provider.Product, error) {
	resp, err := http.Get(fmt.Sprintf("https://world.openfoodfacts.org/api/v0/product/%s.json", barcode))
	if err != nil {
		return provider.Product{}, fmt.Errorf("making request to off: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return provider.Product{}, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	var out struct {
		Product product `json:"product"`
	}

	err = json.NewDecoder(resp.Body).Decode(&out)
	if err != nil {
		return provider.Product{}, fmt.Errorf("reading product json: %w", err)
	}

	return provider.Product{
		Name:        out.Product.Name,
		Description: strings.Join(out.Product.Keywords, ", "),
	}, nil
}
