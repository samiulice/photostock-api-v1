package repositories

import (
	"context"
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

func (r *SubscriptionTypeRepo) Create(ctx context.Context, st *models.SubscriptionPlan) error {
	query := `
		INSERT INTO subscription_plans (title, terms, status, download_limit, time_limit, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5::interval, $6, $7)
		RETURNING id`
	return r.db.QueryRow(ctx, query,
		st.Title, st.Terms, st.Status, st.DownloadLimit, st.ExpiresAt, time.Now(), time.Now(),
	).Scan(&st.ID)
}

func (r *SubscriptionTypeRepo) GetByID(ctx context.Context, id int) (*models.SubscriptionPlan, error) {
	query := `
		SELECT id, title, terms, status, download_limit, time_limit::text, created_at, updated_at
		FROM subscription_plans
		WHERE id = $1`
	st := &models.SubscriptionPlan{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&st.ID, &st.Title, &st.Terms, &st.Status, &st.DownloadLimit, &st.ExpiresAt, &st.CreatedAt, &st.UpdatedAt,
	)
	return st, err
}

func (r *SubscriptionTypeRepo) Update(ctx context.Context, st *models.SubscriptionPlan) error {
	query := `
		UPDATE subscription_plans
		SET title = $2, terms = $3, status = $4, download_limit = $5, time_limit = $6::interval, updated_at = $7
		WHERE id = $1`
	_, err := r.db.Exec(ctx, query,
		st.ID, st.Title, st.Terms, st.Status, st.DownloadLimit, st.ExpiresAt, time.Now(),
	)
	return err
}

func (r *SubscriptionTypeRepo) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM subscription_plans WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *SubscriptionTypeRepo) GetAll(ctx context.Context) ([]models.SubscriptionPlan, error) {
	query := `
		SELECT id, title, terms, status, download_limit, time_limit::text, created_at, updated_at
		FROM subscription_plans`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var types []models.SubscriptionPlan
	for rows.Next() {
		var st models.SubscriptionPlan
		if err := rows.Scan(
			&st.ID, &st.Title, &st.Terms, &st.Status, &st.DownloadLimit, &st.ExpiresAt, &st.CreatedAt, &st.UpdatedAt,
		); err != nil {
			return nil, err
		}
		types = append(types, st)
	}
	return types, nil
}
