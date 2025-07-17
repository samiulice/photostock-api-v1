package repositories

import (
	"context"
	"database/sql"
	"net/url"
	"path"
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
			total_earnings, total_withdraw, total_expenses, address, subscription_id, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		) RETURNING id
	`
	err := r.db.QueryRow(ctx, query,
		user.Username, user.Password, user.Name, user.AvatarID, user.Status, user.Role,
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
			u.total_expenses,
			u.address,
			u.subscription_id,
			u.created_at,
			u.updated_at,
			
			s.id,
			s.user_id,
			s.subscription_plans_id,
			s.payment_amount,
			s.payment_time,
			s.total_downloads,
			s.status,
			s.created_at,
			s.updated_at,
			
			sp.id,
			sp.title,
			sp.terms,
			sp.status,
			sp.download_limit,
			sp.expires_at,
			sp.created_at,
			sp.updated_at
		FROM users u
		LEFT JOIN subscriptions s ON u.subscription_id = s.id
		LEFT JOIN subscription_plans sp ON s.subscription_plans_id = sp.id
		WHERE u.id = $1;
	`

	var (
		user models.User
		sub  models.Subscription
		plan models.SubscriptionPlan

		// Nullable fields
		subID     sql.NullInt64
		subUserID sql.NullInt64
		subPlanID sql.NullInt64

		subPaymentAmount  sql.NullFloat64
		subPaymentTime    sql.NullTime
		subTotalDownloads sql.NullInt64
		subStatus         sql.NullBool
		subCreatedAt      sql.NullTime
		subUpdatedAt      sql.NullTime

		planID        sql.NullInt64
		planTitle     sql.NullString
		planTerms     sql.NullString
		planStatus    sql.NullBool
		planDL        sql.NullInt64
		planExpiresAt sql.NullInt64
		planCreatedAt sql.NullTime
		planUpdatedAt sql.NullTime
	)

	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.Name,
		&user.Email,
		&user.Mobile,
		&user.AvatarID,
		&user.Status,
		&user.Role,
		&user.TotalEarnings,
		&user.TotalWithdraw,
		&user.TotalExpenses,
		&user.Address,
		&user.SubscriptionID,
		&user.CreatedAt,
		&user.UpdatedAt,

		&subID,
		&subUserID,
		&subPlanID,
		&subPaymentAmount,
		&subPaymentTime,
		&subTotalDownloads,
		&subStatus,
		&subCreatedAt,
		&subUpdatedAt,

		&planID,
		&planTitle,
		&planTerms,
		&planStatus,
		&planDL,
		&planExpiresAt,
		&planCreatedAt,
		&planUpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Build subscription if exists
	if subID.Valid {
		sub = models.Subscription{
			ID:                 int(subID.Int64),
			UserID:             int(subUserID.Int64),
			SubscriptionPlanID: int(subPlanID.Int64),
			PaymentAmount:      subPaymentAmount.Float64,
			PaymentTime:        subPaymentTime.Time,
			TotalDownloads:     int(subTotalDownloads.Int64),
			Status:             subStatus.Bool,
			CreatedAt:          subCreatedAt.Time,
			UpdatedAt:          subUpdatedAt.Time,
		}

		// Build subscription plan if exists
		if planID.Valid {
			plan = models.SubscriptionPlan{
				ID:            int(planID.Int64),
				Title:         planTitle.String,
				Terms:         planTerms.String,
				Status:        planStatus.Bool,
				DownloadLimit: int(planDL.Int64),
				ExpiresAt:     int(planExpiresAt.Int64),
				CreatedAt:     planCreatedAt.Time,
				UpdatedAt:     planUpdatedAt.Time,
			}
			sub.PlanDetails = &plan
		}

		user.CurrentSubscription = &sub
	}

	baseURL, _ := url.Parse(models.APIEndPoint)
	baseURL.Path = path.Join(baseURL.Path, "public", "profile", user.AvatarID)
	user.AvatarURL = baseURL.String()
	return &user, nil
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
			u.total_expenses,
			u.address,
			u.subscription_id,
			u.created_at,
			u.updated_at,
			
			s.id,
			s.user_id,
			s.subscription_plans_id,
			s.payment_amount,
			s.payment_time,
			s.total_downloads,
			s.status,
			s.created_at,
			s.updated_at,
			
			sp.id,
			sp.title,
			sp.terms,
			sp.status,
			sp.download_limit,
			sp.expires_at,
			sp.created_at,
			sp.updated_at
		FROM users u
		LEFT JOIN subscriptions s ON u.subscription_id = s.id
		LEFT JOIN subscription_plans sp ON s.subscription_plans_id = sp.id
		WHERE u.username = $1;
	`

	var (
		user models.User
		sub  models.Subscription
		plan models.SubscriptionPlan

		// Nullable fields
		subID             sql.NullInt64
		subUserID         sql.NullInt64
		subPlanID         sql.NullInt64
		subPaymentAmount  sql.NullFloat64
		subPaymentTime    sql.NullTime
		subTotalDownloads sql.NullInt64
		subStatus         sql.NullBool
		subCreatedAt      sql.NullTime
		subUpdatedAt      sql.NullTime

		planID        sql.NullInt64
		planTitle     sql.NullString
		planTerms     sql.NullString
		planStatus    sql.NullBool
		planDL        sql.NullInt64
		planExpiresAt sql.NullInt64
		planCreatedAt sql.NullTime
		planUpdatedAt sql.NullTime
	)

	err := r.db.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.Name,
		&user.Email,
		&user.Mobile,
		&user.AvatarID,
		&user.Status,
		&user.Role,
		&user.TotalEarnings,
		&user.TotalWithdraw,
		&user.TotalExpenses,
		&user.Address,
		&user.SubscriptionID,
		&user.CreatedAt,
		&user.UpdatedAt,

		&subID,
		&subUserID,
		&subPlanID,
		&subPaymentAmount,
		&subPaymentTime,
		&subTotalDownloads,
		&subStatus,
		&subCreatedAt,
		&subUpdatedAt,

		&planID,
		&planTitle,
		&planTerms,
		&planStatus,
		&planDL,
		&planExpiresAt,
		&planCreatedAt,
		&planUpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Build subscription if exists
	if subID.Valid {
		sub = models.Subscription{
			ID:                 int(subID.Int64),
			UserID:             int(subUserID.Int64),
			SubscriptionPlanID: int(subPlanID.Int64),
			PaymentAmount:      subPaymentAmount.Float64,
			PaymentTime:        subPaymentTime.Time,
			TotalDownloads:     int(subTotalDownloads.Int64),
			Status:             subStatus.Bool,
			CreatedAt:          subCreatedAt.Time,
			UpdatedAt:          subUpdatedAt.Time,
		}

		// Build subscription plan if exists
		if planID.Valid {
			plan = models.SubscriptionPlan{
				ID:            int(planID.Int64),
				Title:         planTitle.String,
				Terms:         planTerms.String,
				Status:        planStatus.Bool,
				DownloadLimit: int(planDL.Int64),
				ExpiresAt:     int(planExpiresAt.Int64),
				CreatedAt:     planCreatedAt.Time,
				UpdatedAt:     planUpdatedAt.Time,
			}
			sub.PlanDetails = &plan
		}

		user.CurrentSubscription = &sub

	}

	baseURL, _ := url.Parse(models.APIEndPoint)
	baseURL.Path = path.Join(baseURL.Path, "public", "profile", user.AvatarID)
	user.AvatarURL = baseURL.String()
	return &user, nil
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
			u.total_expenses,
			u.address,
			u.subscription_id,
			u.created_at,
			u.updated_at,
			
			s.id,
			s.user_id,
			s.subscription_plans_id,
			s.payment_amount,
			s.payment_time,
			s.total_downloads,
			s.status,
			s.created_at,
			s.updated_at,
			
			sp.id,
			sp.title,
			sp.terms,
			sp.status,
			sp.download_limit,
			sp.expires_at,
			sp.created_at,
			sp.updated_at
		FROM users u
		LEFT JOIN subscriptions s ON u.subscription_id = s.id AND s.status = true
		LEFT JOIN subscription_plans sp ON s.subscription_plans_id = sp.id
		WHERE u.email = $1;
	`

	var (
		user models.User
		sub  models.Subscription
		plan models.SubscriptionPlan

		// Nullable fields
		subID     sql.NullInt64
		subUserID sql.NullInt64
		subPlanID sql.NullInt64

		subPaymentAmount  sql.NullFloat64
		subPaymentTime    sql.NullTime
		subTotalDownloads sql.NullInt64
		subStatus         sql.NullBool
		subCreatedAt      sql.NullTime
		subUpdatedAt      sql.NullTime

		planID        sql.NullInt64
		planTitle     sql.NullString
		planTerms     sql.NullString
		planStatus    sql.NullBool
		planDL        sql.NullInt64
		planExpiresAt sql.NullInt64
		planCreatedAt sql.NullTime
		planUpdatedAt sql.NullTime
	)

	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.Name,
		&user.Email,
		&user.Mobile,
		&user.AvatarID,
		&user.Status,
		&user.Role,
		&user.TotalEarnings,
		&user.TotalWithdraw,
		&user.TotalExpenses,
		&user.Address,
		&user.SubscriptionID,
		&user.CreatedAt,
		&user.UpdatedAt,

		&subID,
		&subUserID,
		&subPlanID,
		&subPaymentAmount,
		&subPaymentTime,
		&subTotalDownloads,
		&subStatus,
		&subCreatedAt,
		&subUpdatedAt,

		&planID,
		&planTitle,
		&planTerms,
		&planStatus,
		&planDL,
		&planExpiresAt,
		&planCreatedAt,
		&planUpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Build subscription if exists
	if subID.Valid {
		sub = models.Subscription{
			ID:                 int(subID.Int64),
			UserID:             int(subUserID.Int64),
			SubscriptionPlanID: int(subPlanID.Int64),
			PaymentAmount:      subPaymentAmount.Float64,
			PaymentTime:        subPaymentTime.Time,
			TotalDownloads:     int(subTotalDownloads.Int64),
			Status:             subStatus.Bool,
			CreatedAt:          subCreatedAt.Time,
			UpdatedAt:          subUpdatedAt.Time,
		}

		// Build subscription plan if exists
		if planID.Valid {
			plan = models.SubscriptionPlan{
				ID:            int(planID.Int64),
				Title:         planTitle.String,
				Terms:         planTerms.String,
				Status:        planStatus.Bool,
				DownloadLimit: int(planDL.Int64),
				ExpiresAt:     int(planExpiresAt.Int64),
				CreatedAt:     planCreatedAt.Time,
				UpdatedAt:     planUpdatedAt.Time,
			}
			sub.PlanDetails = &plan
		}

		user.CurrentSubscription = &sub
	}

	baseURL, _ := url.Parse(models.APIEndPoint)
	baseURL.Path = path.Join(baseURL.Path, "public", "profile", user.AvatarID)
	user.AvatarURL = baseURL.String()
	return &user, nil
}

