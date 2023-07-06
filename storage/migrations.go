package storage

var pgMigration = []string{
	`CREATE TYPE video_status AS ENUM ('new', 'ready')`,
	`CREATE TABLE video (
id uuid PRIMARY KEY,
status video_status NOT NULL,
youtube_id VARCHAR(255) NOT NULL UNIQUE,
title VARCHAR(255) NOT NULL,
feed_id VARCHAR(255) NOT NULL,
description TEXT,
summary TEXT
)`,
	`CREATE TYPE video_status_new AS ENUM ('new', 'has_metadata', 'has_summary', 'ready')`,
	`ALTER TABLE video
ALTER COLUMN status TYPE video_status_new
USING video::text::video_status_new`,
	`DROP TYPE video_status`,
	`ALTER TYPE video_status_new RENAME TO video_status`,
	`UPDATE video SET summary = '' WHERE summary IS NULL `,
	`UPDATE video SET description = '' WHERE description IS NULL `,
	`ALTER TABLE video 
ALTER COLUMN summary SET DEFAULT '', 
ALTER COLUMN summary SET NOT NULL,
ALTER COLUMN description SET DEFAULT '', 
ALTER COLUMN description SET NOT NULL`,
	`CREATE TYPE feed_status AS ENUM ('new', 'ready')`,
	`CREATE TABLE feed (
id uuid PRIMARY KEY,
status feed_status NOT NULL,
youtube_channel_id VARCHAR(255) NOT NULL UNIQUE,
title VARCHAR(255) NOT NULL
)`,
	`ALTER TABLE video
DROP COLUMN feed_id,
ADD COLUMN youtube_channel_id VARCHAR(255) NOT NULL REFERENCES feed(youtube_channel_id)`,
	`ALTER TABLE video
ADD COLUMN duration VARCHAR(255),
ADD COLUMN published_at VARCHAR(255)`,
	`ALTER TABLE video RENAME COLUMN duration TO youtube_duration`,
	`ALTER TABLE video RENAME COLUMN published_at TO youtube_published_id`,
	`ALTER TABLE video RENAME COLUMN title TO youtube_title`,
	`ALTER TABLE video RENAME COLUMN description TO youtube_description`,
	`ALTER TABLE video RENAME COLUMN youtube_published_id TO youtube_published_at`,
	`UPDATE video SET status = 'new'`,
	`CREATE TYPE video_status_new AS ENUM ('new', 'fetched', 'ready')`,
	`BEGIN;
ALTER TABLE video ADD COLUMN status_new video_status_new;
UPDATE video SET status_new = status::text::video_status_new;
ALTER TABLE video DROP COLUMN status;
ALTER TABLE video RENAME COLUMN status_new TO status;
COMMIT;`,
	`DROP TYPE video_status`,
	`ALTER TYPE video_status_new RENAME TO video_status`,
}
