package repositories

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samiulice/photostock/internal/models"
)

// ============================== Subscriptions Repository ==============================
type SubscriptionRepo struct {
	db *pgxpool.Pool
}

func NewSubscriptionRepo(db *pgxpool.Pool) *SubscriptionRepo {
	return &SubscriptionRepo{db: db}
}

func (r *SubscriptionRepo) Create(ctx context.Context, sub *models.Subscription) error {
	query := `
	INSERT INTO subscriptions (
		user_id, subscription_type_id, payment_status, payment_amount, 
		payment_time, status, created_at, updated_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	RETURNING id`
	return r.db.QueryRow(ctx, query,
		sub.UserID, sub.SubscriptionTypeID, sub.PaymentStatus, sub.PaymentAmount,
		sub.PaymentTime, sub.Status, time.Now(), time.Now(),
	).Scan(&sub.ID)
}

func (r *SubscriptionRepo) GetByID(ctx context.Context, id int) (*models.Subscription, error) {
	query := `
	SELECT id, user_id, subscription_type_id, payment_status, payment_amount, 
		payment_time, status, created_at, updated_at
	FROM subscriptions
	WHERE id = $1`
	sub := &models.Subscription{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&sub.ID, &sub.UserID, &sub.SubscriptionTypeID, &sub.PaymentStatus,
		&sub.PaymentAmount, &sub.PaymentTime, &sub.Status, &sub.CreatedAt, &sub.UpdatedAt,
	)
	return sub, err
}

func (r *SubscriptionRepo) Update(ctx context.Context, sub *models.Subscription) error {
	query := `
	UPDATE subscriptions
	SET 
		user_id = $2, subscription_type_id = $3, payment_status = $4, 
		payment_amount = $5, payment_time = $6, status = $7, updated_at = $8
	WHERE id = $1`
	_, err := r.db.Exec(ctx, query,
		sub.ID, sub.UserID, sub.SubscriptionTypeID, sub.PaymentStatus,
		sub.PaymentAmount, sub.PaymentTime, sub.Status, time.Now(),
	)
	return err
}

func (r *SubscriptionRepo) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM subscriptions WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *SubscriptionRepo) GetAll(ctx context.Context) ([]models.Subscription, error) {
	query := `
	SELECT id, user_id, subscription_type_id, payment_status, payment_amount, 
		payment_time, status, created_at, updated_at
	FROM subscriptions`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []models.Subscription
	for rows.Next() {
		var sub models.Subscription
		if err := rows.Scan(
			&sub.ID, &sub.UserID, &sub.SubscriptionTypeID, &sub.PaymentStatus,
			&sub.PaymentAmount, &sub.PaymentTime, &sub.Status, &sub.CreatedAt, &sub.UpdatedAt,
		); err != nil {
			return nil, err
		}
		subs = append(subs, sub)
	}
	return subs, nil
}
