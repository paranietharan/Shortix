-- 003_create_clicks.down.sql

DROP INDEX IF EXISTS ix_clicks_clicked_at;
DROP INDEX IF EXISTS ix_clicks_url_id;

DROP TABLE IF EXISTS clicks;
