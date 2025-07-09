package repositories

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samiulice/photostock/internal/models"
)

type MediaRepo struct {
	db *pgxpool.Pool
}

func NewMediaRepo(db *pgxpool.Pool) *MediaRepo {
	return &MediaRepo{db: db}
}

// ------------------------------ Media CRUD ------------------------------

// Create inserts a new media record.
func (r *MediaRepo) Create(ctx context.Context, m *models.Media) error {
	query := `
		INSERT INTO medias (
			media_uuid, media_title, description, category_id,
			total_earnings, license_type, uploader_id, uploader_name,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id`
	now := time.Now()
	err := r.db.QueryRow(ctx, query,
		m.MediaUUID, m.MediaTitle, m.Description, m.CategoryID,
		m.TotalEarnings, m.LicenseType, m.UploaderID, m.UploaderName,
		now, now,
	).Scan(&m.ID)
	m.CreatedAt = now
	m.UpdatedAt = now
	return err
}

// GetByID retrieves media by ID.
func (r *MediaRepo) GetByID(ctx context.Context, id int) (*models.Media, error) {
	query := `
		SELECT 
			m.id, m.media_uuid, m.media_title, m.description,
			m.category_id, m.total_earnings, m.license_type,
			m.uploader_id, m.uploader_name, m.created_at, m.updated_at,
			c.id, c.name, c.created_at, c.updated_at
		FROM medias m
		LEFT JOIN media_categories c ON m.category_id = c.id
		WHERE m.id = $1`
	var m models.Media
	var c models.MediaCategory
	err := r.db.QueryRow(ctx, query, id).Scan(
		&m.ID, &m.MediaUUID, &m.MediaTitle, &m.Description,
		&m.CategoryID, &m.TotalEarnings, &m.LicenseType,
		&m.UploaderID, &m.UploaderName, &m.CreatedAt, &m.UpdatedAt,
		&c.ID, &c.Name, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	m.MediaCategory = c
	return &m, nil
}

// GetByMediaUUID retrieves media by media_uuid.
func (r *MediaRepo) GetByMediaUUID(ctx context.Context, media_uuid string) (*models.Media, error) {
	query := `
		SELECT 
			m.id, m.media_uuid, m.media_title, m.description,
			m.category_id, m.total_earnings, m.license_type,
			m.uploader_id, m.uploader_name, m.created_at, m.updated_at,
			c.id, c.name, c.created_at, c.updated_at
		FROM medias m
		LEFT JOIN media_categories c ON m.category_id = c.id
		WHERE m.media_uuid = $1`
	var m models.Media
	var c models.MediaCategory
	err := r.db.QueryRow(ctx, query, media_uuid).Scan(
		&m.ID, &m.MediaUUID, &m.MediaTitle, &m.Description,
		&m.CategoryID, &m.TotalEarnings, &m.LicenseType,
		&m.UploaderID, &m.UploaderName, &m.CreatedAt, &m.UpdatedAt,
		&c.ID, &c.Name, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	m.MediaCategory = c
	return &m, nil
}

// Update modifies a media record.
func (r *MediaRepo) Update(ctx context.Context, m *models.Media) error {
	query := `
		UPDATE medias
		SET media_title = $2,
			description = $3,
			category_id = $4,
			total_earnings = $5,
			license_type = $6,
			uploader_id = $7,
			uploader_name = $8,
			updated_at = $9
		WHERE id = $1`
	_, err := r.db.Exec(ctx, query,
		m.ID, m.MediaTitle, m.Description, m.CategoryID,
		m.TotalEarnings, m.LicenseType, m.UploaderID, m.UploaderName,
		time.Now(),
	)
	return err
}

// Delete removes a media record.
func (r *MediaRepo) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM medias WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// GetAll returns all media with category info.
func (r *MediaRepo) GetAll(ctx context.Context) ([]*models.Media, error) {
	query := `
		SELECT 
			m.id, m.media_uuid, m.media_title, m.description,
			m.category_id, m.total_earnings, m.license_type,
			m.uploader_id, m.uploader_name, m.created_at, m.updated_at,
			c.id, c.name, c.created_at, c.updated_at
		FROM medias m
		LEFT JOIN media_categories c ON m.category_id = c.id`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var medias []*models.Media
	for rows.Next() {
		var m models.Media
		var c models.MediaCategory
		err := rows.Scan(
			&m.ID, &m.MediaUUID, &m.MediaTitle, &m.Description,
			&m.CategoryID, &m.TotalEarnings, &m.LicenseType,
			&m.UploaderID, &m.UploaderName, &m.CreatedAt, &m.UpdatedAt,
			&c.ID, &c.Name, &c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		m.MediaCategory = c
		medias = append(medias, &m)
	}
	return medias, nil
}

func (r *MediaRepo) GetAllByCategoryID(ctx context.Context, id int) ([]*models.Media, error) {
	query := `
		SELECT 
			m.id, m.media_uuid, m.media_title, m.description,
			m.category_id, m.total_earnings, m.license_type,
			m.uploader_id, m.uploader_name, m.created_at, m.updated_at,
			c.id AS category_id, c.name AS category_name, c.created_at AS category_created_at, c.updated_at AS category_updated_at
		FROM medias m
		LEFT JOIN media_categories c ON m.category_id = c.id
		WHERE m.category_id = $1;
		`
	rows, err := r.db.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var medias []*models.Media
	for rows.Next() {
		var m models.Media
		var c models.MediaCategory
		err := rows.Scan(
			&m.ID, &m.MediaUUID, &m.MediaTitle, &m.Description,
			&m.CategoryID, &m.TotalEarnings, &m.LicenseType,
			&m.UploaderID, &m.UploaderName, &m.CreatedAt, &m.UpdatedAt,
			&c.ID, &c.Name, &c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		m.MediaCategory = c
		medias = append(medias, &m)
	}
	return medias, nil
}
