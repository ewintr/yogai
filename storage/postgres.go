package storage

import (
	"database/sql"
	"ewintr.nl/yogai/model"
	"fmt"
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
	query := `INSERT INTO video (id, status, youtube_id, feed_id, title, description, summary)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (id)
DO UPDATE SET
  id = EXCLUDED.id,
  status = EXCLUDED.status,
  youtube_id = EXCLUDED.youtube_id,
  feed_id = EXCLUDED.feed_id,
  title = EXCLUDED.title,
  description = EXCLUDED.description,
  summary = EXCLUDED.summary;`
	_, err := p.db.Exec(query, v.ID, v.Status, v.YoutubeID, v.FeedID, v.Title, v.Description, v.Summary)

	return err
}

func (p *PostgresVideoRepository) FindByStatus(statuses ...model.Status) ([]*model.Video, error) {
	query := `SELECT id, status, youtube_id, feed_id, title, description, summary
FROM video
WHERE status = ANY($1)`
	rows, err := p.db.Query(query, pq.Array(statuses))
	if err != nil {
		return nil, err
	}

	videos := []*model.Video{}
	for rows.Next() {
		v := &model.Video{}
		if err := rows.Scan(&v.ID, &v.Status, &v.YoutubeID, &v.FeedID, &v.Title, &v.Description, &v.Summary); err != nil {
			return nil, err
		}
		videos = append(videos, v)
	}
	rows.Close()

	return videos, nil
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
