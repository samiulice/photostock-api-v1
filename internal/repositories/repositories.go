package repositories

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

// DBRepository contains all individual repositories
type DBRepository struct {
	SubscriptionTypeRepo *SubscriptionTypeRepo
	ProductCategoryRepo  *ProductCategoryRepo
	UserRepo             *UserRepo
	SubscriptionRepo     *SubscriptionRepo
	ProductRepo          *ProductRepo
	DownloadHistoryRepo  *DownloadHistoryRepo
}

// NewDBRepository initializes all repositories with a shared connection pool
func NewDBRepository(db *pgxpool.Pool) *DBRepository {
	return &DBRepository{
		SubscriptionTypeRepo: NewSubscriptionPlanRepo(db),
		ProductCategoryRepo:  NewProductCategoryRepo(db),
		UserRepo:             NewUserRepo(db),
		SubscriptionRepo:     NewSubscriptionRepo(db),
		ProductRepo:          NewProductRepo(db),
		DownloadHistoryRepo:  NewDownloadHistoryRepo(db),
	}
}
