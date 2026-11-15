package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/leetun2k2/go-bedrock/logx"
	"github.com/leetun2k2/go-product/internal/domain"
	"github.com/leetun2k2/go-product/internal/repository"
)

type AdminService struct {
	repo   *repository.Repository
	logger *logx.Logger
}

func NewAdminService(repo *repository.Repository, logger *logx.Logger) *AdminService {
	return &AdminService{
		repo:   repo,
		logger: logger,
	}
}

// --- Product Operations ---

func (s *AdminService) CreateProduct(ctx context.Context, code, name string, description *string) (*domain.Product, error) {
	code = strings.TrimSpace(code)
	name = strings.TrimSpace(name)
	if code == "" || name == "" {
		return nil, fmt.Errorf("code and name are required")
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("generate product id: %w", err)
	}

	now := time.Now().UTC()
	product := &domain.Product{
		ID:          id,
		Code:        code,
		Name:        name,
		Description: description,
		Status:      domain.ProductStatusActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.CreateProduct(ctx, nil, product); err != nil {
		return nil, err
	}
	return product, nil
}

func (s *AdminService) UpdateProduct(ctx context.Context, id uuid.UUID, name string, description *string, status *domain.ProductStatus) (*domain.Product, error) {
	var updated *domain.Product
	err := s.repo.WithTx(ctx, func(tx pgx.Tx) error {
		product, err := s.repo.GetProductByID(ctx, tx, id)
		if err != nil {
			return err
		}

		if strings.TrimSpace(name) != "" {
			product.Name = strings.TrimSpace(name)
		}
		if description != nil {
			product.Description = description
		}
		statusChanged := false
		if status != nil && status.IsValid() && *status != product.Status {
			product.Status = *status
			statusChanged = true
		}
		product.UpdatedAt = time.Now().UTC()

		if err := s.repo.UpdateProduct(ctx, tx, product); err != nil {
			return err
		}

		if statusChanged {
			if err := s.repo.RebuildPricesForProduct(ctx, tx, product.ID); err != nil {
				return fmt.Errorf("rebuild prices for product status change: %w", err)
			}
		}

		updated = product
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *AdminService) GetProduct(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	return s.repo.GetProductByID(ctx, nil, id)
}

func (s *AdminService) GetProductByCode(ctx context.Context, code string) (*domain.Product, error) {
	return s.repo.GetProductByCode(ctx, nil, code)
}

func (s *AdminService) ListProducts(ctx context.Context) ([]domain.Product, error) {
	return s.repo.ListProducts(ctx, nil)
}

// --- Plan Operations ---

func (s *AdminService) CreatePlan(ctx context.Context, productID uuid.UUID, code, name string, description *string, version int) (*domain.Plan, error) {
	code = strings.TrimSpace(code)
	name = strings.TrimSpace(name)
	if code == "" || name == "" {
		return nil, fmt.Errorf("code and name are required")
	}
	if version <= 0 {
		version = 1
	}

	if _, err := s.repo.GetProductByID(ctx, nil, productID); err != nil {
		return nil, err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("generate plan id: %w", err)
	}

	now := time.Now().UTC()
	plan := &domain.Plan{
		ID:          id,
		ProductID:   productID,
		Code:        code,
		Name:        name,
		Description: description,
		Status:      domain.PlanStatusDraft,
		Version:     version,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.CreatePlan(ctx, nil, plan); err != nil {
		return nil, err
	}
	return plan, nil
}

func (s *AdminService) UpdatePlan(ctx context.Context, id uuid.UUID, name string, description *string) (*domain.Plan, error) {
	plan, err := s.repo.GetPlanByID(ctx, nil, id)
	if err != nil {
		return nil, err
	}

	// Plan entitlements and quotas can only be edited in draft status.
	// Once activated or inactivated, entitlements and quotas are strictly immutable.
	if plan.Status != domain.PlanStatusDraft {
		return nil, domain.ErrPlanImmutable
	}

	if strings.TrimSpace(name) != "" {
		plan.Name = strings.TrimSpace(name)
	}
	if description != nil {
		plan.Description = description
	}
	plan.UpdatedAt = time.Now().UTC()

	if err := s.repo.UpdatePlan(ctx, nil, plan); err != nil {
		return nil, err
	}
	return plan, nil
}

// ActivatePlan validates all active offers, prices, and amounts, then updates status and read model atomically.
func (s *AdminService) ActivatePlan(ctx context.Context, id uuid.UUID) (*domain.Plan, error) {
	var activated *domain.Plan
	err := s.repo.WithTx(ctx, func(tx pgx.Tx) error {
		plan, err := s.repo.GetPlanByID(ctx, tx, id)
		if err != nil {
			return err
		}

		// 1. Must have at least one active offer
		activeOffers, err := s.repo.ListActiveOffersByPlanID(ctx, tx, plan.ID)
		if err != nil {
			return fmt.Errorf("list active offers: %w", err)
		}
		if len(activeOffers) == 0 {
			return fmt.Errorf("%w: plan has no active offers", domain.ErrPlanActivationInvalid)
		}

		// 2. Each active offer must have at least one active price
		for _, offer := range activeOffers {
			prices, err := s.repo.ListPricesByOfferID(ctx, tx, offer.ID)
			if err != nil {
				return fmt.Errorf("list prices for offer: %w", err)
			}
			activePriceCount := 0
			for _, p := range prices {
				if p.Status != domain.PriceStatusActive {
					continue
				}
				activePriceCount++

				// Validate pricing rules
				if offer.OfferType == domain.OfferTypeFree {
					if offer.BillingCycle != domain.BillingCycleLifetime {
						return fmt.Errorf("%w: free offer %s must use lifetime billing cycle", domain.ErrPlanActivationInvalid, offer.Code)
					}
					if p.Amount != 0 {
						return fmt.Errorf("%w: free offer price must be 0", domain.ErrPlanActivationInvalid)
					}
				} else if offer.OfferType == domain.OfferTypePaid {
					if p.Amount <= 0 {
						return fmt.Errorf("%w: paid offer price must be greater than 0", domain.ErrPlanActivationInvalid)
					}
				}
			}

			if activePriceCount == 0 {
				return fmt.Errorf("%w: offer %s has no active prices", domain.ErrPlanActivationInvalid, offer.Code)
			}
		}

		// 3. Update plan status
		if err := s.repo.UpdatePlanStatus(ctx, tx, plan.ID, domain.PlanStatusActive); err != nil {
			return fmt.Errorf("update plan status: %w", err)
		}

		// 4. Rebuild read model for all prices of this plan
		if err := s.repo.RebuildPricesForPlan(ctx, tx, plan.ID); err != nil {
			return fmt.Errorf("rebuild prices for plan: %w", err)
		}

		plan.Status = domain.PlanStatusActive
		plan.UpdatedAt = time.Now().UTC()
		activated = plan
		return nil
	})
	if err != nil {
		return nil, err
	}
	return activated, nil
}

func (s *AdminService) DeactivatePlan(ctx context.Context, id uuid.UUID) (*domain.Plan, error) {
	var deactivated *domain.Plan
	err := s.repo.WithTx(ctx, func(tx pgx.Tx) error {
		plan, err := s.repo.GetPlanByID(ctx, tx, id)
		if err != nil {
			return err
		}

		if err := s.repo.UpdatePlanStatus(ctx, tx, plan.ID, domain.PlanStatusInactive); err != nil {
			return fmt.Errorf("update plan status: %w", err)
		}

		// Rebuilding prices will automatically update catalog_visible to false
		if err := s.repo.RebuildPricesForPlan(ctx, tx, plan.ID); err != nil {
			return fmt.Errorf("rebuild prices for plan: %w", err)
		}

		plan.Status = domain.PlanStatusInactive
		plan.UpdatedAt = time.Now().UTC()
		deactivated = plan
		return nil
	})
	if err != nil {
		return nil, err
	}
	return deactivated, nil
}

func (s *AdminService) GetPlan(ctx context.Context, id uuid.UUID) (*domain.Plan, error) {
	return s.repo.GetPlanByID(ctx, nil, id)
}

func (s *AdminService) ListPlansByProductID(ctx context.Context, productID uuid.UUID) ([]domain.Plan, error) {
	return s.repo.ListPlansByProductID(ctx, nil, productID)
}

// --- Offer Operations ---

func (s *AdminService) CreateOffer(ctx context.Context, planID uuid.UUID, code string, offerType domain.OfferType, billingCycle domain.BillingCycle, sortOrder int) (*domain.Offer, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, fmt.Errorf("offer code is required")
	}
	if !offerType.IsValid() {
		return nil, domain.ErrInvalidOfferType
	}
	if !billingCycle.IsValid() {
		return nil, domain.ErrInvalidBillingCycle
	}

	// Free offer must use lifetime billing cycle
	if offerType == domain.OfferTypeFree && billingCycle != domain.BillingCycleLifetime {
		return nil, fmt.Errorf("%w: free offer must have lifetime billing cycle", domain.ErrInvalidBillingCycle)
	}

	plan, err := s.repo.GetPlanByID(ctx, nil, planID)
	if err != nil {
		return nil, err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("generate offer id: %w", err)
	}

	now := time.Now().UTC()
	offer := &domain.Offer{
		ID:           id,
		ProductID:    plan.ProductID,
		PlanID:       plan.ID,
		Code:         code,
		OfferType:    offerType,
		BillingCycle: billingCycle,
		Status:       domain.OfferStatusDraft,
		SortOrder:    sortOrder,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.repo.CreateOffer(ctx, nil, offer); err != nil {
		return nil, err
	}
	return offer, nil
}

func (s *AdminService) UpdateOffer(ctx context.Context, id uuid.UUID, code string, sortOrder int) (*domain.Offer, error) {
	offer, err := s.repo.GetOfferByID(ctx, nil, id)
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(code) != "" {
		offer.Code = strings.TrimSpace(code)
	}
	offer.SortOrder = sortOrder
	offer.UpdatedAt = time.Now().UTC()

	var updated *domain.Offer
	err = s.repo.WithTx(ctx, func(tx pgx.Tx) error {
		if err := s.repo.UpdateOffer(ctx, tx, offer); err != nil {
			return err
		}
		if err := s.repo.RebuildPricesForOffer(ctx, tx, offer.ID); err != nil {
			return fmt.Errorf("rebuild prices for offer: %w", err)
		}
		updated = offer
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *AdminService) ActivateOffer(ctx context.Context, id uuid.UUID) (*domain.Offer, error) {
	var activated *domain.Offer
	err := s.repo.WithTx(ctx, func(tx pgx.Tx) error {
		offer, err := s.repo.GetOfferByID(ctx, tx, id)
		if err != nil {
			return err
		}

		if err := s.repo.UpdateOfferStatus(ctx, tx, offer.ID, domain.OfferStatusActive); err != nil {
			return fmt.Errorf("update offer status: %w", err)
		}

		if err := s.repo.RebuildPricesForOffer(ctx, tx, offer.ID); err != nil {
			return fmt.Errorf("rebuild prices for offer: %w", err)
		}

		offer.Status = domain.OfferStatusActive
		offer.UpdatedAt = time.Now().UTC()
		activated = offer
		return nil
	})
	if err != nil {
		return nil, err
	}
	return activated, nil
}

func (s *AdminService) DeactivateOffer(ctx context.Context, id uuid.UUID) (*domain.Offer, error) {
	var deactivated *domain.Offer
	err := s.repo.WithTx(ctx, func(tx pgx.Tx) error {
		offer, err := s.repo.GetOfferByID(ctx, tx, id)
		if err != nil {
			return err
		}

		if err := s.repo.UpdateOfferStatus(ctx, tx, offer.ID, domain.OfferStatusInactive); err != nil {
			return fmt.Errorf("update offer status: %w", err)
		}

		if err := s.repo.RebuildPricesForOffer(ctx, tx, offer.ID); err != nil {
			return fmt.Errorf("rebuild prices for offer: %w", err)
		}

		offer.Status = domain.OfferStatusInactive
		offer.UpdatedAt = time.Now().UTC()
		deactivated = offer
		return nil
	})
	if err != nil {
		return nil, err
	}
	return deactivated, nil
}

func (s *AdminService) GetOffer(ctx context.Context, id uuid.UUID) (*domain.Offer, error) {
	return s.repo.GetOfferByID(ctx, nil, id)
}

func (s *AdminService) ListOffersByPlanID(ctx context.Context, planID uuid.UUID) ([]domain.Offer, error) {
	return s.repo.ListOffersByPlanID(ctx, nil, planID)
}

// --- Price Operations ---

func (s *AdminService) UpsertPrice(ctx context.Context, offerID uuid.UUID, currency string, amount int64, status domain.PriceStatus) (*domain.Price, error) {
	normCurrency := strings.ToUpper(strings.TrimSpace(currency))
	if len(normCurrency) != 3 {
		return nil, domain.ErrInvalidCurrency
	}
	if !status.IsValid() {
		status = domain.PriceStatusActive
	}

	var saved *domain.Price
	err := s.repo.WithTx(ctx, func(tx pgx.Tx) error {
		offer, err := s.repo.GetOfferByID(ctx, tx, offerID)
		if err != nil {
			return err
		}

		// Amount conventions
		if offer.OfferType == domain.OfferTypeFree && amount != 0 {
			return domain.ErrInvalidAmount
		}
		if offer.OfferType == domain.OfferTypePaid && amount <= 0 {
			return domain.ErrInvalidAmount
		}

		id, err := uuid.NewV7()
		if err != nil {
			return fmt.Errorf("generate price id: %w", err)
		}

		now := time.Now().UTC()
		price := &domain.Price{
			ID:             id,
			OfferID:        offer.ID,
			ProductID:      offer.ProductID,
			PlanID:         offer.PlanID,
			Currency:       normCurrency,
			Amount:         amount,
			Status:         status,
			CatalogVisible: false,
			CatalogData:    []byte("{}"),
			CreatedAt:      now,
			UpdatedAt:      now,
		}

		if err := s.repo.UpsertPrice(ctx, tx, price); err != nil {
			return err
		}

		// Atomically rebuild read model for this price
		if err := s.repo.RebuildPrice(ctx, tx, price.ID); err != nil {
			return fmt.Errorf("rebuild price read model: %w", err)
		}

		refetched, err := s.repo.GetPriceByID(ctx, tx, price.ID)
		if err != nil {
			return err
		}
		saved = refetched
		return nil
	})
	if err != nil {
		return nil, err
	}
	return saved, nil
}

func (s *AdminService) UpdatePriceStatus(ctx context.Context, id uuid.UUID, status domain.PriceStatus) (*domain.Price, error) {
	if !status.IsValid() {
		return nil, fmt.Errorf("invalid price status")
	}

	var updated *domain.Price
	err := s.repo.WithTx(ctx, func(tx pgx.Tx) error {
		if err := s.repo.UpdatePriceStatus(ctx, tx, id, status); err != nil {
			return err
		}
		if err := s.repo.RebuildPrice(ctx, tx, id); err != nil {
			return fmt.Errorf("rebuild price read model: %w", err)
		}

		p, err := s.repo.GetPriceByID(ctx, tx, id)
		if err != nil {
			return err
		}
		updated = p
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *AdminService) GetPrice(ctx context.Context, id uuid.UUID) (*domain.Price, error) {
	return s.repo.GetPriceByID(ctx, nil, id)
}

func (s *AdminService) ListPricesByOfferID(ctx context.Context, offerID uuid.UUID) ([]domain.Price, error) {
	return s.repo.ListPricesByOfferID(ctx, nil, offerID)
}

// --- Quota Rule Operations ---

func (s *AdminService) CreateQuotaRule(ctx context.Context, planID uuid.UUID, ruleKey string, limitAmount int64, windowDurationSeconds int, anchorType domain.AnchorType) (*domain.QuotaRule, error) {
	ruleKey = strings.TrimSpace(ruleKey)
	if ruleKey == "" {
		return nil, fmt.Errorf("rule_key is required")
	}
	if limitAmount <= 0 || windowDurationSeconds <= 0 {
		return nil, domain.ErrInvalidQuotaValues
	}
	if !anchorType.IsValid() {
		return nil, domain.ErrInvalidAnchorType
	}

	var created *domain.QuotaRule
	err := s.repo.WithTx(ctx, func(tx pgx.Tx) error {
		plan, err := s.repo.GetPlanByID(ctx, tx, planID)
		if err != nil {
			return err
		}

		// Plan entitlements and quotas can only be edited in draft status.
		// Once activated or inactivated, entitlements and quotas are strictly immutable.
		if plan.Status != domain.PlanStatusDraft {
			return domain.ErrPlanImmutable
		}

		id, err := uuid.NewV7()
		if err != nil {
			return fmt.Errorf("generate quota id: %w", err)
		}

		now := time.Now().UTC()
		quota := &domain.QuotaRule{
			ID:                    id,
			ProductID:             plan.ProductID,
			PlanID:                plan.ID,
			RuleKey:               ruleKey,
			LimitAmount:           limitAmount,
			WindowDurationSeconds: windowDurationSeconds,
			AnchorType:            anchorType,
			CreatedAt:             now,
			UpdatedAt:             now,
		}

		if err := s.repo.CreateQuotaRule(ctx, tx, quota); err != nil {
			return err
		}

		if err := s.repo.RebuildPricesForPlan(ctx, tx, plan.ID); err != nil {
			return fmt.Errorf("rebuild prices for plan after quota add: %w", err)
		}

		created = quota
		return nil
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (s *AdminService) UpdateQuotaRule(ctx context.Context, id uuid.UUID, limitAmount int64, windowDurationSeconds int) (*domain.QuotaRule, error) {
	if limitAmount <= 0 || windowDurationSeconds <= 0 {
		return nil, domain.ErrInvalidQuotaValues
	}

	var updated *domain.QuotaRule
	err := s.repo.WithTx(ctx, func(tx pgx.Tx) error {
		quota, err := s.repo.GetQuotaRuleByID(ctx, tx, id)
		if err != nil {
			return err
		}

		plan, err := s.repo.GetPlanByID(ctx, tx, quota.PlanID)
		if err != nil {
			return err
		}

		// Plan entitlements and quotas can only be edited in draft status.
		// Once activated or inactivated, entitlements and quotas are strictly immutable.
		if plan.Status != domain.PlanStatusDraft {
			return domain.ErrPlanImmutable
		}

		quota.LimitAmount = limitAmount
		quota.WindowDurationSeconds = windowDurationSeconds
		quota.UpdatedAt = time.Now().UTC()

		if err := s.repo.UpdateQuotaRule(ctx, tx, quota); err != nil {
			return err
		}

		if err := s.repo.RebuildPricesForPlan(ctx, tx, plan.ID); err != nil {
			return fmt.Errorf("rebuild prices for plan after quota update: %w", err)
		}

		updated = quota
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *AdminService) DeleteQuotaRule(ctx context.Context, id uuid.UUID) error {
	return s.repo.WithTx(ctx, func(tx pgx.Tx) error {
		quota, err := s.repo.GetQuotaRuleByID(ctx, tx, id)
		if err != nil {
			return err
		}

		plan, err := s.repo.GetPlanByID(ctx, tx, quota.PlanID)
		if err != nil {
			return err
		}

		// Plan entitlements and quotas can only be edited in draft status.
		// Once activated or inactivated, entitlements and quotas are strictly immutable.
		if plan.Status != domain.PlanStatusDraft {
			return domain.ErrPlanImmutable
		}

		if err := s.repo.DeleteQuotaRule(ctx, tx, id); err != nil {
			return err
		}

		if err := s.repo.RebuildPricesForPlan(ctx, tx, plan.ID); err != nil {
			return fmt.Errorf("rebuild prices for plan after quota delete: %w", err)
		}

		return nil
	})
}

func (s *AdminService) GetQuotaRule(ctx context.Context, id uuid.UUID) (*domain.QuotaRule, error) {
	return s.repo.GetQuotaRuleByID(ctx, nil, id)
}

func (s *AdminService) ListQuotaRulesByPlanID(ctx context.Context, planID uuid.UUID) ([]domain.QuotaRule, error) {
	return s.repo.ListQuotaRulesByPlanID(ctx, nil, planID)
}
