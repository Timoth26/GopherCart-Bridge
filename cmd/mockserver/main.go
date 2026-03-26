package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type Product struct {
	ID    int64   `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
	Stock int     `json:"stock"`
}

type Order struct {
	ID        int64 `json:"id"`
	ProductID int64 `json:"product_id"`
	Quantity  int   `json:"quantity"`
}

var products = []Product{
	{ID: 1, Name: "Laptop", Price: 2999.99, Stock: 10},
	{ID: 2, Name: "Mouse", Price: 49.99, Stock: 100},
	{ID: 3, Name: "Keyboard", Price: 149.99, Stock: 50},
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /products", handleGetProducts)
	mux.HandleFunc("POST /orders", handlePostOrder)

	log.Println("Mock supplier server running on :9090")
	if err := http.ListenAndServe(":9090", mux); err != nil {
		log.Fatal(err)
	}
}

func handleGetProducts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

func handlePostOrder(w http.ResponseWriter, r *http.Request) {
	var order Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("Received order: productID=%d quantity=%d", order.ProductID, order.Quantity)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
		"status":   "accepted",
		"order_id": order.ID,
	})
}
