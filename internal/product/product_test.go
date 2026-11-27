
package product

import (
	"errors"
	"testing"
)

func TestValidate(t *testing.T) {
	invalidInputs := []struct {
		name  string
		input Input
	}{
		{name: "empty name", input: Input{}},
		{name: "whitespace-only name", input: Input{Name: " ", Price: 1}},
		{name: "negative price", input: Input{Name: "x", Price: -1}},
		{name: "name exceeds limit", input: Input{Name: string(make([]byte, 201)), Price: 1}},
	}
	for _, tc := range invalidInputs {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Validate(tc.input)
			if !errors.Is(err, ErrInvalid) {
				t.Fatalf("Validate() error = %v, want %v", err, ErrInvalid)
			}
		})
	}
	t.Run("trims name and accepts zero price", func(t *testing.T) {
		clean, err := Validate(Input{Name: " Tea ", Price: 0})
		if err != nil {
			t.Fatalf("Validate() error = %v", err)
		}
		if clean.Name != "Tea" || clean.Price != 0 {
			t.Fatalf("Validate() = %+v, want name Tea and price 0", clean)
		}
	})
}
