package repositories

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samiulice/photostock/internal/models"
)

// ============================== Product Repository ==============================
type ProductRepo struct {
	db *pgxpool.Pool
}

func NewProductRepo(db *pgxpool.Pool) *ProductRepo {
	return &ProductRepo{db: db}
}

func (r *ProductRepo) Create(ctx context.Context, product *models.Product) error {
	query := `
	INSERT INTO products (
		product_id, product_title, description, product_url, category_id, 
		mrp, max_discount, total_earnings, created_at, updated_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	RETURNING id`
	return r.db.QueryRow(ctx, query,
		product.ProductID, product.ProductTitle, product.Description, product.ProductURL,
		product.CategoryID, product.MRP, product.MaxDiscount, product.TotalEarnings,
		time.Now(), time.Now(),
	).Scan(&product.ID)
}

func (r *ProductRepo) GetByID(ctx context.Context, id int) (*models.Product, error) {
	query := `
	SELECT id, product_id, product_title, description, product_url, category_id, 
		mrp, max_discount, total_earnings, created_at, updated_at
	FROM products
	WHERE id = $1`
	product := &models.Product{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&product.ID, &product.ProductID, &product.ProductTitle, &product.Description,
		&product.ProductURL, &product.CategoryID, &product.MRP, &product.MaxDiscount,
		&product.TotalEarnings, &product.CreatedAt, &product.UpdatedAt,
	)
	return product, err
}

func (r *ProductRepo) Update(ctx context.Context, product *models.Product) error {
	query := `
	UPDATE products
	SET 
		product_id = $2, product_title = $3, description = $4, product_url = $5, 
		category_id = $6, mrp = $7, max_discount = $8, total_earnings = $9, updated_at = $10
	WHERE id = $1`
	_, err := r.db.Exec(ctx, query,
		product.ID, product.ProductID, product.ProductTitle, product.Description,
		product.ProductURL, product.CategoryID, product.MRP, product.MaxDiscount,
		product.TotalEarnings, time.Now(),
	)
	return err
}

func (r *ProductRepo) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM products WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *ProductRepo) GetAll(ctx context.Context) ([]models.Product, error) {
	query := `
	SELECT id, product_id, product_title, description, product_url, category_id, 
		mrp, max_discount, total_earnings, created_at, updated_at
	FROM products`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(
			&p.ID, &p.ProductID, &p.ProductTitle, &p.Description, &p.ProductURL,
			&p.CategoryID, &p.MRP, &p.MaxDiscount, &p.TotalEarnings, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}
