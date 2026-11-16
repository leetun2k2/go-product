package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/leetun2k2/go-product/internal/domain"
)

func (r *Repository) CreateOffer(ctx context.Context, tx pgx.Tx, o *domain.Offer) error {
	q := `
		INSERT INTO product.offers (id, product_id, plan_id, code, offer_type, billing_cycle, status, sort_order, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);
	`
	exec := r.getExecutor(tx)
	_, err := exec.Exec(ctx, q, o.ID, o.ProductID, o.PlanID, o.Code, o.OfferType, o.BillingCycle, o.Status, o.SortOrder, o.CreatedAt, o.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf("%w: offer billing cycle %q already exists for plan", domain.ErrConflict, o.BillingCycle)
		}
		return fmt.Errorf("insert offer: %w", err)
	}
	return nil
}

func (r *Repository) UpdateOffer(ctx context.Context, tx pgx.Tx, o *domain.Offer) error {
	q := `
		UPDATE product.offers
		SET code = $1, sort_order = $2, updated_at = $3
		WHERE id = $4;
	`
	exec := r.getExecutor(tx)
	ct, err := exec.Exec(ctx, q, o.Code, o.SortOrder, o.UpdatedAt, o.ID)
	if err != nil {
		return fmt.Errorf("update offer: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrOfferNotFound
	}
	return nil
}

func (r *Repository) UpdateOfferStatus(ctx context.Context, tx pgx.Tx, id uuid.UUID, status domain.OfferStatus) error {
	q := `
		UPDATE product.offers
		SET status = $1, updated_at = $2
		WHERE id = $3;
	`
	exec := r.getExecutor(tx)
	ct, err := exec.Exec(ctx, q, status, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("update offer status: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrOfferNotFound
	}
	return nil
}

func (r *Repository) GetOfferByID(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*domain.Offer, error) {
	q := `
		SELECT id, product_id, plan_id, code, offer_type, billing_cycle, status, sort_order, created_at, updated_at
		FROM product.offers
		WHERE id = $1;
	`
	exec := r.getExecutor(tx)
	var o domain.Offer
	err := exec.QueryRow(ctx, q, id).Scan(&o.ID, &o.ProductID, &o.PlanID, &o.Code, &o.OfferType, &o.BillingCycle, &o.Status, &o.SortOrder, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrOfferNotFound
		}
		return nil, fmt.Errorf("get offer by id: %w", err)
	}
	return &o, nil
}

func (r *Repository) ListOffersByPlanID(ctx context.Context, tx pgx.Tx, planID uuid.UUID) ([]domain.Offer, error) {
	q := `
		SELECT id, product_id, plan_id, code, offer_type, billing_cycle, status, sort_order, created_at, updated_at
		FROM product.offers
		WHERE plan_id = $1
		ORDER BY sort_order ASC, created_at ASC;
	`
	exec := r.getExecutor(tx)
	rows, err := exec.Query(ctx, q, planID)
	if err != nil {
		return nil, fmt.Errorf("list offers: %w", err)
	}
	defer rows.Close()

	var offers []domain.Offer
	for rows.Next() {
		var o domain.Offer
		if err := rows.Scan(&o.ID, &o.ProductID, &o.PlanID, &o.Code, &o.OfferType, &o.BillingCycle, &o.Status, &o.SortOrder, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan offer: %w", err)
		}
		offers = append(offers, o)
	}
	return offers, rows.Err()
}

func (r *Repository) ListActiveOffersByPlanID(ctx context.Context, tx pgx.Tx, planID uuid.UUID) ([]domain.Offer, error) {
	q := `
		SELECT id, product_id, plan_id, code, offer_type, billing_cycle, status, sort_order, created_at, updated_at
		FROM product.offers
		WHERE plan_id = $1 AND status = 'active'
		ORDER BY sort_order ASC, created_at ASC;
	`
	exec := r.getExecutor(tx)
	rows, err := exec.Query(ctx, q, planID)
	if err != nil {
		return nil, fmt.Errorf("list active offers: %w", err)
	}
	defer rows.Close()

	var offers []domain.Offer
	for rows.Next() {
		var o domain.Offer
		if err := rows.Scan(&o.ID, &o.ProductID, &o.PlanID, &o.Code, &o.OfferType, &o.BillingCycle, &o.Status, &o.SortOrder, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan active offer: %w", err)
		}
		offers = append(offers, o)
	}
	return offers, rows.Err()
}