func (r *UserRepo) Update(ctx context.Context, user *models.User) error {
	query := `
	UPDATE users
	SET 
		username = $2, password = $3, name = $4, avatar_url = $5, status = $6, 
		role = $7, email = $8, mobile = $9, total_earnings = $10, total_withdraw = $11, 
		total_expenses = $12, address = $13, subscription_id = $14, updated_at = $15
	WHERE id = $1`
	_, err := r.db.Exec(ctx, query,
		user.ID, user.Username, user.Password, user.Name, user.AvatarID, user.Status,
		user.Role, user.Email, user.Mobile, user.TotalEarnings, user.TotalWithdraw,
		user.TotalExpenses, user.Address, user.SubscriptionID, time.Now(),
	)
	return err
}
func (r *UserRepo) UpdateBasicInfo(ctx context.Context, user *models.User) error {
	query := `
	UPDATE users
	SET 
		name = $1, email = $2, mobile = $3, address = $4, updated_at = $5
	WHERE id = $6`
	_, err := r.db.Exec(ctx, query,
		user.Name, user.Email, user.Mobile, user.Address, time.Now(), user.ID,
	)
	return err
}
func (r *UserRepo) UpdateProfileAvatarURL(ctx context.Context, id int, avatarID string) error {
	query := `
	UPDATE users
	SET 
		avatar_url = $1, updated_at = $2
	WHERE id = $3`
	_, err := r.db.Exec(ctx, query,
		avatarID, time.Now(), id,
	)
	return err
}
func (r *UserRepo) UpdateSubscriptionPlanByUserID(ctx context.Context, subscriptionID, userID int) error {
	query := `
	UPDATE users
	SET 
		subscription_id = $1,
		total_expenses = total_expenses + COALESCE((
			SELECT price FROM subscription_plans WHERE id = $1
		), 0),
		updated_at = $2
	WHERE id = $3`

	_, err := r.db.Exec(ctx, query,
		subscriptionID,
		time.Now(),
		userID,
	)
	return err
}

