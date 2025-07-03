-- CLEANUP SECTION
DROP TABLE IF EXISTS download_history;
DROP TABLE IF EXISTS upload_history;
DROP TABLE IF EXISTS medias;
DROP TABLE IF EXISTS subscriptions;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS media_categories;
DROP TABLE IF EXISTS subscription_plans;

-- TABLE: subscription_plans
CREATE TABLE subscription_plans (
    id SERIAL PRIMARY KEY,
    title VARCHAR(100) NOT NULL DEFAULT '',
    terms TEXT DEFAULT '',
    status BOOLEAN DEFAULT TRUE,
    download_limit INTEGER DEFAULT 0,
    time_limit INTERVAL DEFAULT INTERVAL '0',
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- TABLE: media_categories
CREATE TABLE media_categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL DEFAULT '',
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- TABLE: users
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL DEFAULT '',
    password TEXT NOT NULL DEFAULT '',
    name VARCHAR(100) DEFAULT '',
    avatar_url TEXT DEFAULT '',
    status BOOLEAN DEFAULT TRUE,
    role VARCHAR(50) DEFAULT 'user',
    email VARCHAR(100) DEFAULT '',
    mobile VARCHAR(100) DEFAULT '',
    total_earnings NUMERIC(20,2) default 0,
    total_withdraw NUMERIC(20,2) default 0,
    address TEXT DEFAULT '',
    subscription_id INTEGER DEFAULT NULL,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_user_subscription FOREIGN KEY (subscription_id)
        REFERENCES subscription_plans (id) ON DELETE SET NULL
);

-- TABLE: subscriptions
CREATE TABLE subscriptions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    subscription_plans_id INTEGER NOT NULL,
    payment_status VARCHAR(20) DEFAULT '',
    payment_amount NUMERIC(10, 2) DEFAULT 0,
    payment_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_subscription_user FOREIGN KEY (user_id)
        REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT fk_subscription_plans FOREIGN KEY (subscription_plans_id)
        REFERENCES subscription_plans (id) ON DELETE CASCADE
);

-- TABLE: medias
CREATE TABLE medias (
    id SERIAL PRIMARY KEY,
    media_uuid VARCHAR(255) UNIQUE NOT NULL DEFAULT '',
    media_title VARCHAR(255) NOT NULL DEFAULT '',
    description TEXT DEFAULT '',
    media_url TEXT NOT NULL DEFAULT '',
    category_id INTEGER DEFAULT NULL,
    license_type INTEGER DEFAULT 0,
    total_earnings NUMERIC(20,2) default 0,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_media_category FOREIGN KEY (category_id)
        REFERENCES media_categories (id) ON DELETE SET NULL
);

-- TABLE: download_history
CREATE TABLE download_history (
    id SERIAL PRIMARY KEY,
    media_uuid VARCHAR(255) NOT NULL DEFAULT '',
    user_id INTEGER NOT NULL,
    price NUMERIC(10, 2) DEFAULT 0,
    downloaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_download_media FOREIGN KEY (media_uuid)
        REFERENCES medias (media_uuid) ON DELETE CASCADE,
    CONSTRAINT fk_download_user FOREIGN KEY (user_id)
        REFERENCES users (id) ON DELETE CASCADE
);
-- TABLE: upload_history
CREATE TABLE upload_history (
    id SERIAL PRIMARY KEY,
    media_uuid VARCHAR(255) NOT NULL DEFAULT '',
    user_id INTEGER NOT NULL,
    uploadeded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_uploadd_media FOREIGN KEY (media_uuid)
        REFERENCES medias (media_uuid) ON DELETE CASCADE,
    CONSTRAINT fk_upload_user FOREIGN KEY (user_id)
        REFERENCES users (id) ON DELETE CASCADE
);

-- INDEXES
CREATE INDEX idx_users_email ON users (email);
CREATE INDEX idx_users_role ON users (role);
CREATE INDEX idx_medias_uuid ON medias (media_uuid);
CREATE INDEX idx_subscription_user_id ON subscriptions (user_id);
CREATE INDEX idx_download_user_id ON download_history (user_id);
CREATE INDEX idx_dupload_user_id ON upload_history (user_id);
