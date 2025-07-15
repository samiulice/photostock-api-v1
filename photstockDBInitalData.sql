

SELECT * from users;
SELECT * from media_categories;
DELETE FROM medias;

ALTER TABLE subscription_plans 
    ADD column expires_at INTEGER DEFAULT 0;

SELECT * FROM subscription_plans;