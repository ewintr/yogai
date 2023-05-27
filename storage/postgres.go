package storage

import (
	"database/sql"
	"fmt"

	"ewintr.nl/yogai/model"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type PostgresInfo struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

type Postgres struct {
	db *sql.DB
}

func NewPostgres(pgInfo PostgresInfo) (*Postgres, error) {
	db, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		pgInfo.Host, pgInfo.Port, pgInfo.User, pgInfo.Password, pgInfo.Database))
	if err != nil {
		return &Postgres{}, err
	}
	p := &Postgres{db: db}
	if err := p.migrate(pgMigration); err != nil {
		return &Postgres{}, err
	}

	return p, nil
}

type PostgresVideoRepository struct {
	*Postgres
}

func NewPostgresVideoRepository(postgres *Postgres) *PostgresVideoRepository {
	return &PostgresVideoRepository{postgres}
}

func (p *PostgresVideoRepository) Save(v *model.Video) error {
	query := `INSERT INTO video (id, status, youtube_id, youtube_channel_id, title, description, summary)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (id)
DO UPDATE SET
  id = EXCLUDED.id,
  status = EXCLUDED.status,
  youtube_id = EXCLUDED.youtube_id,
  youtube_channel_id = EXCLUDED.youtube_channel_id,
  title = EXCLUDED.title,
  description = EXCLUDED.description,
  summary = EXCLUDED.summary;`
	_, err := p.db.Exec(query, v.ID, v.Status, v.YoutubeID, v.YoutubeChannelID, v.Title, v.Description, v.Summary)

	return err
}

func (p *PostgresVideoRepository) FindByStatus(statuses ...model.VideoStatus) ([]*model.Video, error) {
	query := `SELECT id, status, youtube_channel_id, youtube_id, title, description, summary
FROM video
WHERE status = ANY($1)`
	rows, err := p.db.Query(query, pq.Array(statuses))
	if err != nil {
		return nil, err
	}

	videos := []*model.Video{}
	for rows.Next() {
		v := &model.Video{}
		if err := rows.Scan(&v.ID, &v.Status, &v.YoutubeChannelID, &v.YoutubeID, &v.Title, &v.Description, &v.Summary); err != nil {
			return nil, err
		}
		videos = append(videos, v)
	}
	rows.Close()

	return videos, nil
}

type PostgresFeedRepository struct {
	*Postgres
}

func NewPostgresFeedRepository(postgres *Postgres) *PostgresFeedRepository {
	return &PostgresFeedRepository{postgres}
}

func (p *PostgresFeedRepository) Save(f *model.Feed) error {
	query := `INSERT INTO feed (id, status, youtube_channel_id, title)
VALUES ($1, $2, $3, $4)
ON CONFLICT (id)
DO UPDATE SET
  id = EXCLUDED.id,
  status = EXCLUDED.status,
  youtube_channel_id = EXCLUDED.youtube_channel_id,
  title = EXCLUDED.title;`
	_, err := p.db.Exec(query, f.ID, f.Status, f.YoutubeChannelID, f.Title)

	return err
}

func (p *PostgresFeedRepository) FindByStatus(statuses ...model.FeedStatus) ([]*model.Feed, error) {
	query := `SELECT id, status, youtube_channel_id, title
FROM feed
WHERE status = ANY($1)`
	rows, err := p.db.Query(query, pq.Array(statuses))
	if err != nil {
		return nil, err
	}

	feeds := []*model.Feed{}
	for rows.Next() {
		f := &model.Feed{}
		if err := rows.Scan(&f.ID, &f.Status, &f.YoutubeChannelID, &f.Title); err != nil {
			return nil, err
		}
		feeds = append(feeds, f)
	}
	rows.Close()

	return feeds, nil
}

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
}

func (p *Postgres) migrate(wanted []string) error {
	query := `CREATE TABLE IF NOT EXISTS migration
("id" SERIAL PRIMARY KEY, "query" TEXT)`
	_, err := p.db.Exec(query)
	if err != nil {
		return err
	}

	// find existing
	rows, err := p.db.Query(`SELECT query FROM migration ORDER BY id`)
	if err != nil {
		return err
	}

	existing := []string{}
	for rows.Next() {
		var query string
		if err := rows.Scan(&query); err != nil {
			return err
		}
		existing = append(existing, query)
	}
	rows.Close()

	// compare
	missing, err := compareMigrations(wanted, existing)
	if err != nil {
		return err
	}

	// execute missing
	for _, query := range missing {
		if _, err := p.db.Exec(query); err != nil {
			return err
		}

		// register
		if _, err := p.db.Exec(`
INSERT INTO migration
(query) VALUES ($1)
`, query); err != nil {
			return err
		}
	}

	return nil
}

func compareMigrations(wanted, existing []string) ([]string, error) {
	needed := []string{}
	if len(wanted) < len(existing) {
		return []string{}, fmt.Errorf("not enough migrations")
	}

	for i, want := range wanted {
		switch {
		case i >= len(existing):
			needed = append(needed, want)
		case want == existing[i]:
			// do nothing
		case want != existing[i]:
			return []string{}, fmt.Errorf("incompatible migration: %v", want)
		}
	}

	return needed, nil
}
