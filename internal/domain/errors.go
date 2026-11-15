package domain

import (
	"errors"
)

var (
	ErrProductNotFound       = errors.New("product not found")
	ErrPlanNotFound          = errors.New("plan not found")
	ErrOfferNotFound         = errors.New("offer not found")
	ErrPriceNotFound         = errors.New("price not found")
	ErrQuotaNotFound         = errors.New("quota rule not found")
	ErrPlanImmutable         = errors.New("plan entitlements and quotas are immutable once active")
	ErrPlanActivationInvalid = errors.New("active plan requires at least one active offer and each active offer requires an active price")
	ErrInvalidOfferType      = errors.New("invalid offer type")
	ErrInvalidBillingCycle   = errors.New("invalid billing cycle for offer type")
	ErrInvalidAnchorType     = errors.New("invalid anchor type")
	ErrInvalidAmount         = errors.New("price amount must be 0 for free offers and greater than 0 for paid offers")
	ErrInvalidQuotaValues    = errors.New("limit_amount and window_duration_seconds must be greater than 0")
	ErrHierarchyMismatch     = errors.New("ancestor IDs do not match hierarchy")
	ErrConflict              = errors.New("resource already exists or constraint violation")
	ErrInvalidCurrency       = errors.New("invalid currency code")
)
