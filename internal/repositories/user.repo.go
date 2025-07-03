package repositories

import (
	"context"
	"database/sql"
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
			total_earnings, total_withdraw, total_expenses, total_withdraw, total_expenses, address, subscription_id, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		) RETURNING id
	`
	err := r.db.QueryRow(ctx, query,
		user.Username, user.Password, user.Name, user.AvatarURL, user.Status, user.Role,
		user.Email, user.Mobile, user.TotalEarnings, user.TotalWithdraw, user.TotalExpenses, user.Address, user.SubscriptionID,
		time.Now(), time.Now(),
	).Scan(&user.ID)
	return err
}

func (r *UserRepo) GetByID(ctx context.Context, id int) (*models.User, error) {
	query := `
		SELECT 
			u.id,
			u.username,
			u.password,
			u.name,
			u.email,
			u.mobile,
			u.avatar_url,
			u.status,
			u.role,
			u.total_earnings,
			u.total_withdraw,
			u.address,
			u.subscription_id,
			u.created_at,
			u.updated_at,

			sp.id,
			sp.title,
			sp.terms,
			sp.download_limit,
			sp.time_limit,
			sp.status,
			sp.created_at,
			sp.updated_at
		FROM users u
		LEFT JOIN subscription_plans sp ON u.subscription_id = sp.id
		WHERE u.id = $1;
	`

	user := &models.User{}
	sub := &models.SubscriptionPlan{}

	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.Name,
		&user.Email,
		&user.Mobile,
		&user.AvatarURL,
		&user.Status,
		&user.Role,
		&user.TotalEarnings,
		&user.TotalWithdraw,
		&user.Address,
		&user.SubscriptionID,
		&user.CreatedAt,
		&user.UpdatedAt,

		&sub.ID,
		&sub.Title,
		&sub.Terms,
		&sub.DownloadLimit,
		&sub.TimeLimit,
		&sub.Status,
		&sub.CreatedAt,
		&sub.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// If no subscription (sub.ID is 0), set to nil
	if user.SubscriptionID == nil {
		user.SubscriptionPlan = nil
	} else {
		user.SubscriptionPlan = sub
	}

	return user, nil
}

func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `
		SELECT 
			u.id,
			u.username,
			u.password,
			u.name,
			u.email,
			u.mobile,
			u.avatar_url,
			u.status,
			u.role,
			u.total_earnings,
			u.total_withdraw,
			u.address,
			u.subscription_id,
			u.created_at,
			u.updated_at,

			sp.id,
			sp.title,
			sp.terms,
			sp.download_limit,
			sp.time_limit,
			sp.status,
			sp.created_at,
			sp.updated_at
		FROM users u
		LEFT JOIN subscription_plans sp ON u.subscription_id = sp.id
		WHERE u.usename = $1;
	`

	user := &models.User{}
	sub := &models.SubscriptionPlan{}

	err := r.db.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.Name,
		&user.Email,
		&user.Mobile,
		&user.AvatarURL,
		&user.Status,
		&user.Role,
		&user.TotalEarnings,
		&user.TotalWithdraw,
		&user.Address,
		&user.SubscriptionID,
		&user.CreatedAt,
		&user.UpdatedAt,

		&sub.ID,
		&sub.Title,
		&sub.Terms,
		&sub.DownloadLimit,
		&sub.TimeLimit,
		&sub.Status,
		&sub.CreatedAt,
		&sub.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// If no subscription (sub.ID is 0), set to nil
	if user.SubscriptionID == nil {
		user.SubscriptionPlan = nil
	} else {
		user.SubscriptionPlan = sub
	}

	return user, nil
}
func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT 
			u.id,
			u.username,
			u.password,
			u.name,
			u.email,
			u.mobile,
			u.avatar_url,
			u.status,
			u.role,
			u.total_earnings,
			u.total_withdraw,
			u.address,
			u.subscription_id,
			u.created_at,
			u.updated_at,

			sp.id,
			sp.title,
			sp.terms,
			sp.download_limit,
			sp.time_limit,
			sp.status,
			sp.created_at,
			sp.updated_at
		FROM users u
		LEFT JOIN subscription_plans sp ON u.subscription_id = sp.id
		WHERE u.email = $1;
	`

	// Nullable fields
	var (
		subID        sql.NullInt64
		subTitle     sql.NullString
		subTerms     sql.NullString
		subDL        sql.NullInt64
		subTimeLimit sql.NullString
		subStatus    sql.NullBool
		subCreatedAt sql.NullTime
		subUpdatedAt sql.NullTime
	)

	user := &models.User{}

	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.Name,
		&user.Email,
		&user.Mobile,
		&user.AvatarURL,
		&user.Status,
		&user.Role,
		&user.TotalEarnings,
		&user.TotalWithdraw,
		&user.Address,
		&user.SubscriptionID,
		&user.CreatedAt,
		&user.UpdatedAt,

		&subID,
		&subTitle,
		&subTerms,
		&subDL,
		&subTimeLimit,
		&subStatus,
		&subCreatedAt,
		&subUpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if subID.Valid {
		user.SubscriptionPlan = &models.SubscriptionPlan{
			ID:            int(subID.Int64),
			Title:         subTitle.String,
			Terms:         subTerms.String,
			DownloadLimit: int(subDL.Int64),
			TimeLimit:     subTimeLimit.String,
			Status:        subStatus.Bool,
			CreatedAt:     subCreatedAt.Time,
			UpdatedAt:     subUpdatedAt.Time,
		}
	} else {
		user.SubscriptionPlan = nil
	}

	return user, nil
}

func (r *UserRepo) Update(ctx context.Context, user *models.User) error {
	query := `
	UPDATE users
	SET 
		username = $2, password = $3, name = $4, avatar_url = $5, status = $6, 
		role = $7, email = $8, mobile = $9, total_earnings = $10, total_withdraw = $11, total_expenses = $12, address = $13, 
		subscription_id = $14, updated_at = $15
	WHERE id = $1`
	_, err := r.db.Exec(ctx, query,
		user.ID, user.Username, user.Password, user.Name, user.AvatarURL, user.Status,
		user.Role, user.Email, user.Mobile, user.TotalEarnings, user.TotalWithdraw, user.TotalExpenses, user.Address,
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
		total_earnings, total_withdraw, total_expenses, address, subscription_id, created_at, updated_at
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
			&user.Role, &user.Email, &user.Mobile, &user.TotalEarnings, &user.TotalWithdraw, &user.TotalExpenses, &user.Address,
			&user.SubscriptionID, &user.CreatedAt, &user.UpdatedAt,
		); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}
	return users, nil
}
