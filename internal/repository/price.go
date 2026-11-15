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

func (r *Repository) UpsertPrice(ctx context.Context, tx pgx.Tx, p *domain.Price) error {
	q := `
		INSERT INTO prices (id, offer_id, product_id, plan_id, currency, amount, status, catalog_visible, catalog_data, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (offer_id, currency) DO UPDATE
		SET amount = EXCLUDED.amount,
		    status = EXCLUDED.status,
		    updated_at = EXCLUDED.updated_at
		RETURNING id;
	`
	exec := r.getExecutor(tx)
	err := exec.QueryRow(ctx, q,
		p.ID, p.OfferID, p.ProductID, p.PlanID, p.Currency, p.Amount, p.Status,
		p.CatalogVisible, p.CatalogData, p.CreatedAt, p.UpdatedAt,
	).Scan(&p.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf("%w: price for currency %q already exists", domain.ErrConflict, p.Currency)
		}
		return fmt.Errorf("upsert price: %w", err)
	}
	return nil
}

func (r *Repository) UpdatePrice(ctx context.Context, tx pgx.Tx, p *domain.Price) error {
	q := `
		UPDATE prices
		SET amount = $1, status = $2, updated_at = $3
		WHERE id = $4;
	`
	exec := r.getExecutor(tx)
	ct, err := exec.Exec(ctx, q, p.Amount, p.Status, p.UpdatedAt, p.ID)
	if err != nil {
		return fmt.Errorf("update price: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrPriceNotFound
	}
	return nil
}

func (r *Repository) UpdatePriceStatus(ctx context.Context, tx pgx.Tx, id uuid.UUID, status domain.PriceStatus) error {
	q := `
		UPDATE prices
		SET status = $1, updated_at = $2
		WHERE id = $3;
	`
	exec := r.getExecutor(tx)
	ct, err := exec.Exec(ctx, q, status, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("update price status: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrPriceNotFound
	}
	return nil
}

func (r *Repository) UpdatePriceReadModel(ctx context.Context, tx pgx.Tx, priceID uuid.UUID, catalogVisible bool, catalogData []byte) error {
	q := `
		UPDATE prices
		SET catalog_visible = $1, catalog_data = $2, updated_at = now()
		WHERE id = $3;
	`
	exec := r.getExecutor(tx)
	ct, err := exec.Exec(ctx, q, catalogVisible, catalogData, priceID)
	if err != nil {
		return fmt.Errorf("update price read model: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrPriceNotFound
	}
	return nil
}

func (r *Repository) GetPriceByID(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*domain.Price, error) {
	q := `
		SELECT id, offer_id, product_id, plan_id, currency, amount, status, catalog_visible, catalog_data, created_at, updated_at
		FROM prices
		WHERE id = $1;
	`
	exec := r.getExecutor(tx)
	var p domain.Price
	err := exec.QueryRow(ctx, q, id).Scan(
		&p.ID, &p.OfferID, &p.ProductID, &p.PlanID, &p.Currency, &p.Amount,
		&p.Status, &p.CatalogVisible, &p.CatalogData, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPriceNotFound
		}
		return nil, fmt.Errorf("get price by id: %w", err)
	}
	return &p, nil
}

func (r *Repository) GetPriceByOfferAndCurrency(ctx context.Context, tx pgx.Tx, offerID uuid.UUID, currency string) (*domain.Price, error) {
	q := `
		SELECT id, offer_id, product_id, plan_id, currency, amount, status, catalog_visible, catalog_data, created_at, updated_at
		FROM prices
		WHERE offer_id = $1 AND currency = $2;
	`
	exec := r.getExecutor(tx)
	var p domain.Price
	err := exec.QueryRow(ctx, q, offerID, currency).Scan(
		&p.ID, &p.OfferID, &p.ProductID, &p.PlanID, &p.Currency, &p.Amount,
		&p.Status, &p.CatalogVisible, &p.CatalogData, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPriceNotFound
		}
		return nil, fmt.Errorf("get price by offer and currency: %w", err)
	}
	return &p, nil
}

func (r *Repository) ListPricesByOfferID(ctx context.Context, tx pgx.Tx, offerID uuid.UUID) ([]domain.Price, error) {
	q := `
		SELECT id, offer_id, product_id, plan_id, currency, amount, status, catalog_visible, catalog_data, created_at, updated_at
		FROM prices
		WHERE offer_id = $1
		ORDER BY currency ASC;
	`
	exec := r.getExecutor(tx)
	rows, err := exec.Query(ctx, q, offerID)
	if err != nil {
		return nil, fmt.Errorf("list prices by offer: %w", err)
	}
	defer rows.Close()

	var prices []domain.Price
	for rows.Next() {
		var p domain.Price
		if err := rows.Scan(
			&p.ID, &p.OfferID, &p.ProductID, &p.PlanID, &p.Currency, &p.Amount,
			&p.Status, &p.CatalogVisible, &p.CatalogData, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan price: %w", err)
		}
		prices = append(prices, p)
	}
	return prices, rows.Err()
}

func (r *Repository) ListPricesByPlanID(ctx context.Context, tx pgx.Tx, planID uuid.UUID) ([]domain.Price, error) {
	q := `
		SELECT id, offer_id, product_id, plan_id, currency, amount, status, catalog_visible, catalog_data, created_at, updated_at
		FROM prices
		WHERE plan_id = $1
		ORDER BY currency ASC;
	`
	exec := r.getExecutor(tx)
	rows, err := exec.Query(ctx, q, planID)
	if err != nil {
		return nil, fmt.Errorf("list prices by plan: %w", err)
	}
	defer rows.Close()

	var prices []domain.Price
	for rows.Next() {
		var p domain.Price
		if err := rows.Scan(
			&p.ID, &p.OfferID, &p.ProductID, &p.PlanID, &p.Currency, &p.Amount,
			&p.Status, &p.CatalogVisible, &p.CatalogData, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan price: %w", err)
		}
		prices = append(prices, p)
	}
	return prices, rows.Err()
}

func (r *Repository) ListPricesByProductID(ctx context.Context, tx pgx.Tx, productID uuid.UUID) ([]domain.Price, error) {
	q := `
		SELECT id, offer_id, product_id, plan_id, currency, amount, status, catalog_visible, catalog_data, created_at, updated_at
		FROM prices
		WHERE product_id = $1
		ORDER BY currency ASC;
	`
	exec := r.getExecutor(tx)
	rows, err := exec.Query(ctx, q, productID)
	if err != nil {
		return nil, fmt.Errorf("list prices by product: %w", err)
	}
	defer rows.Close()

	var prices []domain.Price
	for rows.Next() {
		var p domain.Price
		if err := rows.Scan(
			&p.ID, &p.OfferID, &p.ProductID, &p.PlanID, &p.Currency, &p.Amount,
			&p.Status, &p.CatalogVisible, &p.CatalogData, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan price: %w", err)
		}
		prices = append(prices, p)
	}
	return prices, rows.Err()
}

// ListCatalogPrices fulfills the single-query catalog read specification:
// SELECT id AS price_id, offer_id, plan_id, currency, amount, catalog_data
// FROM prices
// WHERE product_id = $1 AND currency = $2 AND status = 'active' AND catalog_visible = true;
func (r *Repository) ListCatalogPrices(ctx context.Context, tx pgx.Tx, productID uuid.UUID, currency string) ([]domain.CatalogItem, error) {
	q := `
		SELECT id AS price_id, offer_id, plan_id, currency, amount, catalog_data
		FROM prices
		WHERE product_id = $1
		  AND currency = $2
		  AND status = 'active'
		  AND catalog_visible = true
		ORDER BY (catalog_data->'offer'->>'sort_order')::int ASC, created_at ASC;
	`
	exec := r.getExecutor(tx)
	rows, err := exec.Query(ctx, q, productID, currency)
	if err != nil {
		return nil, fmt.Errorf("list catalog prices: %w", err)
	}
	defer rows.Close()

	var items []domain.CatalogItem
	for rows.Next() {
		var item domain.CatalogItem
		if err := rows.Scan(&item.PriceID, &item.OfferID, &item.PlanID, &item.Currency, &item.Amount, &item.CatalogData); err != nil {
			return nil, fmt.Errorf("scan catalog item: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
