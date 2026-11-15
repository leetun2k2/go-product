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

func (r *Repository) CreateQuotaRule(ctx context.Context, tx pgx.Tx, q *domain.QuotaRule) error {
	query := `
		INSERT INTO quota_rules (id, product_id, plan_id, rule_key, limit_amount, window_duration_seconds, anchor_type, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);
	`
	exec := r.getExecutor(tx)
	_, err := exec.Exec(ctx, query, q.ID, q.ProductID, q.PlanID, q.RuleKey, q.LimitAmount, q.WindowDurationSeconds, q.AnchorType, q.CreatedAt, q.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf("%w: quota rule %q with anchor %q already exists for plan", domain.ErrConflict, q.RuleKey, q.AnchorType)
		}
		return fmt.Errorf("insert quota rule: %w", err)
	}
	return nil
}

func (r *Repository) UpdateQuotaRule(ctx context.Context, tx pgx.Tx, q *domain.QuotaRule) error {
	query := `
		UPDATE quota_rules
		SET limit_amount = $1, window_duration_seconds = $2, updated_at = $3
		WHERE id = $4;
	`
	exec := r.getExecutor(tx)
	ct, err := exec.Exec(ctx, query, q.LimitAmount, q.WindowDurationSeconds, q.UpdatedAt, q.ID)
	if err != nil {
		return fmt.Errorf("update quota rule: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrQuotaNotFound
	}
	return nil
}

func (r *Repository) DeleteQuotaRule(ctx context.Context, tx pgx.Tx, id uuid.UUID) error {
	query := `
		DELETE FROM quota_rules
		WHERE id = $1;
	`
	exec := r.getExecutor(tx)
	ct, err := exec.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete quota rule: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrQuotaNotFound
	}
	return nil
}

func (r *Repository) GetQuotaRuleByID(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*domain.QuotaRule, error) {
	query := `
		SELECT id, product_id, plan_id, rule_key, limit_amount, window_duration_seconds, anchor_type, created_at, updated_at
		FROM quota_rules
		WHERE id = $1;
	`
	exec := r.getExecutor(tx)
	var q domain.QuotaRule
	err := exec.QueryRow(ctx, query, id).Scan(
		&q.ID, &q.ProductID, &q.PlanID, &q.RuleKey, &q.LimitAmount,
		&q.WindowDurationSeconds, &q.AnchorType, &q.CreatedAt, &q.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrQuotaNotFound
		}
		return nil, fmt.Errorf("get quota rule by id: %w", err)
	}
	return &q, nil
}

func (r *Repository) ListQuotaRulesByPlanID(ctx context.Context, tx pgx.Tx, planID uuid.UUID) ([]domain.QuotaRule, error) {
	query := `
		SELECT id, product_id, plan_id, rule_key, limit_amount, window_duration_seconds, anchor_type, created_at, updated_at
		FROM quota_rules
		WHERE plan_id = $1
		ORDER BY rule_key ASC, anchor_type ASC;
	`
	exec := r.getExecutor(tx)
	rows, err := exec.Query(ctx, query, planID)
	if err != nil {
		return nil, fmt.Errorf("list quota rules: %w", err)
	}
	defer rows.Close()

	var rules []domain.QuotaRule
	for rows.Next() {
		var q domain.QuotaRule
		if err := rows.Scan(
			&q.ID, &q.ProductID, &q.PlanID, &q.RuleKey, &q.LimitAmount,
			&q.WindowDurationSeconds, &q.AnchorType, &q.CreatedAt, &q.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan quota rule: %w", err)
		}
		rules = append(rules, q)
	}
	return rules, rows.Err()
}
