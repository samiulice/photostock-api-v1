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
	INSERT INTO media_categories (name)
	VALUES ($1)
	RETURNING id`
	return r.db.QueryRow(ctx, query, c.Name).Scan(&c.ID)
}

func (r *MediaCategoryRepo) GetByID(ctx context.Context, id int) (*models.MediaCategory, error) {
	query := `
	SELECT id, name, created_at, updated_at
	FROM media_categories
	WHERE id = $1`
	c := &models.MediaCategory{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.Name, &c.CreatedAt, &c.UpdatedAt,
	)
	return c, err
}

func (r *MediaCategoryRepo) Update(ctx context.Context, c *models.MediaCategory) error {
	query := `
	UPDATE media_categories
	SET name = $2, updated_at = $3
	WHERE id = $1`
	_, err := r.db.Exec(ctx, query, c.ID, c.Name, time.Now())
	return err
}

func (r *MediaCategoryRepo) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM media_categories WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *MediaCategoryRepo) GetAll(ctx context.Context) ([]*models.MediaCategory, error) {
	query := `SELECT id, name, created_at, updated_at FROM media_categories`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []*models.MediaCategory
	for rows.Next() {
		var c models.MediaCategory
		if err := rows.Scan(
			&c.ID, &c.Name, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, err
		}
		categories = append(categories, &c)
	}
	return categories, nil
}
