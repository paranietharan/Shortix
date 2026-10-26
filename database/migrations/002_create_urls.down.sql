-- 002_create_urls.down.sql

DROP INDEX IF EXISTS ux_urls_custom_alias_not_null;
DROP INDEX IF EXISTS ux_urls_short_code;

DROP TABLE IF EXISTS urls;
