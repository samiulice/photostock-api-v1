-- CLEANUP SECTION
DROP SCHEMA public CASCADE;
CREATE SCHEMA public;
GRANT ALL ON SCHEMA public TO photostock_db_kms3_user;
GRANT ALL ON SCHEMA public TO public;

-- Create independent tables first (no foreign keys)
CREATE TABLE subscription_plans (
    id SERIAL PRIMARY KEY,
    title VARCHAR(100) NOT NULL DEFAULT '',
    terms TEXT DEFAULT '',
    status BOOLEAN DEFAULT TRUE,
    price NUMERIC(20,2) DEFAULT 0,
    download_limit INTEGER DEFAULT 0,
    expires_at INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE media_categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL DEFAULT '',
    thumbnail_uuid TEXT DEFAULT '',
    total_uploads INTEGER DEFAULT 0,
    total_downloads INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create users without subscription_id FK
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL DEFAULT '',
    password TEXT NOT NULL DEFAULT '',
    name VARCHAR(100) DEFAULT '',
    avatar_url TEXT DEFAULT '',
    status BOOLEAN DEFAULT TRUE,
    role VARCHAR(50) DEFAULT 'user',
    email VARCHAR(100) UNIQUE NOT NULL DEFAULT '',
    mobile VARCHAR(100) DEFAULT '',
    total_earnings NUMERIC(20,2) DEFAULT 0,
    total_withdraw NUMERIC(20,2) DEFAULT 0,
    total_expenses NUMERIC(20,2) DEFAULT 0,
    address TEXT DEFAULT '',
    subscription_id INTEGER DEFAULT NULL, -- Will add FK later
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create subscriptions (depends on users and subscription_plans)
CREATE TABLE subscriptions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    subscription_plans_id INTEGER NOT NULL,
    payment_amount NUMERIC(10, 2) DEFAULT 0,
    payment_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    total_downloads INTEGER DEFAULT 0,
    status BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_subscription_user FOREIGN KEY (user_id)
        REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT fk_subscription_plans FOREIGN KEY (subscription_plans_id)
        REFERENCES subscription_plans (id) ON DELETE CASCADE
);

-- Add FK to users now that subscriptions exists
ALTER TABLE users ADD CONSTRAINT fk_user_subscription 
    FOREIGN KEY (subscription_id) REFERENCES subscriptions (id) ON DELETE SET NULL;

-- Create medias (depends on users and media_categories)
CREATE TABLE medias (
    id SERIAL PRIMARY KEY,
    media_uuid VARCHAR(255) UNIQUE NOT NULL,
    media_title VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '',
    category_id INTEGER,
    license_type INTEGER NOT NULL DEFAULT 0,
    uploader_id INTEGER,
    uploader_name VARCHAR(255) NOT NULL DEFAULT '',
    total_downloads INTEGER DEFAULT 0,
    total_earnings NUMERIC(20,2) DEFAULT 0,
    file_type VARCHAR(50) NOT NULL DEFAULT '',
    file_ext VARCHAR(50) NOT NULL DEFAULT '',
    file_name VARCHAR(255) NOT NULL DEFAULT '',
    file_size VARCHAR(50) NOT NULL DEFAULT '',
    resolution VARCHAR(50) DEFAULT '',  -- e.g. "1920x1080px"
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_media_category FOREIGN KEY (category_id)
        REFERENCES media_categories (id) ON DELETE SET NULL,
    CONSTRAINT fk_uploader_user FOREIGN KEY (uploader_id)
        REFERENCES users (id) ON DELETE SET NULL
);

-- Create history tables (depend on users and medias)
CREATE TABLE download_history (
    id SERIAL PRIMARY KEY,
    media_uuid VARCHAR(255) NOT NULL DEFAULT '',
    user_id INTEGER NOT NULL,
    price NUMERIC(10, 2) DEFAULT 0,
    file_type VARCHAR(50) NOT NULL DEFAULT '',
    file_ext VARCHAR(50) NOT NULL DEFAULT '',
    file_name VARCHAR(255) NOT NULL DEFAULT '',
    file_size VARCHAR(50) NOT NULL DEFAULT '',
    resolution VARCHAR(50) DEFAULT '',  -- e.g. "1920x1080px"
    downloaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_download_media FOREIGN KEY (media_uuid)
        REFERENCES medias (media_uuid) ON DELETE CASCADE,
    CONSTRAINT fk_download_user FOREIGN KEY (user_id)
        REFERENCES users (id) ON DELETE CASCADE
);

CREATE TABLE upload_history (
    id SERIAL PRIMARY KEY,
    media_uuid VARCHAR(255) NOT NULL DEFAULT '',
    user_id INTEGER NOT NULL,
    file_type VARCHAR(50) NOT NULL DEFAULT '',
    file_ext VARCHAR(50) NOT NULL DEFAULT '',
    file_name VARCHAR(255) NOT NULL DEFAULT '',
    file_size VARCHAR(50) NOT NULL DEFAULT '',
    resolution VARCHAR(50) DEFAULT '',  -- e.g. "1920x1080px"
    uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_upload_media FOREIGN KEY (media_uuid)
        REFERENCES medias (media_uuid) ON DELETE CASCADE,

    CONSTRAINT fk_upload_user FOREIGN KEY (user_id)
        REFERENCES users (id) ON DELETE CASCADE
);


-- Create indexes
CREATE INDEX idx_users_email ON users (email);
CREATE INDEX idx_users_role ON users (role);
CREATE INDEX idx_medias_uuid ON medias (media_uuid);
CREATE INDEX idx_subscription_user_id ON subscriptions (user_id);
CREATE INDEX idx_download_user_id ON download_history (user_id);
CREATE INDEX idx_upload_user_id ON upload_history (user_id);