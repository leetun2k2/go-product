package domain

import (
	"time"

	"github.com/google/uuid"
)

type PlanStatus string

const (
	PlanStatusDraft    PlanStatus = "draft"
	PlanStatusActive   PlanStatus = "active"
	PlanStatusInactive PlanStatus = "inactive"
)

func (s PlanStatus) IsValid() bool {
	return s == PlanStatusDraft || s == PlanStatusActive || s == PlanStatusInactive
}

type Plan struct {
	ID          uuid.UUID  `json:"id"`
	ProductID   uuid.UUID  `json:"product_id"`
	Code        string     `json:"code"`
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
	Status      PlanStatus `json:"status"`
	Version     int        `json:"version"`
	ActivatedAt *time.Time `json:"activated_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
