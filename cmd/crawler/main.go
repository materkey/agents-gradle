package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/m0n0x41d/agents-gradle/internal"
	_ "modernc.org/sqlite"
)

func main() {
	startURL := flag.String("start", "https://docs.gradle.org/current/userguide/userguide.html", "Gradle documentation URL to start crawling")
	scopePath := flag.String("scope", "/current/", "URL path prefix that bounds the crawl")
	dbPath := flag.String("db", "cmd/gradle-rag/gradle.db", "SQLite database output path")
	versionFile := flag.String("version-file", "GRADLE_DOCS_VERSION", "metadata file to write after a successful crawl")
	maxPages := flag.Int("max-pages", 0, "maximum pages to fetch; 0 means no cap")
	workers := flag.Int("workers", 8, "number of concurrent fetch workers")
	timeout := flag.Duration("timeout", 20*time.Second, "per-request timeout")
	flag.Parse()

	startedAt := time.Now().UTC()
	chunks, stats, err := internal.Crawl(context.Background(), internal.CrawlOptions{
		StartURL:  *startURL,
		ScopePath: *scopePath,
		MaxPages:  *maxPages,
		Workers:   *workers,
		Timeout:   *timeout,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Crawl error: %v\n", err)
		os.Exit(1)
	}
	if len(chunks) == 0 {
		fmt.Fprintln(os.Stderr, "Crawl produced no indexable chunks")
		os.Exit(1)
	}

	if err := internal.BuildIndex(*dbPath, chunks); err != nil {
		fmt.Fprintf(os.Stderr, "Index error: %v\n", err)
		os.Exit(1)
	}

	meta := map[string]string{
		"docs_source":     *startURL,
		"scope_path":      *scopePath,
		"crawled_at":      startedAt.Format(time.RFC3339),
		"page_count":      strconv.Itoa(stats.PagesIndexed),
		"scheduled_count": strconv.Itoa(stats.PagesScheduled),
		"chunk_count":     strconv.Itoa(stats.ChunksIndexed),
		"gradle_version":  stats.Version,
	}
	for key, value := range meta {
		if value == "" {
			continue
		}
		if err := internal.SetMeta(*dbPath, key, value); err != nil {
			fmt.Fprintf(os.Stderr, "Meta error for %s: %v\n", key, err)
			os.Exit(1)
		}
	}

	if *versionFile != "" {
		content := fmt.Sprintf("gradle_version=%s\ncrawled_at=%s\ndocs_source=%s\npages=%d\nchunks=%d\n",
			stats.Version, startedAt.Format(time.RFC3339), *startURL, stats.PagesIndexed, stats.ChunksIndexed)
		if err := os.WriteFile(*versionFile, []byte(content), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Version file error: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Printf("Crawled %d scheduled / %d fetched pages; indexed %d pages and %d chunks into %s\n",
		stats.PagesScheduled, stats.PagesFetched, stats.PagesIndexed, stats.ChunksIndexed, *dbPath)
	if stats.Version != "" {
		fmt.Printf("Gradle docs version: %s\n", stats.Version)
	}
	for _, warning := range stats.Warnings {
		fmt.Fprintf(os.Stderr, "Warning: %s\n", warning)
	}
}
