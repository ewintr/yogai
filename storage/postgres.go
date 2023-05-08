package storage

import (
	"database/sql"
	"ewintr.nl/yogai/model"
	"fmt"
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
	query := `INSERT INTO video (id, status, youtube_url, feed_id, title, description)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (id)
DO UPDATE SET
  id = EXCLUDED.id,
  status = EXCLUDED.status,
  youtube_url = EXCLUDED.youtube_url,
  feed_id = EXCLUDED.feed_id,
  title = EXCLUDED.title,
  description = EXCLUDED.description;`
	_, err := p.db.Exec(query, v.ID, v.Status, v.YoutubeURL, v.FeedID, v.Title, v.Description)

	return err
}

var pgMigration = []string{
	`CREATE TYPE video_status AS ENUM ('new', 'needs_summary', 'ready')`,
	`CREATE TABLE video (
    id uuid PRIMARY KEY,
    status video_status NOT NULL,
    youtube_url VARCHAR(255) NOT NULL,
    title VARCHAR(255) NOT NULL,
    feed_id VARCHAR(255) NOT NULL,
    description TEXT,
    summary TEXT
)`,
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
