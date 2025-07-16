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
	query := `
	INSERT INTO upload_history (media_uuid, user_id, file_type, file_name, file_size, uploaded_at)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id`
	return r.db.QueryRow(ctx, query,
		h.MediaUUID,
		h.UserID,
		h.FileType,
		h.FileName,
		h.FileSize,
		h.UploadedAt,
	).Scan(&h.ID)
}

func (r *UploadHistoryRepo) GetByID(ctx context.Context, id int) (*models.UploadHistory, error) {
	query := `
	SELECT id, media_uuid, user_id, file_type, file_name, file_size, uploaded_at, created_at, updated_at
	FROM upload_history
	WHERE id = $1`
	h := &models.UploadHistory{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&h.ID,
		&h.MediaUUID,
		&h.UserID,
		&h.FileType,
		&h.FileName,
		&h.FileSize,
		&h.UploadedAt,
		&h.CreatedAt,
		&h.UpdatedAt,
	)
	return h, err
}

func (r *UploadHistoryRepo) Update(ctx context.Context, h *models.UploadHistory) error {
	query := `
	UPDATE upload_history
	SET media_uuid = $1, user_id = $2, file_type = $3, file_name = $4, file_size = $5,
	uploaded_at = $6, updated_at = $7
	WHERE id = $8`
	_, err := r.db.Exec(ctx, query,
		h.MediaUUID,
		h.UserID,
		h.FileType,
		h.FileName,
		h.FileSize,
		h.UploadedAt,
		time.Now(),
		h.ID,
	)
	return err
}

func (r *UploadHistoryRepo) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM upload_history WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *UploadHistoryRepo) GetAll(ctx context.Context) ([]*models.UploadHistory, error) {
	query := `SELECT id, media_uuid, user_id, file_type, file_name, file_size, uploaded_at, created_at, updated_at FROM upload_history`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []*models.UploadHistory
	for rows.Next() {
		var h models.UploadHistory
		if err := rows.Scan(
			&h.ID,
			&h.MediaUUID,
			&h.UserID,
			&h.FileType,
			&h.FileName,
			&h.FileSize,
			&h.UploadedAt,
			&h.CreatedAt,
			&h.UpdatedAt,
		); err != nil {
			return nil, err
		}
		history = append(history, &h)
	}
	return history, nil
}

func (r *UploadHistoryRepo) GetAllByUserID(ctx context.Context, id int) ([]*models.UploadHistory, error) {
	query := `SELECT id, media_uuid, user_id, file_type, file_name, file_size, uploaded_at, created_at, updated_at FROM upload_history WHERE user_id = $1`
	rows, err := r.db.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []*models.UploadHistory
	for rows.Next() {
		var h models.UploadHistory
		if err := rows.Scan(
			&h.ID,
			&h.MediaUUID,
			&h.UserID,
			&h.FileType,
			&h.FileName,
			&h.FileSize,
			&h.UploadedAt,
			&h.CreatedAt,
			&h.UpdatedAt,
		); err != nil {
			return nil, err
		}
		history = append(history, &h)
	}
	return history, nil
}
