package storage

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

type Postgres struct {
	db *sql.DB
}

func NewPostgres(db *sql.DB) (*Postgres, error) {
	p := &Postgres{db: db}
	if err := p.migrate(pgMigration); err != nil {
		return &Postgres{}, err
	}

	return p, nil
}

var pgMigration = []string{
	`CREATE TABLE channel (
    id SERIAL PRIMARY KEY,
    url VARCHAR(255) NOT NULL,
    feed_url VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL
)`,
	`CREATE TABLE video (
    id SERIAL PRIMARY KEY,
    channel_id INTEGER REFERENCES channel(id) ON DELETE CASCADE,
    url VARCHAR(255) NOT NULL,
    title VARCHAR(255) NOT NULL,
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
