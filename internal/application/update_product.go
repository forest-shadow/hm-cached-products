package application

import "github.com/forest-shadow/hm-cached-products/internal/product"

type UpdateProductCommand struct {
	productID       int64
	expectedVersion int64
	input           product.Input
}

func NewUpdateProductCommand(
	productID int64,
	expectedVersion int64,
	input product.Input,
) (UpdateProductCommand, error) {
	if productID <= 0 || expectedVersion <= 0 {
		return UpdateProductCommand{}, product.ErrInvalid
	}

	clean, err := product.Validate(input)
	if err != nil {
		return UpdateProductCommand{}, err
	}

	return UpdateProductCommand{
		productID:       productID,
		expectedVersion: expectedVersion,
		input:           clean,
	}, nil
}

func (c UpdateProductCommand) ProductID() int64 {
	return c.productID
}

func (c UpdateProductCommand) ExpectedVersion() int64 {
	return c.expectedVersion
}

func (c UpdateProductCommand) Input() product.Input {
	return c.input
}

func (c UpdateProductCommand) validate() error {
	if c.productID <= 0 || c.expectedVersion <= 0 {
		return product.ErrInvalid
	}
	_, err := product.Validate(c.input)
	return err
}
