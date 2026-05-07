package internal

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func TestSearchFindsIndexedGradleDocs(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "gradle.db")
	chunks := []Chunk{
		{
			URL:     "https://docs.gradle.org/current/userguide/configuration_cache.html#requirements",
			Title:   "Configuration Cache",
			Heading: "Requirements",
			Body:    "The configuration cache improves build performance by caching the result of the configuration phase.",
		},
	}
	if err := BuildIndex(dbPath, chunks); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	results, err := Search(db, "configuration cache", 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("results = %d, want 1", len(results))
	}
	if results[0].URL != chunks[0].URL {
		t.Fatalf("result URL = %q", results[0].URL)
	}
}
