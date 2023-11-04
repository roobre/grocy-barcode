package off

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"roob.re/grocy-barcode/provider"
)

type product struct {
	Name     string   `json:"product_name"`
	Keywords []string `json:"_keywords"`
}

type OpenFoodFacts struct{}

func (OpenFoodFacts) Product(barcode string) (provider.Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	r, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("https://world.openfoodfacts.org/api/v0/product/%s.json", barcode), nil)
	if err != nil {
		return provider.Product{}, fmt.Errorf("building request: %w", err)
	}

	resp, err := http.DefaultClient.Do(r)
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

	if out.Product.Name == "" {
		return provider.Product{}, errors.New("got empty response from off")
	}

	return provider.Product{
		Name:        out.Product.Name,
		Description: strings.Join(out.Product.Keywords, ", "),
	}, nil
}