func (r *UserRepo) DeleteByID(ctx context.Context, id int) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *UserRepo) Deactivate(ctx context.Context, id int, status bool) error {
	query := `
	UPDATE users
	SET 
		status = $2, updated_at = $3
	WHERE id = $1`
	_, err := r.db.Exec(ctx, query,
		id, status, time.Now(),
	)
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
			&user.ID, &user.Username, &user.Password, &user.Name, &user.AvatarID, &user.Status,
			&user.Role, &user.Email, &user.Mobile, &user.TotalEarnings, &user.TotalWithdraw,
			&user.TotalExpenses, &user.Address, &user.SubscriptionID, &user.CreatedAt, &user.UpdatedAt,
		); err != nil {
			return nil, err
		}
		baseURL, _ := url.Parse(models.APIEndPoint)
		baseURL.Path = path.Join(baseURL.Path, "public", "profile", user.AvatarID)
		user.AvatarURL = baseURL.String()
		users = append(users, &user)
	}
	return users, nil
}

func (r *UserRepo) DecrementDownloadLimit(ctx context.Context, userID int) error {
	// Decrement download count in subscription
	query := `
		UPDATE subscriptions 
		SET total_downloads = total_downloads - 1 
		WHERE user_id = $1 AND total_downloads > 0
	`
	_, err := r.db.Exec(ctx, query, userID)
	return err
}

func (r *UserRepo) IncrementDownloadCounts(ctx context.Context, userID int) error {
	// Increment download count in subscription
	query := `
		UPDATE subscriptions 
		SET total_downloads = total_downloads + 1 
		WHERE user_id = $1
	`
	_, err := r.db.Exec(ctx, query, userID)
	return err
}
