package repositories

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samiulice/photostock/internal/models"
)

type DownloadHistoryRepo struct {
	db *pgxpool.Pool
}

func NewDownloadHistoryRepo(db *pgxpool.Pool) *DownloadHistoryRepo {
	return &DownloadHistoryRepo{db: db}
}

// Create inserts a new download history record
func (r *DownloadHistoryRepo) Create(ctx context.Context, h *models.DownloadHistory) error {
	query := `
	INSERT INTO download_history (
		media_uuid, user_id, file_type, file_ext, file_name, file_size, resolution, downloaded_at
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	RETURNING id`
	return r.db.QueryRow(ctx, query,
		h.MediaUUID, h.UserID,
		h.FileType, h.FileExt, h.FileName, h.FileSize, h.Resolution,
		h.DownloadedAt,
	).Scan(&h.ID)
}

// GetByID retrieves a download history by its ID
func (r *DownloadHistoryRepo) GetByID(ctx context.Context, id int) (*models.DownloadHistory, error) {
	query := `
	SELECT id, media_uuid, user_id, file_type, file_ext, file_name, file_size, resolution,
	       downloaded_at, created_at, updated_at
	FROM download_history
	WHERE id = $1`
	h := &models.DownloadHistory{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&h.ID, &h.MediaUUID, &h.UserID, &h.FileType, &h.FileExt, &h.FileName, &h.FileSize, &h.Resolution,
		&h.DownloadedAt, &h.CreatedAt, &h.UpdatedAt,
	)
	return h, err
}

// Update modifies a download history record
func (r *DownloadHistoryRepo) Update(ctx context.Context, h *models.DownloadHistory) error {
	query := `
	UPDATE download_history
	SET media_uuid = $1, user_id = $2,
		file_type = $3, file_ext = $4, file_name = $5, file_size = $6, resolution = $7,
		downloaded_at = $8, updated_at = $9
	WHERE id = $10`
	_, err := r.db.Exec(ctx, query,
		h.MediaUUID, h.UserID,
		h.FileType, h.FileExt, h.FileName, h.FileSize, h.Resolution,
		h.DownloadedAt, time.Now(), h.ID,
	)
	return err
}

// Delete removes a download history record by ID
func (r *DownloadHistoryRepo) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM download_history WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// GetAll returns all download history records
func (r *DownloadHistoryRepo) GetAll(ctx context.Context) ([]*models.DownloadHistory, error) {
	query := `
	SELECT id, media_uuid, user_id, file_type, file_ext, file_name, file_size, resolution,
	       downloaded_at, created_at, updated_at
	FROM download_history`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []*models.DownloadHistory
	for rows.Next() {
		var h models.DownloadHistory
		if err := rows.Scan(
			&h.ID, &h.MediaUUID, &h.UserID, &h.FileType, &h.FileExt, &h.FileName, &h.FileSize, &h.Resolution,
			&h.DownloadedAt, &h.CreatedAt, &h.UpdatedAt,
		); err != nil {
			return nil, err
		}
		history = append(history, &h)
	}
	return history, nil
}

// GetAllByUserID returns all download history records for a specific user
func (r *DownloadHistoryRepo) GetAllByUserID(ctx context.Context, userID int) ([]*models.DownloadHistory, error) {
	query := `
	SELECT id, media_uuid, user_id, file_type, file_ext, file_name, file_size, resolution,
	       downloaded_at, created_at, updated_at
	FROM download_history
	WHERE user_id = $1`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []*models.DownloadHistory
	for rows.Next() {
		var h models.DownloadHistory
		if err := rows.Scan(
			&h.ID, &h.MediaUUID, &h.UserID, &h.FileType, &h.FileExt, &h.FileName, &h.FileSize, &h.Resolution,
			&h.DownloadedAt, &h.CreatedAt, &h.UpdatedAt,
		); err != nil {
			return nil, err
		}
		history = append(history, &h)
	}
	return history, nil
}
