package repositories

import (
	"context"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samiulice/photostock/internal/models"
)

// ============================== SubscriptionPlan Repository ==============================
type SubscriptionTypeRepo struct {
	db *pgxpool.Pool
}

func NewSubscriptionPlanRepo(db *pgxpool.Pool) *SubscriptionTypeRepo {
	return &SubscriptionTypeRepo{db: db}
}

func (r *SubscriptionTypeRepo) Create(ctx context.Context, sp *models.SubscriptionPlan) error {
	sp.Terms = strings.Join(sp.TermsList, "[[]]") // Concatenate terms
	query := `
		INSERT INTO subscription_plans (title, terms, status, price, download_limit, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`
	return r.db.QueryRow(ctx, query,
		sp.Title, sp.Terms, sp.Status, sp.Price, sp.DownloadLimit, sp.ExpiresAt, time.Now(), time.Now(),
	).Scan(&sp.ID)
}

func (r *SubscriptionTypeRepo) Update(ctx context.Context, sp *models.SubscriptionPlan) error {
	sp.Terms = strings.Join(sp.TermsList, ",") // Concatenate terms
	query := `
		UPDATE subscription_plans
		SET title = $2, terms = $3, status = $4, price = $5, download_limit = $6, expires_at = $7, updated_at = $8
		WHERE id = $1`
	_, err := r.db.Exec(ctx, query,
		sp.ID, sp.Title, sp.Terms, sp.Status, sp.Price, sp.DownloadLimit, sp.ExpiresAt, time.Now(),
	)
	return err
}

func (r *SubscriptionTypeRepo) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM subscription_plans WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *SubscriptionTypeRepo) GetAll(ctx context.Context) ([]*models.SubscriptionPlan, error) {
	query := `
		SELECT id, title, terms, status, price, download_limit, expires_at, created_at, updated_at
		FROM subscription_plans`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []*models.SubscriptionPlan
	for rows.Next() {
		var sp models.SubscriptionPlan

		err := rows.Scan(
			&sp.ID, &sp.Title, &sp.Terms, &sp.Status, &sp.Price, &sp.DownloadLimit, &sp.ExpiresAt, &sp.CreatedAt, &sp.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Split terms for client use
		sp.TermsList = strings.Split(sp.Terms, "[[]]")

		plans = append(plans, &sp)
	}
	return plans, nil
}

func (r *SubscriptionTypeRepo) GetByID(ctx context.Context, id int) (*models.SubscriptionPlan, error) {
	query := `
		SELECT id, title, terms, status, price, download_limit, expires_at, created_at, updated_at
		FROM subscription_plans
		WHERE id = $1`
	var sp models.SubscriptionPlan
	err := r.db.QueryRow(ctx, query, id).Scan(
		&sp.ID, &sp.Title, &sp.Terms, &sp.Status, &sp.Price, &sp.DownloadLimit, &sp.ExpiresAt, &sp.CreatedAt, &sp.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	sp.TermsList = strings.Split(sp.Terms, "[[]]")
	return &sp, nil
}
