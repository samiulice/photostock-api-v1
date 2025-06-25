package repositories

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samiulice/photostock/internal/models"
)

// ============================== ProductCategory Repository ==============================
type ProductCategoryRepo struct {
	db *pgxpool.Pool
}

func NewProductCategoryRepo(db *pgxpool.Pool) *ProductCategoryRepo {
	return &ProductCategoryRepo{db: db}
}

func (r *ProductCategoryRepo) Create(ctx context.Context, pc *models.ProductCategory) error {
	query := `
	INSERT INTO product_categories (name, created_at, updated_at)
	VALUES ($1, $2, $3)
	RETURNING id`
	return r.db.QueryRow(ctx, query, pc.Name, time.Now(), time.Now()).Scan(&pc.ID)
}

func (r *ProductCategoryRepo) GetByID(ctx context.Context, id int) (*models.ProductCategory, error) {
	query := `
	SELECT id, name, created_at, updated_at
	FROM product_categories
	WHERE id = $1`
	pc := &models.ProductCategory{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&pc.ID, &pc.Name, &pc.CreatedAt, &pc.UpdatedAt,
	)
	return pc, err
}

func (r *ProductCategoryRepo) Update(ctx context.Context, pc *models.ProductCategory) error {
	query := `
	UPDATE product_categories
	SET name = $2, updated_at = $3
	WHERE id = $1`
	_, err := r.db.Exec(ctx, query, pc.ID, pc.Name, time.Now())
	return err
}

func (r *ProductCategoryRepo) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM product_categories WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *ProductCategoryRepo) GetAll(ctx context.Context) ([]models.ProductCategory, error) {
	query := `SELECT id, name, created_at, updated_at FROM product_categories`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.ProductCategory
	for rows.Next() {
		var pc models.ProductCategory
		if err := rows.Scan(
			&pc.ID, &pc.Name, &pc.CreatedAt, &pc.UpdatedAt,
		); err != nil {
			return nil, err
		}
		categories = append(categories, pc)
	}
	return categories, nil
}