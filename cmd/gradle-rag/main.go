package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/m0n0x41d/agents-gradle/internal"
	_ "modernc.org/sqlite"
)

var version = "dev"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "search":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: gradle-rag search <query> [--limit N] [--full]")
			os.Exit(1)
		}
		cmdSearch(os.Args[2:])
	case "page":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: gradle-rag page <url-or-heading>")
			os.Exit(1)
		}
		cmdPage(strings.Join(os.Args[2:], " "))
	case "info":
		cmdInfo()
	default:
		cmdSearch(os.Args[1:])
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `gradle-rag — lexical search over current Gradle documentation

Commands:
  search <query> [--limit N] [--full]   Search Gradle docs
  page <url-or-heading>                  Print one indexed section
  info                                   Show index metadata

Examples:
  gradle-rag search "configuration cache problems" --limit 5
  gradle-rag search "TaskProvider register" --full
  gradle-rag page "https://docs.gradle.org/current/userguide/configuration_cache.html#config_cache:requirements"
  gradle-rag info
`)
}

func openDB() (*sql.DB, func(), error) {
	embeddedDB, err := readEmbeddedDB()
	if err != nil {
		return nil, nil, err
	}

	tmpDir, err := os.MkdirTemp("", "gradle-rag-*")
	if err != nil {
		return nil, nil, fmt.Errorf("create temp dir: %w", err)
	}

	dbPath := filepath.Join(tmpDir, "gradle.db")
	if err := os.WriteFile(dbPath, embeddedDB, 0644); err != nil {
		os.RemoveAll(tmpDir)
		return nil, nil, fmt.Errorf("write temp db: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		return nil, nil, fmt.Errorf("open db: %w", err)
	}

	cleanup := func() {
		db.Close()
		os.RemoveAll(tmpDir)
	}
	return db, cleanup, nil
}

func readEmbeddedDB() ([]byte, error) {
	data, err := embeddedDBs.ReadFile("db/gradle.db")
	if err == nil {
		return data, nil
	}
	data, fallbackErr := embeddedDBs.ReadFile("db/placeholder.db")
	if fallbackErr == nil {
		return data, nil
	}
	return nil, fmt.Errorf("read embedded db: %w; fallback: %v", err, fallbackErr)
}

func cmdSearch(args []string) {
	limit := 10
	full := false
	var queryParts []string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--limit":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "Error: --limit requires a value")
				os.Exit(1)
			}
			i++
			n, err := strconv.Atoi(args[i])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: invalid --limit %q\n", args[i])
				os.Exit(1)
			}
			limit = n
		case "--full":
			full = true
		default:
			queryParts = append(queryParts, args[i])
		}
	}

	query := strings.Join(queryParts, " ")
	if strings.TrimSpace(query) == "" {
		fmt.Fprintln(os.Stderr, "Error: empty query")
		os.Exit(1)
	}

	db, cleanup, err := openDB()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer cleanup()

	results, err := internal.Search(db, query, limit)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Search error: %v\n", err)
		os.Exit(1)
	}
	if len(results) == 0 {
		fmt.Println("No results found.")
		return
	}

	for i, r := range results {
		fmt.Printf("### %d. %s — %s\n\n", i+1, r.Title, r.Heading)
		fmt.Printf("Source: %s\n\n", r.URL)
		if full {
			page, err := internal.GetByURLOrHeading(db, r.URL)
			if err == nil {
				fmt.Println(page.Snippet)
			} else {
				fmt.Println(r.Snippet)
			}
		} else {
			fmt.Println(r.Snippet)
		}
		fmt.Println()
	}
}

func cmdPage(key string) {
	db, cleanup, err := openDB()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer cleanup()

	r, err := internal.GetByURLOrHeading(db, key)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Section not found: %s\n", key)
		os.Exit(1)
	}
	fmt.Printf("## %s — %s\n\nSource: %s\n\n%s\n", r.Title, r.Heading, r.URL, r.Snippet)
}

func cmdInfo() {
	db, cleanup, err := openDB()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer cleanup()

	fmt.Printf("gradle-rag version: %s\n", version)
	for _, key := range []string{"gradle_version", "docs_source", "scope_path", "crawled_at", "page_count", "scheduled_count", "chunk_count"} {
		if value, err := internal.GetMeta(db, key); err == nil {
			fmt.Printf("%s: %s\n", key, value)
		}
	}
}
