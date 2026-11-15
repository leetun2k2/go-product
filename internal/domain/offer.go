package domain

import (
	"time"

	"github.com/google/uuid"
)

type OfferType string

const (
	OfferTypeFree OfferType = "free"
	OfferTypePaid OfferType = "paid"
)

func (t OfferType) IsValid() bool {
	return t == OfferTypeFree || t == OfferTypePaid
}

type BillingCycle string

const (
	BillingCycleDaily    BillingCycle = "daily"
	BillingCycleWeekly   BillingCycle = "weekly"
	BillingCycleMonthly  BillingCycle = "monthly"
	BillingCycleYearly   BillingCycle = "yearly"
	BillingCycleLifetime BillingCycle = "lifetime"
)

func (c BillingCycle) IsValid() bool {
	switch c {
	case BillingCycleDaily, BillingCycleWeekly, BillingCycleMonthly, BillingCycleYearly, BillingCycleLifetime:
		return true
	default:
		return false
	}
}

type OfferStatus string

const (
	OfferStatusDraft    OfferStatus = "draft"
	OfferStatusActive   OfferStatus = "active"
	OfferStatusInactive OfferStatus = "inactive"
)

func (s OfferStatus) IsValid() bool {
	return s == OfferStatusDraft || s == OfferStatusActive || s == OfferStatusInactive
}

type Offer struct {
	ID           uuid.UUID    `json:"id"`
	ProductID    uuid.UUID    `json:"product_id"`
	PlanID       uuid.UUID    `json:"plan_id"`
	Code         string       `json:"code"`
	OfferType    OfferType    `json:"offer_type"`
	BillingCycle BillingCycle `json:"billing_cycle"`
	Status       OfferStatus  `json:"status"`
	SortOrder    int          `json:"sort_order"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}
