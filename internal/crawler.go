package internal

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// CrawlOptions controls how the Gradle documentation crawler behaves.
type CrawlOptions struct {
	StartURL    string
	ScopePath   string
	MaxPages    int
	Workers     int
	Timeout     time.Duration
	UserAgent   string
	IndexFilter func(string) bool
}

// CrawlStats records what actually happened during a crawl.
type CrawlStats struct {
	PagesScheduled int
	PagesFetched   int
	PagesIndexed   int
	ChunksIndexed  int
	Skipped        int
	Version        string
	Warnings       []string
}

// Crawl downloads official Gradle documentation pages and extracts searchable chunks.
func Crawl(ctx context.Context, opts CrawlOptions) ([]Chunk, CrawlStats, error) {
	if opts.StartURL == "" {
		opts.StartURL = "https://docs.gradle.org/current/userguide/userguide.html"
	}
	if opts.ScopePath == "" {
		opts.ScopePath = "/current/"
	}
	if opts.Workers <= 0 {
		opts.Workers = 8
	}
	if opts.Timeout <= 0 {
		opts.Timeout = 20 * time.Second
	}
	if opts.UserAgent == "" {
		opts.UserAgent = "agents-gradle-docs-crawler/0.1 (+https://docs.gradle.org/current/)"
	}
	if opts.IndexFilter == nil {
		opts.IndexFilter = ShouldIndexURL
	}

	start, err := url.Parse(opts.StartURL)
	if err != nil {
		return nil, CrawlStats{}, fmt.Errorf("parse start url: %w", err)
	}

	client := &http.Client{Timeout: opts.Timeout}
	queue := make(chan string, 10000)
	var jobs sync.WaitGroup
	var workers sync.WaitGroup

	var mu sync.Mutex
	visited := map[string]bool{}
	var chunks []Chunk
	stats := CrawlStats{}

	addWarning := func(format string, args ...any) {
		mu.Lock()
		defer mu.Unlock()
		if len(stats.Warnings) < 25 {
			stats.Warnings = append(stats.Warnings, fmt.Sprintf(format, args...))
		}
	}

	enqueue := func(raw string) {
		normalized, ok := NormalizeDocURL(opts.StartURL, raw)
		if !ok {
			return
		}
		if !ShouldVisitURL(normalized, start.Host, opts.ScopePath) {
			return
		}

		mu.Lock()
		if visited[normalized] || (opts.MaxPages > 0 && stats.PagesScheduled >= opts.MaxPages) {
			mu.Unlock()
			return
		}
		visited[normalized] = true
		stats.PagesScheduled++
		jobs.Add(1)
		mu.Unlock()

		select {
		case queue <- normalized:
		case <-ctx.Done():
			jobs.Done()
		}
	}

	process := func(pageURL string) {
		defer jobs.Done()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
		if err != nil {
			addWarning("request %s: %v", pageURL, err)
			return
		}
		req.Header.Set("User-Agent", opts.UserAgent)
		req.Header.Set("Accept", "text/html,application/xhtml+xml")

		resp, err := client.Do(req)
		if err != nil {
			addWarning("fetch %s: %v", pageURL, err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			addWarning("fetch %s: HTTP %d", pageURL, resp.StatusCode)
			return
		}
		contentType := strings.ToLower(resp.Header.Get("Content-Type"))
		if contentType != "" && !strings.Contains(contentType, "text/html") && !strings.Contains(contentType, "application/xhtml") {
			mu.Lock()
			stats.Skipped++
			mu.Unlock()
			return
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			addWarning("read %s: %v", pageURL, err)
			return
		}
		finalURL := resp.Request.URL.String()
		doc, err := ExtractDocument(finalURL, bytes.NewReader(data))
		if err != nil {
			addWarning("extract %s: %v", finalURL, err)
			return
		}

		for _, link := range doc.Links {
			enqueue(link)
		}

		if opts.IndexFilter(finalURL) && len(doc.Chunks) > 0 {
			mu.Lock()
			baseID := len(chunks)
			for i := range doc.Chunks {
				doc.Chunks[i].ID = baseID + i
			}
			chunks = append(chunks, doc.Chunks...)
			stats.PagesFetched++
			stats.PagesIndexed++
			stats.ChunksIndexed += len(doc.Chunks)
			if stats.Version == "" && doc.Version != "" {
				stats.Version = doc.Version
			}
			mu.Unlock()
		} else {
			mu.Lock()
			stats.PagesFetched++
			stats.Skipped++
			mu.Unlock()
		}
	}

	for i := 0; i < opts.Workers; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for pageURL := range queue {
				process(pageURL)
			}
		}()
	}

	enqueue(opts.StartURL)
	go func() {
		jobs.Wait()
		close(queue)
	}()
	workers.Wait()

	if ctx.Err() != nil {
		return chunks, stats, ctx.Err()
	}
	return chunks, stats, nil
}

// ShouldVisitURL returns whether the crawler may fetch this URL.
func ShouldVisitURL(rawURL, host, scopePath string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	if !strings.EqualFold(u.Host, host) {
		return false
	}
	if !strings.HasPrefix(u.Path, scopePath) {
		return false
	}
	ext := strings.ToLower(filepath.Ext(u.Path))
	switch ext {
	case ".html", ".htm", "":
		return true
	default:
		return false
	}
}

// ShouldIndexURL skips generated navigation/search pages while still allowing them as crawl sources.
func ShouldIndexURL(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	base := strings.ToLower(filepath.Base(u.Path))
	switch base {
	case "", ".", "search.html", "index-all.html", "allclasses-index.html", "allpackages-index.html",
		"overview-tree.html", "deprecated-list.html", "help-doc.html", "constant-values.html",
		"serialized-form.html":
		return false
	default:
		return true
	}
}
