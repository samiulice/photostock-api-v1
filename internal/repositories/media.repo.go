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
			license_type, uploader_id, uploader_name,
			total_downloads, total_earnings,
			file_type, file_ext, file_name, file_size, resolution,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4,
			$5, $6, $7,
			$8, $9,
			$10, $11, $12, $13, $14,
			$15, $16
		)
		RETURNING id`
	now := time.Now()
	err := r.db.QueryRow(ctx, query,
		m.MediaUUID, m.MediaTitle, m.Description, m.CategoryID,
		m.LicenseType, m.UploaderID, m.UploaderName,
		m.TotalDownloads, m.TotalEarnings,
		m.FileType, m.FileExt, m.FileName, m.FileSize, m.Resolution,
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
			m.id, m.media_uuid, m.media_title, m.description, m.category_id,
			m.license_type, m.uploader_id, m.uploader_name, m.total_downloads,
			m.total_earnings, m.file_type, m.file_ext, m.file_name, m.file_size,
			m.resolution, m.created_at, m.updated_at,
			c.id, c.name, c.created_at, c.updated_at
		FROM medias m
		LEFT JOIN media_categories c ON m.category_id = c.id
		WHERE m.id = $1`
	var m models.Media
	var c models.MediaCategory
	err := r.db.QueryRow(ctx, query, id).Scan(
		&m.ID, &m.MediaUUID, &m.MediaTitle, &m.Description, &m.CategoryID,
		&m.LicenseType, &m.UploaderID, &m.UploaderName, &m.TotalDownloads,
		&m.TotalEarnings, &m.FileType, &m.FileExt, &m.FileName, &m.FileSize,
		&m.Resolution, &m.CreatedAt, &m.UpdatedAt,
		&c.ID, &c.Name, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	m.MediaCategory = c
	return &m, nil
}

// GetByMediaUUID retrieves media by media_uuid.
func (r *MediaRepo) GetByMediaUUID(ctx context.Context, mediaUUID string) (*models.Media, error) {
	query := `
		SELECT 
			m.id, m.media_uuid, m.media_title, m.description, m.category_id,
			m.license_type, m.uploader_id, m.uploader_name, m.total_downloads,
			m.total_earnings, m.file_type, m.file_ext, m.file_name, m.file_size,
			m.resolution, m.created_at, m.updated_at,
			c.id, c.name, c.created_at, c.updated_at
		FROM medias m
		LEFT JOIN media_categories c ON m.category_id = c.id
		WHERE m.media_uuid = $1`
	var m models.Media
	var c models.MediaCategory
	err := r.db.QueryRow(ctx, query, mediaUUID).Scan(
		&m.ID, &m.MediaUUID, &m.MediaTitle, &m.Description, &m.CategoryID,
		&m.LicenseType, &m.UploaderID, &m.UploaderName, &m.TotalDownloads,
		&m.TotalEarnings, &m.FileType, &m.FileExt, &m.FileName, &m.FileSize,
		&m.Resolution, &m.CreatedAt, &m.UpdatedAt,
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
		SET media_uuid = $2,
			media_title = $3,
			description = $4,
			category_id = $5,
			license_type = $6,
			uploader_id = $7,
			uploader_name = $8,
			total_downloads = $9,
			total_earnings = $10,
			file_type = $11,
			file_ext = $12,
			file_name = $13,
			file_size = $14,
			resolution = $15,
			updated_at = $16
		WHERE id = $1`
	_, err := r.db.Exec(ctx, query,
		m.ID, m.MediaUUID, m.MediaTitle, m.Description, m.CategoryID,
		m.LicenseType, m.UploaderID, m.UploaderName, m.TotalDownloads, m.TotalEarnings,
		m.FileType, m.FileExt, m.FileName, m.FileSize, m.Resolution,
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
			m.id, m.media_uuid, m.media_title, m.description, m.category_id,
			m.license_type, m.uploader_id, m.uploader_name, m.total_downloads,
			m.total_earnings, m.file_type, m.file_ext, m.file_name, m.file_size,
			m.resolution, m.created_at, m.updated_at,
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
			&m.ID, &m.MediaUUID, &m.MediaTitle, &m.Description, &m.CategoryID,
			&m.LicenseType, &m.UploaderID, &m.UploaderName, &m.TotalDownloads,
			&m.TotalEarnings, &m.FileType, &m.FileExt, &m.FileName, &m.FileSize,
			&m.Resolution, &m.CreatedAt, &m.UpdatedAt,
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


// GetAllByCategoryID returns all media for a specific category.
func (r *MediaRepo) GetAllByCategoryID(ctx context.Context, id int) ([]*models.Media, error) {
	query := `
		SELECT 
			m.id, m.media_uuid, m.media_title, m.description, m.category_id,
			m.license_type, m.uploader_id, m.uploader_name, m.total_downloads,
			m.total_earnings, m.file_type, m.file_ext, m.file_name, m.file_size,
			m.resolution, m.created_at, m.updated_at,
			c.id, c.name, c.created_at, c.updated_at
		FROM medias m
		LEFT JOIN media_categories c ON m.category_id = c.id
		WHERE m.category_id = $1`
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
			&m.ID, &m.MediaUUID, &m.MediaTitle, &m.Description, &m.CategoryID,
			&m.LicenseType, &m.UploaderID, &m.UploaderName, &m.TotalDownloads,
			&m.TotalEarnings, &m.FileType, &m.FileExt, &m.FileName, &m.FileSize,
			&m.Resolution, &m.CreatedAt, &m.UpdatedAt,
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

// IncrementDownloadCount increases total_downloads by 1 for a given media ID.
func (r *MediaRepo) IncrementDownloadCountByID(ctx context.Context, id int) error {
	query := `
		UPDATE medias
		SET total_downloads = total_downloads + 1,
			updated_at = $2
		WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id, time.Now())
	return err
}
