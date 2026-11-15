package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/leetun2k2/go-bedrock/logx"
	"github.com/leetun2k2/go-product/internal/domain"
	"github.com/leetun2k2/go-product/internal/repository"
)

type CatalogService struct {
	repo   *repository.Repository
	logger *logx.Logger
}

func NewCatalogService(repo *repository.Repository, logger *logx.Logger) *CatalogService {
	return &CatalogService{
		repo:   repo,
		logger: logger,
	}
}

// GetCatalog fulfills the Website catalog requirement:
// Sends product_id and currency. Product Service reads visible price rows in a single query;
// catalog_data contains plan, offer, and quota. Free and paid share the same response shape.
func (s *CatalogService) GetCatalog(ctx context.Context, productID uuid.UUID, currency string) ([]domain.CatalogItem, error) {
	normCurrency := strings.ToUpper(strings.TrimSpace(currency))
	if len(normCurrency) != 3 {
		return nil, domain.ErrInvalidCurrency
	}

	items, err := s.repo.ListCatalogPrices(ctx, nil, productID, normCurrency)
	if err != nil {
		return nil, fmt.Errorf("get catalog: %w", err)
	}

	return items, nil
}

// GetPaymentPrice fulfills the Payment lookup requirement:
// Sends offer_id and currency to fetch the canonical amount. Payment does not trust client-supplied prices.
// Returns an error if the offer lacks a price or if the offer/plan/product is inactive.
func (s *CatalogService) GetPaymentPrice(ctx context.Context, offerID uuid.UUID, currency string) (*domain.Price, error) {
	normCurrency := strings.ToUpper(strings.TrimSpace(currency))
	if len(normCurrency) != 3 {
		return nil, domain.ErrInvalidCurrency
	}

	price, err := s.repo.GetPriceByOfferAndCurrency(ctx, nil, offerID, normCurrency)
	if err != nil {
		return nil, err
	}

	if price.Status != domain.PriceStatusActive {
		return nil, domain.ErrPriceNotFound
	}

	offer, err := s.repo.GetOfferByID(ctx, nil, price.OfferID)
	if err != nil {
		return nil, err
	}
	if offer.Status != domain.OfferStatusActive {
		return nil, domain.ErrOfferNotFound
	}

	plan, err := s.repo.GetPlanByID(ctx, nil, offer.PlanID)
	if err != nil {
		return nil, err
	}
	if plan.Status != domain.PlanStatusActive {
		return nil, domain.ErrPlanNotFound
	}

	product, err := s.repo.GetProductByID(ctx, nil, plan.ProductID)
	if err != nil {
		return nil, err
	}
	if product.Status != domain.ProductStatusActive {
		return nil, domain.ErrProductNotFound
	}

	return price, nil
}

// GetSubscriptionPlan fulfills the Subscription lookup requirement:
// Reads plan by ID to grant entitlements after purchase or free registration.
// Inactive plans remain readable internally to service existing entitlements.
func (s *CatalogService) GetSubscriptionPlan(ctx context.Context, planID uuid.UUID) (*domain.Plan, []domain.QuotaRule, error) {
	plan, err := s.repo.GetPlanByID(ctx, nil, planID)
	if err != nil {
		return nil, nil, err
	}

	quotas, err := s.repo.ListQuotaRulesByPlanID(ctx, nil, planID)
	if err != nil {
		return nil, nil, fmt.Errorf("list quota rules for plan: %w", err)
	}

	return plan, quotas, nil
}

// GetUsageQuotas fulfills the Usage lookup requirement:
// Reads all quota rules for a plan to configure usage counters.
func (s *CatalogService) GetUsageQuotas(ctx context.Context, planID uuid.UUID) ([]domain.QuotaRule, error) {
	_, err := s.repo.GetPlanByID(ctx, nil, planID)
	if err != nil {
		return nil, err
	}

	quotas, err := s.repo.ListQuotaRulesByPlanID(ctx, nil, planID)
	if err != nil {
		return nil, fmt.Errorf("list quota rules for plan: %w", err)
	}

	return quotas, nil
}
