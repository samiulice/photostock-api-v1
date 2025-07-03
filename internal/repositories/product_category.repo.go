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

func (r *MediaCategoryRepo) Create(ctx context.Context, pc *models.MediaCategory) error {
	query := `
	INSERT INTO media_categories (name, created_at, updated_at)
	VALUES ($1, $2, $3)
	RETURNING id`
	return r.db.QueryRow(ctx, query, pc.Name, time.Now(), time.Now()).Scan(&pc.ID)
}

func (r *MediaCategoryRepo) GetByID(ctx context.Context, id int) (*models.MediaCategory, error) {
	query := `
	SELECT id, name, created_at, updated_at
	FROM media_categories
	WHERE id = $1`
	pc := &models.MediaCategory{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&pc.ID, &pc.Name, &pc.CreatedAt, &pc.UpdatedAt,
	)
	return pc, err
}

func (r *MediaCategoryRepo) Update(ctx context.Context, pc *models.MediaCategory) error {
	query := `
	UPDATE media_categories
	SET name = $2, updated_at = $3
	WHERE id = $1`
	_, err := r.db.Exec(ctx, query, pc.ID, pc.Name, time.Now())
	return err
}

func (r *MediaCategoryRepo) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM media_categories WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *MediaCategoryRepo) GetAll(ctx context.Context) ([]models.MediaCategory, error) {
	query := `SELECT id, name, created_at, updated_at FROM media_categories`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.MediaCategory
	for rows.Next() {
		var pc models.MediaCategory
		if err := rows.Scan(
			&pc.ID, &pc.Name, &pc.CreatedAt, &pc.UpdatedAt,
		); err != nil {
			return nil, err
		}
		categories = append(categories, pc)
	}
	return categories, nil
}
