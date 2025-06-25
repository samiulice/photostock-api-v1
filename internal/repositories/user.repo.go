package repositories

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samiulice/photostock/internal/models"
)

// ============================== User Repository ==============================
type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (
			username, password, name, avatar_url, status, role, email, mobile, 
			total_earnings, address, subscription_id, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		) RETURNING id
	`
	err := r.db.QueryRow(ctx, query,
		user.Username, user.Password, user.Name, user.AvatarURL, user.Status, user.Role,
		user.Email, user.Mobile, user.TotalEarnings, user.Address, user.SubscriptionID,
		time.Now(), time.Now(),
	).Scan(&user.ID)
	return err
}

func (r *UserRepo) GetByID(ctx context.Context, id int) (*models.User, error) {
	query := `
	SELECT id, username, password, name, avatar_url, status, role, email, mobile,
		total_earnings, address, subscription_id, created_at, updated_at
	FROM users
	WHERE id = $1`
	user := &models.User{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.Password, &user.Name, &user.AvatarURL, &user.Status, &user.Role,
		&user.Email, &user.Mobile, &user.TotalEarnings, &user.Address, &user.SubscriptionID,
		&user.CreatedAt, &user.UpdatedAt,
	)
	return user, err
}

func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `
	SELECT id, username, password, name, avatar_url, status, role, email, mobile,
		total_earnings, address, subscription_id, created_at, updated_at
	FROM users
	WHERE username = $1`
	user := &models.User{}
	err := r.db.QueryRow(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.Password, &user.Name, &user.AvatarURL, &user.Status, &user.Role,
		&user.Email, &user.Mobile, &user.TotalEarnings, &user.Address, &user.SubscriptionID,
		&user.CreatedAt, &user.UpdatedAt,
	)
	return user, err
}

func (r *UserRepo) Update(ctx context.Context, user *models.User) error {
	query := `
	UPDATE users
	SET 
		username = $2, password = $3, name = $4, avatar_url = $5, status = $6, 
		role = $7, email = $8, mobile = $9, total_earnings = $10, address = $11, 
		subscription_id = $12, updated_at = $13
	WHERE id = $1`
	_, err := r.db.Exec(ctx, query,
		user.ID, user.Username, user.Password, user.Name, user.AvatarURL, user.Status,
		user.Role, user.Email, user.Mobile, user.TotalEarnings, user.Address,
		user.SubscriptionID, time.Now(),
	)
	return err
}

func (r *UserRepo) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *UserRepo) GetAll(ctx context.Context) ([]*models.User, error) {
	query := `
	SELECT id, username, password, name, avatar_url, status, role, email, mobile,
		total_earnings, address, subscription_id, created_at, updated_at
	FROM users`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(
			&user.ID, &user.Username, &user.Password, &user.Name, &user.AvatarURL, &user.Status,
			&user.Role, &user.Email, &user.Mobile, &user.TotalEarnings, &user.Address,
			&user.SubscriptionID, &user.CreatedAt, &user.UpdatedAt,
		); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}
	return users, nil
}
