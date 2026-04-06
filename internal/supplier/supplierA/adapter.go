package supplierA

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/Timoth26/GopherCart-Bridge/internal/domain"
	"github.com/Timoth26/GopherCart-Bridge/internal/supplier"
)

type Client struct {
	supplier.Base
	endpoint *url.URL
}

type supplierProduct struct {
	ID    int64   `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
	Stock int     `json:"stock"`
}

type supplierOrder struct {
	ID        int   `json:"id"`
	ProductID int64 `json:"product_id"`
	Quantity  int   `json:"quantity"`
}

var _ domain.SupplierClient = (*Client)(nil)

func NewClient(baseURL string, timeout time.Duration, log *slog.Logger) (*Client, error) {
	endpoint, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse supplier base URL: %w", err)
	}
	return &Client{
		Base:     supplier.NewBase(timeout, log),
		endpoint: endpoint,
	}, nil
}

func (c *Client) FetchProducts(ctx context.Context) ([]domain.Product, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint.JoinPath("products").String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch products: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var supplierProducts []supplierProduct
	if err := json.NewDecoder(resp.Body).Decode(&supplierProducts); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return mapToDomain(supplierProducts), nil
}

func mapToDomain(products []supplierProduct) []domain.Product {
	result := make([]domain.Product, len(products))
	for i, p := range products {
		result[i] = domain.Product{
			ID:    p.ID,
			Name:  p.Name,
			Price: p.Price,
			Stock: p.Stock,
		}
	}
	return result
}

func (c *Client) SendOrder(ctx context.Context, order domain.Order) error {
	for _, item := range order.Items {
		body, err := json.Marshal(supplierOrder{
			ID:        order.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})
		if err != nil {
			return fmt.Errorf("marshal order item (product %d): %w", item.ProductID, err)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint.JoinPath("orders").String(), bytes.NewReader(body))
		if err != nil {
			return fmt.Errorf("create request (product %d): %w", item.ProductID, err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.HTTP.Do(req)
		if err != nil {
			return fmt.Errorf("send order item (product %d): %w", item.ProductID, err)
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			return fmt.Errorf("unexpected status code for product %d: %d", item.ProductID, resp.StatusCode)
		}
	}

	return nil
}
