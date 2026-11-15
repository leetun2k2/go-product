package domain

import (
	"time"

	"github.com/google/uuid"
)

type AnchorType string

const (
	AnchorTypeFirstUse         AnchorType = "FIRST_USE"
	AnchorTypeSubscriptionDate AnchorType = "SUBSCRIPTION_DATE"
)

func (a AnchorType) IsValid() bool {
	return a == AnchorTypeFirstUse || a == AnchorTypeSubscriptionDate
}

type QuotaRule struct {
	ID                    uuid.UUID  `json:"id"`
	ProductID             uuid.UUID  `json:"product_id"`
	PlanID                uuid.UUID  `json:"plan_id"`
	RuleKey               string     `json:"rule_key"`
	LimitAmount           int64      `json:"limit_amount"`
	WindowDurationSeconds int        `json:"window_duration_seconds"`
	AnchorType            AnchorType `json:"anchor_type"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}
