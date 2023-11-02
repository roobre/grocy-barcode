package provider

import "errors"

type Multi []Provider

func (m Multi) Product(barcode string) (Product, error) {
	var errs []error

	for _, p := range m {
		product, err := p.Product(barcode)
		if err == nil {
			return product, nil
		}

		errs = append(errs, err)
	}

	return Product{}, errors.Join(errs...)
}
