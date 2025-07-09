package repositories

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samiulice/photostock/internal/models"
)

// ============================== DownloadHistoryRepo Repository ==============================
type DownloadHistoryRepo struct {
	db *pgxpool.Pool
}

func NewDownloadHistoryRepo(db *pgxpool.Pool) *DownloadHistoryRepo {
	return &DownloadHistoryRepo{db: db}
}

func (r *DownloadHistoryRepo) Create(ctx context.Context, h *models.DownloadHistory) error {
	// id SERIAL PRIMARY KEY,
	// media_uuid VARCHAR(255) NOT NULL DEFAULT '',
	// user_id INTEGER NOT NULL,
	// Downloaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	// created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
	// Updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
	// CONSTRAINT fk_Downloadd_media FOREIGN KEY (media_uuid)
	//     REFERENCES medias (media_uuid) ON DELETE CASCADE,
	// CONSTRAINT fk_Download_user FOREIGN KEY (user_id)
	//     REFERENCES users (id) ON DELETE CASCADE
	query := `
	INSERT INTO download_history (media_uuid, user_id, downloaded_at)
	VALUES ($1,$2,$3)
	RETURNING id`
	return r.db.QueryRow(ctx, query, h.MediaUUID, h.UserID, h.DownloadedAt).Scan(&h.ID)
}

func (r *DownloadHistoryRepo) GetByID(ctx context.Context, id int) (*models.DownloadHistory, error) {
	query := `
	SELECT id, media_uuid, user_id, downloaded_at, created_at, updated_at
	FROM download_history
	WHERE id = $1`
	h := &models.DownloadHistory{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&h.ID, &h.MediaUUID, &h.UserID, &h.DownloadedAt, &h.CreatedAt, &h.UpdatedAt,
	)
	return h, err
}

func (r *DownloadHistoryRepo) Update(ctx context.Context, h *models.DownloadHistory) error {
	query := `
	update download_history
	SET media_uuid = $1, user_id = $2, downloaded_at = $3, updated_at = $4
	WHERE id = $5`
	_, err := r.db.Exec(ctx, query, h.MediaUUID, h.UserID, h.DownloadedAt, time.Now(), h.ID)
	return err
}

func (r *DownloadHistoryRepo) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM download_history WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *DownloadHistoryRepo) GetAll(ctx context.Context) ([]*models.DownloadHistory, error) {
	query := `SELECT id, media_uuid, user_id, downloaded_at, created_at, updated_at FROM download_history`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []*models.DownloadHistory
	for rows.Next() {
		var h models.DownloadHistory
		if err := rows.Scan(
			&h.ID, &h.MediaUUID, &h.UserID, &h.DownloadedAt, &h.CreatedAt, &h.UpdatedAt,
		); err != nil {
			return nil, err
		}
		history = append(history, &h)
	}
	return history, nil
}

func (r *DownloadHistoryRepo) GetAllByUserID(ctx context.Context, id int) ([]*models.DownloadHistory, error) {
	query := `SELECT id, media_uuid, user_id, downloaded_at, created_at, updated_at FROM download_history WHERE user_id = $1`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []*models.DownloadHistory
	for rows.Next() {
		var h models.DownloadHistory
		if err := rows.Scan(&h.ID, &h.MediaUUID, &h.UserID, &h.DownloadedAt, &h.CreatedAt, &h.UpdatedAt, &h.ID); err != nil {
			return nil, err
		}
		history = append(history, &h)
	}
	return history, nil
}
