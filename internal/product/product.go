package product

import (
	"errors"
	"strings"
	"time"
)

var ErrInvalid = errors.New("invalid product")

type Product struct {
	ID        int64
	Name      string
	Price     int64
	Version   int64
	UpdatedAt time.Time
}

type Input struct {
	Name  string
	Price int64
}

func Validate(input Input) (Input, error) {
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" || len(input.Name) > 200 || input.Price < 0 {
		return Input{}, ErrInvalid
	}
	return input, nil
}
