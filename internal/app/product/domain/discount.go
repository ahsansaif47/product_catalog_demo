package domain

import (
	"time"
)

// Discount represents a percentage-based discount with a validity period
type Discount struct {
	percentage int64
	startDate  time.Time
	endDate    time.Time
}

// NewDiscount creates a new Discount value object
func NewDiscount(percentage int64, startDate, endDate time.Time) (*Discount, error) {
	if percentage < 0 || percentage > 100 {
		return nil, ErrDiscountOutOfRange
	}

	if endDate.Before(startDate) {
		return nil, ErrInvalidDateRange
	}

	return &Discount{
		percentage: percentage,
		startDate:  startDate,
		endDate:    endDate,
	}, nil
}

// Percentage returns the discount percentage
func (d *Discount) Percentage() int64 {
	if d == nil {
		return 0
	}
	return d.percentage
}

// StartDate returns when the discount becomes active
func (d *Discount) StartDate() time.Time {
	if d == nil {
		return time.Time{}
	}
	return d.startDate
}

// EndDate returns when the discount expires
func (d *Discount) EndDate() time.Time {
	if d == nil {
		return time.Time{}
	}
	return d.endDate
}

// IsActiveAt checks if the discount is active at the given time
func (d *Discount) IsActiveAt(t time.Time) bool {
	if d == nil {
		return false
	}

	return !t.Before(d.startDate) && !t.After(d.endDate)
}

// IsValidAt checks if the discount period is valid for a given time
// This is used when applying a discount to ensure it covers the current time
func (d *Discount) IsValidAt(now time.Time) bool {
	if d == nil {
		return false
	}

	// Discount must be currently active or start in the future
	return now.Before(d.endDate) || now.Equal(d.endDate)
}

// Equals checks if two discounts are equal
func (d *Discount) Equals(other *Discount) bool {
	if d == nil && other == nil {
		return true
	}
	if d == nil || other == nil {
		return false
	}

	return d.percentage == other.percentage &&
		d.startDate.Equal(other.startDate) &&
		d.endDate.Equal(other.endDate)
}
