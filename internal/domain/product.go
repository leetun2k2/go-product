package domain

import (
	"time"

	"github.com/google/uuid"
)

type ProductStatus string

const (
	ProductStatusActive   ProductStatus = "active"
	ProductStatusInactive ProductStatus = "inactive"
)

func (s ProductStatus) IsValid() bool {
	return s == ProductStatusActive || s == ProductStatusInactive
}

type Product struct {
	ID          uuid.UUID     `json:"id"`
	Code        string        `json:"code"`
	Name        string        `json:"name"`
	Description *string       `json:"description,omitempty"`
	Status      ProductStatus `json:"status"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}
