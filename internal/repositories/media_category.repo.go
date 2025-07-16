package repositories

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samiulice/photostock/internal/models"
)

// ============================== MediaCategory Repository ==============================
type MediaCategoryRepo struct {
	db *pgxpool.Pool
}

func NewMediaCategoryRepo(db *pgxpool.Pool) *MediaCategoryRepo {
	return &MediaCategoryRepo{db: db}
}

func (r *MediaCategoryRepo) Create(ctx context.Context, c *models.MediaCategory) error {
	query := `
	INSERT INTO media_categories (name, thumbnail_uuid)
	VALUES ($1,$2)
	RETURNING id`
	return r.db.QueryRow(ctx, query, c.Name, c.ThumbnailURL).Scan(&c.ID)
}

func (r *MediaCategoryRepo) GetByID(ctx context.Context, id int) (*models.MediaCategory, error) {
	query := `
	SELECT id, name, thumbnail_uuid, total_uploads, total_downloads, created_at, updated_at
	FROM media_categories
	WHERE id = $1`
	c := &models.MediaCategory{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.Name, &c.ThumbnailURL, &c.UploadCount, &c.DownloadCount, &c.CreatedAt, &c.UpdatedAt,
	)
	return c, err
}

func (r *MediaCategoryRepo) Update(ctx context.Context, c *models.MediaCategory) error {
	query := `
	UPDATE media_categories
	SET name = $2, thumbnail_uuid = $3, updated_at = $4
	WHERE id = $1`
	_, err := r.db.Exec(ctx, query, c.ID, c.Name, c.ThumbnailURL, time.Now())
	return err
}

// IncrementUploads increases total_uploads by 1
func (r *MediaCategoryRepo) IncrementUploads(ctx context.Context, categoryID int64) error {
	query := `
		UPDATE media_categories
		SET total_uploads = total_uploads + 1,
		    updated_at = $2
		WHERE id = $1`
	_, err := r.db.Exec(ctx, query, categoryID, time.Now())
	return err
}

// DecrementUploads decreases total_uploads by 1, ensuring it doesn't go below 0
func (r *MediaCategoryRepo) DecrementUploads(ctx context.Context, categoryID int64) error {
	query := `
		UPDATE media_categories
		SET total_uploads = GREATEST(total_uploads - 1, 0),
		    updated_at = $2
		WHERE id = $1`
	_, err := r.db.Exec(ctx, query, categoryID, time.Now())
	return err
}

// IncrementDownloads increases total_downloads by 1
func (r *MediaCategoryRepo) IncrementDownloads(ctx context.Context, categoryID int64) error {
	query := `
		UPDATE media_categories
		SET total_downloads = total_downloads + 1,
		    updated_at = $2
		WHERE id = $1`
	_, err := r.db.Exec(ctx, query, categoryID, time.Now())
	return err
}

// DecrementDownloads decreases total_downloads by 1, ensuring it doesn't go below 0
func (r *MediaCategoryRepo) DecrementDownloads(ctx context.Context, categoryID int64) error {
	query := `
		UPDATE media_categories
		SET total_downloads = GREATEST(total_downloads - 1, 0),
		    updated_at = $2
		WHERE id = $1`
	_, err := r.db.Exec(ctx, query, categoryID, time.Now())
	return err
}

func (r *MediaCategoryRepo) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM media_categories WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *MediaCategoryRepo) GetAll(ctx context.Context) ([]*models.MediaCategory, error) {
	query := `SELECT id, name, thumbnail_uuid, total_uploads, total_downloads, created_at, updated_at FROM media_categories`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []*models.MediaCategory
	for rows.Next() {
		var c models.MediaCategory
		if err := rows.Scan(
			&c.ID, &c.Name, &c.ThumbnailURL, &c.UploadCount, &c.DownloadCount, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, err
		}
		categories = append(categories, &c)
	}
	return categories, nil
}


