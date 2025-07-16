package repositories

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samiulice/photostock/internal/models"
)

type SubscriptionRepo struct {
	db *pgxpool.Pool
}

func NewSubscriptionRepo(db *pgxpool.Pool) *SubscriptionRepo {
	return &SubscriptionRepo{db: db}
}

// Create a new subscription
func (r *SubscriptionRepo) Create(ctx context.Context, sub *models.Subscription) error {
	query := `
		INSERT INTO subscriptions (
			user_id, subscription_plans_id, payment_amount, 
			payment_time, total_downloads, status
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`
	return r.db.QueryRow(ctx, query,
		sub.UserID, sub.SubscriptionPlanID,
		sub.PaymentAmount, sub.PaymentTime, sub.TotalDownloads, sub.Status,
	).Scan(&sub.ID)
}

// Get a subscription by ID (with optional join on plan)
func (r *SubscriptionRepo) GetByID(ctx context.Context, id int) (*models.Subscription, error) {
	query := `
		SELECT s.id, s.user_id, s.subscription_plans_id,
		        s.payment_amount, s.payment_time,
		       s.total_downloads, s.status, s.created_at, s.updated_at,
		       p.id, p.title, p.terms, p.status, p.download_limit, p.time_limit::text, p.created_at, p.updated_at
		FROM subscriptions s
		LEFT JOIN subscription_plans p ON s.subscription_plans_id = p.id
		WHERE s.id = $1`

	sub := &models.Subscription{PlanDetails: &models.SubscriptionPlan{}}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&sub.ID, &sub.UserID, &sub.SubscriptionPlanID, &sub.PaymentAmount, &sub.PaymentTime,
		&sub.TotalDownloads, &sub.Status, &sub.CreatedAt, &sub.UpdatedAt,
		&sub.PlanDetails.ID, &sub.PlanDetails.Title, &sub.PlanDetails.Terms,
		&sub.PlanDetails.Status, &sub.PlanDetails.DownloadLimit,
		&sub.PlanDetails.ExpiresAt, &sub.PlanDetails.CreatedAt, &sub.PlanDetails.UpdatedAt,
	)
	return sub, err
}

// Update a subscription
func (r *SubscriptionRepo) Update(ctx context.Context, sub *models.Subscription) error {
	query := `
		UPDATE subscriptions
		SET user_id = $2,
		    subscription_plans_id = $3,
		    payment_amount = $4,
		    payment_time = $5,
		    total_downloads = $6,
		    status = $7,
		    updated_at = $8
		WHERE id = $1`
	_, err := r.db.Exec(ctx, query,
		sub.ID, sub.UserID, sub.SubscriptionPlanID,
		sub.PaymentAmount, sub.PaymentTime, sub.TotalDownloads, sub.Status, time.Now(),
	)
	return err
}

// Delete a subscription
func (r *SubscriptionRepo) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM subscriptions WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// Get all subscriptions
func (r *SubscriptionRepo) GetAll(ctx context.Context) ([]*models.Subscription, error) {
	query := `
		SELECT s.id, s.user_id, s.subscription_plans_id,
		        s.payment_amount, s.payment_time,
		       s.total_downloads, s.status, s.created_at, s.updated_at,
		       p.id, p.title, p.terms, p.status, p.download_limit, p.time_limit::text, p.created_at, p.updated_at
		FROM subscriptions s
		LEFT JOIN subscription_plans p ON s.subscription_plans_id = p.id
		ORDER BY s.created_at DESC`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []*models.Subscription
	for rows.Next() {
		var sub models.Subscription
		sub.PlanDetails = &models.SubscriptionPlan{}
		err := rows.Scan(
			&sub.ID, &sub.UserID, &sub.SubscriptionPlanID, &sub.PaymentAmount, &sub.PaymentTime,
			&sub.TotalDownloads, &sub.Status, &sub.CreatedAt, &sub.UpdatedAt,
			&sub.PlanDetails.ID, &sub.PlanDetails.Title, &sub.PlanDetails.Terms,
			&sub.PlanDetails.Status, &sub.PlanDetails.DownloadLimit,
			&sub.PlanDetails.ExpiresAt, &sub.PlanDetails.CreatedAt, &sub.PlanDetails.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		subs = append(subs, &sub)
	}
	return subs, nil
}

func (r *SubscriptionRepo) GetByUserID(ctx context.Context, userID int) ([]*models.Subscription, error) {
	query := `
		SELECT s.id, s.user_id, s.subscription_plans_id,
		        s.payment_amount, s.payment_time,
		       s.total_downloads, s.status, s.created_at, s.updated_at,
		       p.id, p.title, p.terms, p.status, p.download_limit, p.time_limit::text, p.created_at, p.updated_at
		FROM subscriptions s
		LEFT JOIN subscription_plans p ON s.subscription_plans_id = p.id
		WHERE s.user_id = $1
		ORDER BY s.created_at DESC`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []*models.Subscription
	for rows.Next() {
		var sub models.Subscription
		sub.PlanDetails = &models.SubscriptionPlan{}
		err := rows.Scan(
			&sub.ID, &sub.UserID, &sub.SubscriptionPlanID, &sub.PaymentAmount, &sub.PaymentTime,
			&sub.TotalDownloads, &sub.Status, &sub.CreatedAt, &sub.UpdatedAt,
			&sub.PlanDetails.ID, &sub.PlanDetails.Title, &sub.PlanDetails.Terms,
			&sub.PlanDetails.Status, &sub.PlanDetails.DownloadLimit,
			&sub.PlanDetails.ExpiresAt, &sub.PlanDetails.CreatedAt, &sub.PlanDetails.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		subs = append(subs, &sub)
	}
	return subs, nil
}
