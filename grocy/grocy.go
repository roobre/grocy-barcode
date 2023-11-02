package grocy

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"roob.re/grocy-barcode/off"
)

type Client struct {
	Server   string
	APIKey   string
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

	response, err := c.Get("/stock/products/by-barcode/"+barcode, nil)
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

func (c Client) ConsumeOrCreate(barcode string) error {
	return nil
}

func (c Client) Create(barcode string) (createdResponse, error) {
	offProduct, err := off.Query(barcode)
	if err != nil {
		return createdResponse{}, fmt.Errorf("querying off: %w", err)
	}

	if offProduct.Name == "" {
		return createdResponse{}, errors.New("OFF returned an empty response")
	}

	log.Printf("Found product %s", offProduct)

	productCreatePayload := map[string]any{
		"name":           offProduct.Name,
		"description":    strings.Join(offProduct.Keywords, ", "),
		"active":         1,
		"location_id":    c.Defaults.ProductLocationID,
		"qu_id_stock":    c.Defaults.ProductUnitID,
		"qu_id_purchase": c.Defaults.ProductUnitID,
		"qu_id_consume":  c.Defaults.ProductUnitID,
		"qu_id_price":    c.Defaults.ProductUnitID,
	}
	productResponse := createdResponse{}
	response, err := c.Post("/objects/products", productCreatePayload, &productResponse)
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
	response, err = c.Post("/objects/product_barcodes", prodBarcodeRequest, prodBarcodeResponse)
	if err != nil {
		return createdResponse{}, fmt.Errorf("creating product_barcode: %w", err)
	}

	// Grocy returns 200 while it should return 201, accept both.
	if !statusIs(response.StatusCode, http.StatusOK, http.StatusCreated) {
		return createdResponse{}, fmt.Errorf("unexpected status %d", response.StatusCode)
	}

	log.Printf("Created product %s and associated to %s", offProduct, barcode)

	return prodBarcodeResponse, nil
}

func (c Client) Add(barcode string) error {
	addRequest := map[string]any{
		"amount": 1,
	}
	response, err := c.Post(fmt.Sprintf("/stock/products/by-barcode/%s/add", barcode), addRequest, nil)
	if err != nil {
		return fmt.Errorf("adding product: %w", err)
	}

	if !statusIs(response.StatusCode, http.StatusOK, http.StatusCreated) {
		return fmt.Errorf("got unexpected status %d", response.StatusCode)
	}

	return nil
}

func (c Client) Get(path string, dest any) (*http.Response, error) {
	resp, err := c.Do(http.MethodGet, path, nil)
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

func (c Client) Post(path string, data any, dest any) (*http.Response, error) {
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(data)
	if err != nil {
		return nil, fmt.Errorf("json-encoding data: %w", err)
	}

	resp, err := c.Do(http.MethodPost, path, buf)
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

func (c Client) Do(method, path string, body io.Reader) (*http.Response, error) {
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
