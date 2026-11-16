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

func (r *Repository) CreatePlan(ctx context.Context, tx pgx.Tx, p *domain.Plan) error {
	q := `
		INSERT INTO product.plans (id, product_id, code, name, description, status, version, activated_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);
	`
	exec := r.getExecutor(tx)
	_, err := exec.Exec(ctx, q, p.ID, p.ProductID, p.Code, p.Name, p.Description, p.Status, p.Version, p.ActivatedAt, p.CreatedAt, p.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf("%w: plan code %q version %d", domain.ErrConflict, p.Code, p.Version)
		}
		return fmt.Errorf("insert plan: %w", err)
	}
	return nil
}

func (r *Repository) UpdatePlan(ctx context.Context, tx pgx.Tx, p *domain.Plan) error {
	q := `
		UPDATE product.plans
		SET name = $1, description = $2, updated_at = $3
		WHERE id = $4;
	`
	exec := r.getExecutor(tx)
	ct, err := exec.Exec(ctx, q, p.Name, p.Description, p.UpdatedAt, p.ID)
	if err != nil {
		return fmt.Errorf("update plan: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrPlanNotFound
	}
	return nil
}

func (r *Repository) UpdatePlanStatus(ctx context.Context, tx pgx.Tx, id uuid.UUID, status domain.PlanStatus) error {
	q := `
		UPDATE product.plans
		SET status = $1, updated_at = $2
		WHERE id = $3;
	`
	exec := r.getExecutor(tx)
	ct, err := exec.Exec(ctx, q, status, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("update plan status: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrPlanNotFound
	}
	return nil
}

func (r *Repository) ActivatePlan(ctx context.Context, tx pgx.Tx, id uuid.UUID) error {
	now := time.Now().UTC()
	q := `
		UPDATE product.plans
		SET status = $1, activated_at = COALESCE(activated_at, $2), updated_at = $2
		WHERE id = $3;
	`
	exec := r.getExecutor(tx)
	ct, err := exec.Exec(ctx, q, domain.PlanStatusActive, now, id)
	if err != nil {
		return fmt.Errorf("activate plan: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrPlanNotFound
	}
	return nil
}

func (r *Repository) GetPlanByID(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*domain.Plan, error) {
	q := `
		SELECT id, product_id, code, name, description, status, version, activated_at, created_at, updated_at
		FROM product.plans
		WHERE id = $1;
	`
	exec := r.getExecutor(tx)
	var p domain.Plan
	err := exec.QueryRow(ctx, q, id).Scan(&p.ID, &p.ProductID, &p.Code, &p.Name, &p.Description, &p.Status, &p.Version, &p.ActivatedAt, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPlanNotFound
		}
		return nil, fmt.Errorf("get plan by id: %w", err)
	}
	return &p, nil
}

func (r *Repository) GetPlanByIDForUpdate(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*domain.Plan, error) {
	q := `
		SELECT id, product_id, code, name, description, status, version, activated_at, created_at, updated_at
		FROM product.plans
		WHERE id = $1
		FOR UPDATE;
	`
	exec := r.getExecutor(tx)
	var p domain.Plan
	err := exec.QueryRow(ctx, q, id).Scan(&p.ID, &p.ProductID, &p.Code, &p.Name, &p.Description, &p.Status, &p.Version, &p.ActivatedAt, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPlanNotFound
		}
		return nil, fmt.Errorf("get plan by id for update: %w", err)
	}
	return &p, nil
}

func (r *Repository) ListPlansByProductID(ctx context.Context, tx pgx.Tx, productID uuid.UUID) ([]domain.Plan, error) {
	q := `
		SELECT id, product_id, code, name, description, status, version, activated_at, created_at, updated_at
		FROM product.plans
		WHERE product_id = $1
		ORDER BY version DESC, created_at ASC;
	`
	exec := r.getExecutor(tx)
	rows, err := exec.Query(ctx, q, productID)
	if err != nil {
		return nil, fmt.Errorf("list plans: %w", err)
	}
	defer rows.Close()

	var plans []domain.Plan
	for rows.Next() {
		var p domain.Plan
		if err := rows.Scan(&p.ID, &p.ProductID, &p.Code, &p.Name, &p.Description, &p.Status, &p.Version, &p.ActivatedAt, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan plan: %w", err)
		}
		plans = append(plans, p)
	}
	return plans, rows.Err()
}
