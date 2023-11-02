package provider

type Provider interface {
	Product(barcode string) (Product, error)
}

type Product struct {
	Name        string
	Description string
}

func (p Product) String() string {
	return p.Name
}
