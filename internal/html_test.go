package internal

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestExtractDocumentUsesContentAndSkipsNavigation(t *testing.T) {
	const page = `<!doctype html>
<html>
<head><title>Gradle Test Version 9.5.0</title></head>
<body>
<nav><a href="noise.html">Navigation Noise</a></nav>
<main>
  <div id="content">
    <h1 id="intro">Configuration Cache</h1>
    <p>The configuration cache improves build performance.</p>
    <h2 id="requirements">Requirements</h2>
    <p>Tasks must not touch Project at execution time.</p>
  </div>
</main>
</body>
</html>`
	doc, err := ExtractDocument("https://docs.gradle.org/current/userguide/configuration_cache.html", strings.NewReader(page))
	if err != nil {
		t.Fatal(err)
	}
	if doc.Version != "9.5.0" {
		t.Fatalf("version = %q", doc.Version)
	}
	if len(doc.Chunks) != 2 {
		t.Fatalf("chunks = %d, want 2: %#v", len(doc.Chunks), doc.Chunks)
	}
	if strings.Contains(doc.Chunks[0].Body, "Navigation Noise") {
		t.Fatalf("navigation text leaked into chunk: %q", doc.Chunks[0].Body)
	}
	if doc.Chunks[1].URL != "https://docs.gradle.org/current/userguide/configuration_cache.html#requirements" {
		t.Fatalf("chunk URL = %q", doc.Chunks[1].URL)
	}
}

func TestExtractLinksNormalizesAndIncludesRefresh(t *testing.T) {
	const page = `<html><head><meta http-equiv="refresh" content="0; url=./userguide/userguide.html"></head>
<body><a href="dsl/index.html#top">DSL</a><a href="https://gradle.org">external</a></body></html>`
	root, err := html.Parse(strings.NewReader(page))
	if err != nil {
		t.Fatal(err)
	}
	links := ExtractLinks(root, "https://docs.gradle.org/current/")
	want := []string{
		"https://docs.gradle.org/current/dsl/index.html",
		"https://docs.gradle.org/current/userguide/userguide.html",
		"https://gradle.org/",
	}
	if strings.Join(links, "\n") != strings.Join(want, "\n") {
		t.Fatalf("links = %#v, want %#v", links, want)
	}
}
