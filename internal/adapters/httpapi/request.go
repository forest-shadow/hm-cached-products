package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/forest-shadow/hm-cached-products/internal/product"
)

// Shifting 1 left by 20 bits gives 2^20 bytes (1,048,576 bytes, or 1 MiB).
const maxRequestBodySize = 1 << 20

type createProductRequest struct {
	Name  *string `json:"name"`
	Price *int64  `json:"price"`
}

func (r createProductRequest) toProductInput() (product.Input, error) {
	if r.Name == nil {
		return product.Input{}, errNameRequired
	}
	if r.Price == nil {
		return product.Input{}, errPriceRequired
	}
	return product.Input{Name: *r.Name, Price: *r.Price}, nil
}

type updateProductRequest struct {
	Name            *string `json:"name"`
	Price           *int64  `json:"price"`
	ExpectedVersion *int64  `json:"expected_version"`
}

type updateProductParams struct {
	Input           product.Input
	ExpectedVersion int64
}

func (r updateProductRequest) toUpdateProductParams() (updateProductParams, error) {
	if r.Name == nil {
		return updateProductParams{}, errNameRequired
	}
	if r.Price == nil {
		return updateProductParams{}, errPriceRequired
	}
	if r.ExpectedVersion == nil || *r.ExpectedVersion <= 0 {
		return updateProductParams{}, errInvalidVersion
	}
	return updateProductParams{
		Input: product.Input{
			Name:  *r.Name,
			Price: *r.Price,
		},
		ExpectedVersion: *r.ExpectedVersion,
	}, nil
}

func decodeJSON[T any](w http.ResponseWriter, r *http.Request) (T, error) {
	var payload T
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&payload); err != nil {
		return payload, err
	}
	if err := decoder.Decode(new(any)); err != nil {
		if errors.Is(err, io.EOF) {
			return payload, nil
		}
		return payload, err
	}
	return payload, errMultipleJSONBody
}
