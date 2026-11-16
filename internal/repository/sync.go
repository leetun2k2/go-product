package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/leetun2k2/go-product/internal/domain"
)

func (r *Repository) RebuildPrice(ctx context.Context, tx pgx.Tx, priceID uuid.UUID) error {
	price, err := r.GetPriceByID(ctx, tx, priceID)
	if err != nil {
		return fmt.Errorf("get price for rebuild: %w", err)
	}

	offer, err := r.GetOfferByID(ctx, tx, price.OfferID)
	if err != nil {
		return fmt.Errorf("get offer for rebuild: %w", err)
	}

	plan, err := r.GetPlanByID(ctx, tx, offer.PlanID)
	if err != nil {
		return fmt.Errorf("get plan for rebuild: %w", err)
	}

	product, err := r.GetProductByID(ctx, tx, plan.ProductID)
	if err != nil {
		return fmt.Errorf("get product for rebuild: %w", err)
	}

	quotas, err := r.ListQuotaRulesByPlanID(ctx, tx, plan.ID)
	if err != nil {
		return fmt.Errorf("list quotas for rebuild: %w", err)
	}

	quotaSnapshots := make([]domain.CatalogQuotaSnapshot, 0, len(quotas))
	for _, q := range quotas {
		quotaSnapshots = append(quotaSnapshots, domain.CatalogQuotaSnapshot{
			ID:                    q.ID,
			RuleKey:               q.RuleKey,
			LimitAmount:           q.LimitAmount,
			WindowDurationSeconds: q.WindowDurationSeconds,
			AnchorType:            q.AnchorType,
		})
	}

	catalogVisible := product.Status == domain.ProductStatusActive &&
		plan.Status == domain.PlanStatusActive &&
		offer.Status == domain.OfferStatusActive &&
		price.Status == domain.PriceStatusActive

	catalogData := domain.CatalogData{
		Product: domain.CatalogProductSnapshot{
			ID:          product.ID,
			Code:        product.Code,
			Name:        product.Name,
			Description: product.Description,
		},
		Plan: domain.CatalogPlanSnapshot{
			ID:          plan.ID,
			Code:        plan.Code,
			Name:        plan.Name,
			Description: plan.Description,
			Version:     plan.Version,
		},
		Offer: domain.CatalogOfferSnapshot{
			ID:           offer.ID,
			Code:         offer.Code,
			OfferType:    offer.OfferType,
			BillingCycle: offer.BillingCycle,
			SortOrder:    offer.SortOrder,
		},
		Quotas: quotaSnapshots,
	}

	encoded, err := json.Marshal(catalogData)
	if err != nil {
		return fmt.Errorf("marshal catalog data: %w", err)
	}

	qUpdate := `
		UPDATE product.prices
		SET product_id = $1, plan_id = $2, catalog_visible = $3, catalog_data = $4, updated_at = now()
		WHERE id = $5;
	`
	exec := r.getExecutor(tx)
	if _, err := exec.Exec(ctx, qUpdate, product.ID, plan.ID, catalogVisible, encoded, price.ID); err != nil {
		return fmt.Errorf("update price read model: %w", err)
	}

	return nil
}

func (r *Repository) RebuildPricesForOffer(ctx context.Context, tx pgx.Tx, offerID uuid.UUID) error {
	prices, err := r.ListPricesByOfferID(ctx, tx, offerID)
	if err != nil {
		return err
	}
	for _, p := range prices {
		if err := r.RebuildPrice(ctx, tx, p.ID); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) RebuildPricesForPlan(ctx context.Context, tx pgx.Tx, planID uuid.UUID) error {
	prices, err := r.ListPricesByPlanID(ctx, tx, planID)
	if err != nil {
		return err
	}
	for _, p := range prices {
		if err := r.RebuildPrice(ctx, tx, p.ID); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) RebuildPricesForProduct(ctx context.Context, tx pgx.Tx, productID uuid.UUID) error {
	prices, err := r.ListPricesByProductID(ctx, tx, productID)
	if err != nil {
		return err
	}
	for _, p := range prices {
		if err := r.RebuildPrice(ctx, tx, p.ID); err != nil {
			return err
		}
	}
	return nil
}
