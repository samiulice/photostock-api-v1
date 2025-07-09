package repositories

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samiulice/photostock/internal/models"
)

// ============================== UploadHistoryRepo Repository ==============================
type UploadHistoryRepo struct {
	db *pgxpool.Pool
}

func NewUploadHistoryRepo(db *pgxpool.Pool) *UploadHistoryRepo {
	return &UploadHistoryRepo{db: db}
}

func (r *UploadHistoryRepo) Create(ctx context.Context, h *models.UploadHistory) error {
	// id SERIAL PRIMARY KEY,
	// media_uuid VARCHAR(255) NOT NULL DEFAULT '',
	// user_id INTEGER NOT NULL,
	// uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	// created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
	// updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
	// CONSTRAINT fk_uploadd_media FOREIGN KEY (media_uuid)
	//     REFERENCES medias (media_uuid) ON DELETE CASCADE,
	// CONSTRAINT fk_upload_user FOREIGN KEY (user_id)
	//     REFERENCES users (id) ON DELETE CASCADE
	query := `
	INSERT INTO upload_history (media_uuid, user_id, uploaded_at)
	VALUES ($1,$2,$3)
	RETURNING id`
	return r.db.QueryRow(ctx, query, h.MediaUUID, h.UserID, h.UploadedAt).Scan(&h.ID)
}

func (r *UploadHistoryRepo) GetByID(ctx context.Context, id int) (*models.UploadHistory, error) {
	query := `
	SELECT id, media_uuid, user_id, uploaded_at, created_at, updated_at
	FROM upload_history
	WHERE id = $1`
	h := &models.UploadHistory{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&h.ID, &h.MediaUUID, &h.UserID, &h.UploadedAt, &h.CreatedAt, &h.UpdatedAt,
	)
	return h, err
}

func (r *UploadHistoryRepo) Update(ctx context.Context, h *models.UploadHistory) error {
	query := `
	UPDATE upload_history
	SET media_uuid = $1, user_id = $2, uploaded_at = $3, updated_at = $4
	WHERE id = $5`
	_, err := r.db.Exec(ctx, query, h.MediaUUID, h.UserID, h.UploadedAt, time.Now(), h.ID)
	return err
}

func (r *UploadHistoryRepo) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM upload_history WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *UploadHistoryRepo) GetAll(ctx context.Context) ([]*models.UploadHistory, error) {
	query := `SELECT id, media_uuid, user_id, uploaded_at, created_at, updated_at FROM upload_history`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []*models.UploadHistory
	for rows.Next() {
		var h models.UploadHistory
		if err := rows.Scan(
			&h.ID, &h.MediaUUID, &h.UserID, &h.UploadedAt, &h.CreatedAt, &h.UpdatedAt,
		); err != nil {
			return nil, err
		}
		history = append(history, &h)
	}
	return history, nil
}

func (r *UploadHistoryRepo) GetAllByUserID(ctx context.Context, id int) ([]*models.UploadHistory, error) {
	query := `SELECT id, media_uuid, user_id, uploaded_at, created_at, updated_at FROM upload_history WHERE user_id = $1`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []*models.UploadHistory
	for rows.Next() {
		var h models.UploadHistory
		if err := rows.Scan(&h.ID, &h.MediaUUID, &h.UserID, &h.UploadedAt, &h.CreatedAt, &h.UpdatedAt, &h.ID); err != nil {
			return nil, err
		}
		history = append(history, &h)
	}
	return history, nil
}
