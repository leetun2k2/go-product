package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type PriceStatus string

const (
	PriceStatusActive   PriceStatus = "active"
	PriceStatusInactive PriceStatus = "inactive"
)

func (s PriceStatus) IsValid() bool {
	return s == PriceStatusActive || s == PriceStatusInactive
}

type CatalogProductSnapshot struct {
	ID          uuid.UUID `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
}

type CatalogPlanSnapshot struct {
	ID          uuid.UUID `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	Version     int       `json:"version"`
}

type CatalogOfferSnapshot struct {
	ID           uuid.UUID    `json:"id"`
	Code         string       `json:"code"`
	OfferType    OfferType    `json:"offer_type"`
	BillingCycle BillingCycle `json:"billing_cycle"`
	SortOrder    int          `json:"sort_order"`
}

type CatalogQuotaSnapshot struct {
	ID                    uuid.UUID  `json:"id"`
	RuleKey               string     `json:"rule_key"`
	LimitAmount           int64      `json:"limit_amount"`
	WindowDurationSeconds int        `json:"window_duration_seconds"`
	AnchorType            AnchorType `json:"anchor_type"`
}

type CatalogData struct {
	Product CatalogProductSnapshot `json:"product"`
	Plan    CatalogPlanSnapshot    `json:"plan"`
	Offer   CatalogOfferSnapshot   `json:"offer"`
	Quotas  []CatalogQuotaSnapshot `json:"quotas"`
}

type Price struct {
	ID             uuid.UUID       `json:"id"`
	OfferID        uuid.UUID       `json:"offer_id"`
	ProductID      uuid.UUID       `json:"product_id"`
	PlanID         uuid.UUID       `json:"plan_id"`
	Currency       string          `json:"currency"`
	Amount         int64           `json:"amount"`
	Status         PriceStatus     `json:"status"`
	CatalogVisible bool            `json:"catalog_visible"`
	CatalogData    json.RawMessage `json:"catalog_data"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

type CatalogItem struct {
	PriceID     uuid.UUID       `json:"price_id"`
	OfferID     uuid.UUID       `json:"offer_id"`
	PlanID      uuid.UUID       `json:"plan_id"`
	Currency    string          `json:"currency"`
	Amount      int64           `json:"amount"`
	CatalogData json.RawMessage `json:"catalog_data"`
}
