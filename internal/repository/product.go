package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/leetun2k2/go-product/internal/domain"
)

func (r *Repository) CreateProduct(ctx context.Context, tx pgx.Tx, p *domain.Product) error {
	q := `
		INSERT INTO products (id, code, name, description, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7);
	`
	exec := r.getExecutor(tx)
	_, err := exec.Exec(ctx, q, p.ID, p.Code, p.Name, p.Description, p.Status, p.CreatedAt, p.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf("%w: product code %q", domain.ErrConflict, p.Code)
		}
		return fmt.Errorf("insert product: %w", err)
	}
	return nil
}

func (r *Repository) UpdateProduct(ctx context.Context, tx pgx.Tx, p *domain.Product) error {
	q := `
		UPDATE products
		SET name = $1, description = $2, status = $3, updated_at = $4
		WHERE id = $5;
	`
	exec := r.getExecutor(tx)
	ct, err := exec.Exec(ctx, q, p.Name, p.Description, p.Status, p.UpdatedAt, p.ID)
	if err != nil {
		return fmt.Errorf("update product: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrProductNotFound
	}
	return nil
}

func (r *Repository) GetProductByID(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*domain.Product, error) {
	q := `
		SELECT id, code, name, description, status, created_at, updated_at
		FROM products
		WHERE id = $1;
	`
	exec := r.getExecutor(tx)
	var p domain.Product
	err := exec.QueryRow(ctx, q, id).Scan(&p.ID, &p.Code, &p.Name, &p.Description, &p.Status, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrProductNotFound
		}
		return nil, fmt.Errorf("get product by id: %w", err)
	}
	return &p, nil
}

func (r *Repository) GetProductByCode(ctx context.Context, tx pgx.Tx, code string) (*domain.Product, error) {
	q := `
		SELECT id, code, name, description, status, created_at, updated_at
		FROM products
		WHERE code = $1;
	`
	exec := r.getExecutor(tx)
	var p domain.Product
	err := exec.QueryRow(ctx, q, code).Scan(&p.ID, &p.Code, &p.Name, &p.Description, &p.Status, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrProductNotFound
		}
		return nil, fmt.Errorf("get product by code: %w", err)
	}
	return &p, nil
}

func (r *Repository) ListProducts(ctx context.Context, tx pgx.Tx) ([]domain.Product, error) {
	q := `
		SELECT id, code, name, description, status, created_at, updated_at
		FROM products
		ORDER BY created_at ASC;
	`
	exec := r.getExecutor(tx)
	rows, err := exec.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}
	defer rows.Close()

	var products []domain.Product
	for rows.Next() {
		var p domain.Product
		if err := rows.Scan(&p.ID, &p.Code, &p.Name, &p.Description, &p.Status, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan product: %w", err)
		}
		products = append(products, p)
	}
	return products, rows.Err()
}
