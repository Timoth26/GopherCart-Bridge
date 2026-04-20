package domain

import "errors"

var (
	ErrProductNotFound      = errors.New("product not found")
	ErrProductAlreadyExists = errors.New("product already exists")

	ErrOrderNotFound = errors.New("order not found")
)
