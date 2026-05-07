package internal

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// Chunk is a searchable section of a Gradle documentation page.
type Chunk struct {
	ID      int
	URL     string
	Title   string
	Heading string
	Body    string
}

// SearchResult represents a single lexical search hit.
type SearchResult struct {
	Title   string
	Heading string
	URL     string
	Snippet string
	Rank    float64
}

// BuildIndex creates a fresh SQLite FTS5 database from documentation chunks.
func BuildIndex(dbPath string, chunks []Chunk) error {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return fmt.Errorf("create db dir: %w", err)
	}
	if err := os.Remove(dbPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove old db: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	stmts := []string{
		`CREATE VIRTUAL TABLE gradle_docs_fts USING fts5(title, heading, url UNINDEXED, body, tokenize='porter unicode61')`,
		`CREATE TABLE meta (key TEXT PRIMARY KEY, value TEXT)`,
	}
	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("exec %q: %w", stmt, err)
		}
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	ins, err := tx.Prepare(`INSERT INTO gradle_docs_fts (title, heading, url, body) VALUES (?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("prepare insert: %w", err)
	}
	defer ins.Close()

	for _, c := range chunks {
		if _, err := ins.Exec(c.Title, c.Heading, c.URL, c.Body); err != nil {
			return fmt.Errorf("insert chunk %d (%s): %w", c.ID, c.URL, err)
		}
	}
	return tx.Commit()
}

// SetMeta writes a key-value pair to the meta table.
func SetMeta(dbPath, key, value string) error {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Exec(`INSERT OR REPLACE INTO meta (key, value) VALUES (?, ?)`, key, value)
	return err
}

// GetMeta reads a value from the meta table.
func GetMeta(db *sql.DB, key string) (string, error) {
	var val string
	err := db.QueryRow(`SELECT value FROM meta WHERE key = ?`, key).Scan(&val)
	if err != nil {
		return "", err
	}
	return val, nil
}

// Search queries the FTS5 index and returns matching sections.
func Search(db *sql.DB, query string, limit int) ([]SearchResult, error) {
	if limit <= 0 {
		limit = 10
	}

	ftsQuery := BuildFTSQuery(query)
	if ftsQuery == "" {
		return nil, fmt.Errorf("empty query")
	}

	rows, err := db.Query(`
		SELECT title, heading, url, snippet(gradle_docs_fts, 3, '>>>', '<<<', '...', 72), rank
		FROM gradle_docs_fts
		WHERE gradle_docs_fts MATCH ?
		ORDER BY rank
		LIMIT ?
	`, ftsQuery, limit)
	if err != nil {
		return nil, fmt.Errorf("search query: %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var r SearchResult
		if err := rows.Scan(&r.Title, &r.Heading, &r.URL, &r.Snippet, &r.Rank); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

// BuildFTSQuery turns a user query into a prefix OR query for SQLite FTS5.
func BuildFTSQuery(query string) string {
	terms := strings.FieldsFunc(query, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})

	seen := map[string]bool{}
	var ftsTerms []string
	for _, term := range terms {
		term = strings.ToLower(strings.TrimSpace(term))
		if len(term) < 2 || seen[term] {
			continue
		}
		seen[term] = true
		term = strings.ReplaceAll(term, `"`, `""`)
		ftsTerms = append(ftsTerms, fmt.Sprintf(`"%s"*`, term))
	}
	return strings.Join(ftsTerms, " OR ")
}

// GetByURLOrHeading returns the complete body for a page section.
func GetByURLOrHeading(db *sql.DB, key string) (SearchResult, error) {
	var r SearchResult
	err := db.QueryRow(`
		SELECT title, heading, url, body, 0.0
		FROM gradle_docs_fts
		WHERE url = ? OR heading = ?
		ORDER BY CASE WHEN url = ? THEN 0 ELSE 1 END
		LIMIT 1
	`, key, key, key).Scan(&r.Title, &r.Heading, &r.URL, &r.Snippet, &r.Rank)
	if err != nil {
		return SearchResult{}, err
	}
	return r, nil
}
