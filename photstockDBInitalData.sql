

SELECT * from users;
SELECT * from media_categories;


SELECT * FROM subscription_plans;
select * from upload_history;


ALTER TABLE upload_history
ADD COLUMN file_type VARCHAR(50),
ADD COLUMN file_name VARCHAR(255),
ADD COLUMN file_size VARCHAR(50);
