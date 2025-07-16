

SELECT * from users;
SELECT * from medias;
SELECT * FROM subscription_plans;

select * from upload_history;

SELECT * FROM download_history;
alter table media_categories
add COLUMN total_uploads INTEGER DEFAULT 0,
add COLUMN total_downloads INTEGER DEFAULT 0;
-- Select all rows from 'media_categories'
SELECT * FROM media_categories;
INSERT INTO download_history (
    media_uuid, user_id, price, file_type, file_ext, file_name, file_size, resolution, downloaded_at
) VALUES 
('e0cd79c8-23a2-45ab-8628-add3ab869137_1752663566083999500.jpg', 1, 10.00, 'image/jpeg', 'jpg', 'sunset.jpg', '1.2MB', '1920x1080px', NOW()),

('e0cd79c8-23a2-45ab-8628-add3ab869137_1752663566083999500.jpg', 1, 0.00, 'image/png', 'png', 'logo.png', '350KB', '800x600px', NOW()),

('e0cd79c8-23a2-45ab-8628-add3ab869137_1752663566083999500.jpg', 2, 5.50, 'image/webp', 'webp', 'compressed.webp', '700KB', '1280x720px', NOW()),

('e0cd79c8-23a2-45ab-8628-add3ab869137_1752663566083999500.jpg', 2, 2.00, 'video/mp4', 'mp4', 'promo.mp4', '20MB', '1920x1080px', NOW()),

('e0cd79c8-23a2-45ab-8628-add3ab869137_1752663566083999500.jpg', 2, 0.00, 'image/gif', 'gif', 'animation.gif', '900KB', '480x360px', NOW());





