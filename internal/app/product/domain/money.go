package domain

import (
	"math/big"
)

// Money represents a monetary value using rational numbers for precise calculations
type Money struct {
	value *big.Rat
}

// NewMoney creates a new Money value from numerator and denominator
func NewMoney(numerator, denominator int64) (*Money, error) {
	if denominator == 0 {
		return nil, ErrInvalidPrice
	}

	rat := big.NewRat(numerator, denominator)
	if rat.Sign() <= 0 {
		return nil, ErrInvalidPrice
	}

	return &Money{value: rat}, nil
}

// Value returns the underlying big.Rat value
func (m *Money) Value() *big.Rat {
	if m == nil {
		return nil
	}
	return m.value
}

// Numerator returns the numerator of the rational number
func (m *Money) Numerator() int64 {
	if m == nil {
		return 0
	}
	return m.value.Num().Int64()
}

// Denominator returns the denominator of the rational number
func (m *Money) Denominator() int64 {
	if m == nil {
		return 0
	}
	return m.value.Denom().Int64()
}

// ApplyPercentage applies a percentage discount and returns the discounted amount
// For example, applying 20% to $100 returns $80
func (m *Money) ApplyPercentage(percentage int64) (*Money, error) {
	if m == nil {
		return nil, ErrInvalidPrice
	}

	if percentage < 0 || percentage > 100 {
		return nil, ErrDiscountOutOfRange
	}

	// Calculate discount amount: price * (percentage / 100)
	discount := new(big.Rat).Mul(m.value, big.NewRat(percentage, 100))

	// Subtract discount from original price
	finalPrice := new(big.Rat).Sub(m.value, discount)

	return &Money{value: finalPrice}, nil
}

// Add adds another Money value to this one
func (m *Money) Add(other *Money) (*Money, error) {
	if m == nil || other == nil {
		return nil, ErrInvalidPrice
	}

	sum := new(big.Rat).Add(m.value, other.value)
	return &Money{value: sum}, nil
}

// Equals checks if two Money values are equal
func (m *Money) Equals(other *Money) bool {
	if m == nil || other == nil {
		return m == nil && other == nil
	}
	return m.value.Cmp(other.value) == 0
}

// GreaterThan checks if this Money is greater than other
func (m *Money) GreaterThan(other *Money) bool {
	if m == nil {
		return false
	}
	if other == nil {
		return true
	}
	return m.value.Cmp(other.value) > 0
}

// String returns the decimal representation
func (m *Money) String() string {
	if m == nil {
		return "0"
	}
	f, _ := m.value.Float64()
	return formatMoney(f)
}

func formatMoney(f float64) string {
	// Simple formatting - in production you'd want proper decimal formatting
	return string(rune(f))
}
