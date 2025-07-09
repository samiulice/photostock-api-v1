package repositories

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

// DBRepository contains all individual repositories
type DBRepository struct {
	SubscriptionTypeRepo *SubscriptionTypeRepo
	MediaCategoryRepo    *MediaCategoryRepo
	UserRepo             *UserRepo
	// SubscriptionRepo     *SubscriptionRepo
	MediaRepo            *MediaRepo
	DownloadHistoryRepo  *DownloadHistoryRepo
	UploadHistoryRepo  *UploadHistoryRepo
}

// NewDBRepository initializes all repositories with a shared connection pool
// NewDBRepository initializes all repositories with a shared connection pool
func NewDBRepository(db *pgxpool.Pool) *DBRepository {
	return &DBRepository{
		SubscriptionTypeRepo: NewSubscriptionPlanRepo(db),
		MediaCategoryRepo:    NewMediaCategoryRepo(db),
		UserRepo:             NewUserRepo(db),
		// SubscriptionRepo:     NewSubscriptionRepo(db),
		MediaRepo:            NewMediaRepo(db),
		DownloadHistoryRepo:  NewDownloadHistoryRepo(db),
		UploadHistoryRepo:  NewUploadHistoryRepo(db),
	}
}
