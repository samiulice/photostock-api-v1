package repositories

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samiulice/photostock/internal/models"
)

// ============================== DownloadHistory Repository ==============================
type DownloadHistoryRepo struct {
	db *pgxpool.Pool
}

func NewDownloadHistoryRepo(db *pgxpool.Pool) *DownloadHistoryRepo {
	return &DownloadHistoryRepo{db: db}
}

func (r *DownloadHistoryRepo) Create(ctx context.Context, dh *models.DownloadHistory) error {
	query := `
	INSERT INTO download_histories (
		product_id, user_id, price, downloaded_at, created_at, updated_at
	) VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id`
	return r.db.QueryRow(ctx, query,
		dh.ProductID, dh.UserID, dh.Price, dh.DownloadedAt, time.Now(), time.Now(),
	).Scan(&dh.ID)
}

func (r *DownloadHistoryRepo) GetByID(ctx context.Context, id int) (*models.DownloadHistory, error) {
	query := `
	SELECT id, product_id, user_id, price, downloaded_at, created_at, updated_at
	FROM download_histories
	WHERE id = $1`
	dh := &models.DownloadHistory{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&dh.ID, &dh.ProductID, &dh.UserID, &dh.Price, &dh.DownloadedAt, &dh.CreatedAt, &dh.UpdatedAt,
	)
	return dh, err
}

func (r *DownloadHistoryRepo) Update(ctx context.Context, dh *models.DownloadHistory) error {
	query := `
	UPDATE download_histories
	SET 
		product_id = $2, user_id = $3, price = $4, downloaded_at = $5, updated_at = $6
	WHERE id = $1`
	_, err := r.db.Exec(ctx, query,
		dh.ID, dh.ProductID, dh.UserID, dh.Price, dh.DownloadedAt, time.Now(),
	)
	return err
}

func (r *DownloadHistoryRepo) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM download_histories WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *DownloadHistoryRepo) GetAll(ctx context.Context) ([]models.DownloadHistory, error) {
	query := `
	SELECT id, product_id, user_id, price, downloaded_at, created_at, updated_at
	FROM download_histories`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []models.DownloadHistory
	for rows.Next() {
		var dh models.DownloadHistory
		if err := rows.Scan(
			&dh.ID, &dh.ProductID, &dh.UserID, &dh.Price, &dh.DownloadedAt, &dh.CreatedAt, &dh.UpdatedAt,
		); err != nil {
			return nil, err
		}
		history = append(history, dh)
	}
	return history, nil
}
