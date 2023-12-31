package grocy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"roob.re/grocy-barcode/provider"
)

type Client struct {
	Server   string
	APIKey   string
	Provider provider.Provider
	Defaults Defaults
}

type Defaults struct {
	ProductLocationID int
	ProductUnitID     int
}

type createdResponse struct {
	ObjectID string `json:"created_object_id"`
}

func (c Client) AddOrCreate(barcode string) error {
	log.Printf("Adding or creating %s", barcode)

	response, err := c.get("/stock/products/by-barcode/"+barcode, nil)
	if err != nil {
		return fmt.Errorf("checking for existence of product: %w", err)
	}

	// Grocy returns 400 while it should return 404, accept both.
	if !statusIs(response.StatusCode, http.StatusOK, http.StatusBadRequest, http.StatusNotFound) {
		return fmt.Errorf("checking existence of product: unexpected status code %d", response.StatusCode)
	}

	if response.StatusCode != http.StatusOK {
		log.Printf("%s not registered, creating product", barcode)
		_, err := c.Create(barcode)
		if err != nil {
			return fmt.Errorf("creating missing product: %w", err)
		}
	}

	log.Printf("Adding one unit of %s", barcode)
	err = c.Add(barcode)
	if err != nil {
		return fmt.Errorf("purchasing product: %w", err)
	}

	return nil
}

func (c Client) Create(barcode string) (createdResponse, error) {
	product, err := c.Provider.Product(barcode)
	if err != nil {
		return createdResponse{}, fmt.Errorf("querying provider: %w", err)
	}

	log.Printf("Found product %s", product)

	productCreatePayload := map[string]any{
		"name":                     product.Name,
		"description":              product.Description,
		"active":                   1,
		"default_best_before_days": -1,
		"location_id":              c.Defaults.ProductLocationID,
		"qu_id_stock":              c.Defaults.ProductUnitID,
		"qu_id_purchase":           c.Defaults.ProductUnitID,
		"qu_id_consume":            c.Defaults.ProductUnitID,
		"qu_id_price":              c.Defaults.ProductUnitID,
	}
	productResponse := createdResponse{}
	response, err := c.post("/objects/products", productCreatePayload, &productResponse)
	if err != nil {
		return createdResponse{}, fmt.Errorf("creating product: %w", err)
	}

	// Grocy returns 200 while it should return 201, accept both.
	if !statusIs(response.StatusCode, http.StatusOK, http.StatusCreated) {
		return createdResponse{}, fmt.Errorf("unexpected status %d", response.StatusCode)
	}

	prodBarcodeRequest := map[string]any{
		"product_id": productResponse.ObjectID,
		"barcode":    barcode,
	}

	prodBarcodeResponse := createdResponse{}
	response, err = c.post("/objects/product_barcodes", prodBarcodeRequest, prodBarcodeResponse)
	if err != nil {
		return createdResponse{}, fmt.Errorf("creating product_barcode: %w", err)
	}

	// Grocy returns 200 while it should return 201, accept both.
	if !statusIs(response.StatusCode, http.StatusOK, http.StatusCreated) {
		return createdResponse{}, fmt.Errorf("unexpected status %d", response.StatusCode)
	}

	log.Printf("Created product %s and associated to %s", product, barcode)

	return prodBarcodeResponse, nil
}

func (c Client) Add(barcode string) error {
	return c.byBarcode("add", barcode)
}

func (c Client) Consume(barcode string) error {
	log.Printf("Consuming one unit of %s", barcode)
	return c.byBarcode("consume", barcode)
}

func (c Client) Open(barcode string) error {
	log.Printf("Opening one unit of %s", barcode)
	return c.byBarcode("open", barcode)
}

func (c Client) byBarcode(action, barcode string) error {
	request := map[string]any{
		"amount": 1,
	}

	response, err := c.post(fmt.Sprintf("/stock/products/by-barcode/%s/%s", barcode, action), request, nil)
	if err != nil {
		return fmt.Errorf("adding product: %w", err)
	}

	if !statusIs(response.StatusCode, http.StatusOK, http.StatusCreated) {
		return fmt.Errorf("got unexpected status %d", response.StatusCode)
	}

	return nil
}

func (c Client) get(path string, dest any) (*http.Response, error) {
	resp, err := c.do(http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("performing request to %q: %w", path, err)
	}

	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	if dest == nil {
		return resp, nil
	}

	err = json.NewDecoder(resp.Body).Decode(&dest)
	if err != nil {
		return resp, fmt.Errorf("decoding body: %w", err)
	}

	return resp, nil
}

func (c Client) post(path string, data any, dest any) (*http.Response, error) {
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(data)
	if err != nil {
		return nil, fmt.Errorf("json-encoding data: %w", err)
	}

	resp, err := c.do(http.MethodPost, path, buf)
	if err != nil {
		return nil, fmt.Errorf("performing request to %q: %w", path, err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if dest == nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		return resp, nil
	}

	err = json.NewDecoder(resp.Body).Decode(&dest)
	if err != nil {
		return resp, fmt.Errorf("decoding body: %w", err)
	}

	return resp, nil
}

func (c Client) do(method, path string, body io.Reader) (*http.Response, error) {
	u, err := url.JoinPath(c.Server, "api", path)
	if err != nil {
		return nil, fmt.Errorf("building url: %w", err)
	}

	req, err := http.NewRequest(method, u, body)
	if err != nil {
		return nil, fmt.Errorf("building request %w", err)
	}

	req.Header.Add("content-type", "application/json")
	req.Header.Add("accept", "application/json")
	req.Header.Add("GROCY-API-KEY", c.APIKey)

	return http.DefaultClient.Do(req)
}

func statusIs(actual int, expecteds ...int) bool {
	for _, expected := range expecteds {
		if actual == expected {
			return true
		}
	}

	return false
}
