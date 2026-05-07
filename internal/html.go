package internal

import (
	"fmt"
	"io"
	"net/url"
	"path"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"golang.org/x/net/html"
)

// Document is a parsed Gradle documentation HTML page.
type Document struct {
	URL     string
	Title   string
	Version string
	Chunks  []Chunk
	Links   []string
}

// ExtractDocument extracts source links and searchable chunks from one HTML page.
func ExtractDocument(pageURL string, r io.Reader) (Document, error) {
	root, err := html.Parse(r)
	if err != nil {
		return Document{}, fmt.Errorf("parse html: %w", err)
	}

	title := cleanText(findTitle(root))
	if title == "" {
		title = pageURL
	}

	content := findContentRoot(root)
	if content == nil {
		content = root
	}

	extractor := &sectionExtractor{
		pageURL:        pageURL,
		title:          title,
		currentHeading: title,
		currentURL:     pageURL,
	}
	extractor.walk(content)
	extractor.flush()

	var chunks []Chunk
	for _, chunk := range extractor.chunks {
		if len(strings.TrimSpace(chunk.Body)) > 20 {
			chunk.ID = len(chunks)
			chunks = append(chunks, chunk)
		}
	}

	return Document{
		URL:     pageURL,
		Title:   title,
		Version: findGradleVersion(root, title),
		Chunks:  chunks,
		Links:   ExtractLinks(root, pageURL),
	}, nil
}

// ExtractLinks returns normalized in-page links from an HTML document.
func ExtractLinks(root *html.Node, baseURL string) []string {
	seen := map[string]bool{}
	var links []string

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch strings.ToLower(n.Data) {
			case "a", "link":
				if href, ok := nodeAttr(n, "href"); ok {
					if normalized, ok := NormalizeDocURL(baseURL, href); ok && !seen[normalized] {
						seen[normalized] = true
						links = append(links, normalized)
					}
				}
			case "meta":
				if strings.EqualFold(attrOrEmpty(n, "http-equiv"), "refresh") {
					if href := refreshURL(attrOrEmpty(n, "content")); href != "" {
						if normalized, ok := NormalizeDocURL(baseURL, href); ok && !seen[normalized] {
							seen[normalized] = true
							links = append(links, normalized)
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(root)
	sort.Strings(links)
	return links
}

// NormalizeDocURL resolves and normalizes a candidate documentation URL.
func NormalizeDocURL(baseURL, href string) (string, bool) {
	href = strings.TrimSpace(href)
	if href == "" {
		return "", false
	}
	lower := strings.ToLower(href)
	if strings.HasPrefix(lower, "javascript:") || strings.HasPrefix(lower, "mailto:") || strings.HasPrefix(lower, "tel:") {
		return "", false
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return "", false
	}
	ref, err := url.Parse(href)
	if err != nil {
		return "", false
	}

	u := base.ResolveReference(ref)
	u.Fragment = ""
	u.RawQuery = ""
	u.ForceQuery = false
	u.Scheme = strings.ToLower(u.Scheme)
	u.Host = strings.ToLower(u.Host)
	u.Path = cleanURLPath(u.Path)
	return u.String(), true
}

func cleanURLPath(p string) string {
	if p == "" {
		return "/"
	}
	trailing := strings.HasSuffix(p, "/")
	cleaned := path.Clean(p)
	if trailing && cleaned != "/" {
		cleaned += "/"
	}
	return cleaned
}

func findTitle(root *html.Node) string {
	var walk func(*html.Node) string
	walk = func(n *html.Node) string {
		if n.Type == html.ElementNode && strings.EqualFold(n.Data, "title") {
			return textContent(n)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if s := walk(c); s != "" {
				return s
			}
		}
		return ""
	}
	return walk(root)
}

func findContentRoot(root *html.Node) *html.Node {
	if n := findElement(root, func(n *html.Node) bool { return attrOrEmpty(n, "id") == "content" }); n != nil {
		return n
	}
	if n := findElement(root, func(n *html.Node) bool {
		return strings.EqualFold(n.Data, "main") && strings.EqualFold(attrOrEmpty(n, "role"), "main")
	}); n != nil {
		return n
	}
	if n := findElement(root, func(n *html.Node) bool { return strings.EqualFold(n.Data, "main") }); n != nil {
		return n
	}
	return findElement(root, func(n *html.Node) bool { return strings.EqualFold(n.Data, "body") })
}

func findElement(root *html.Node, pred func(*html.Node) bool) *html.Node {
	if root.Type == html.ElementNode && pred(root) {
		return root
	}
	for c := root.FirstChild; c != nil; c = c.NextSibling {
		if n := findElement(c, pred); n != nil {
			return n
		}
	}
	return nil
}

type sectionExtractor struct {
	pageURL        string
	title          string
	currentHeading string
	currentURL     string
	body           strings.Builder
	chunks         []Chunk
}

func (e *sectionExtractor) walk(n *html.Node) {
	if n.Type == html.TextNode {
		e.writeText(n.Data)
		return
	}
	if n.Type != html.ElementNode {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			e.walk(c)
		}
		return
	}
	if shouldSkipElement(n) {
		return
	}
	if isHeading(n) {
		heading := cleanText(textContent(n))
		if heading != "" {
			e.flush()
			e.currentHeading = heading
			e.currentURL = withFragment(e.pageURL, headingFragment(n))
		}
		return
	}
	if strings.EqualFold(n.Data, "br") {
		e.writeNewline()
		return
	}
	if isBlockElement(n) {
		e.writeNewline()
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		e.walk(c)
	}
	if isBlockElement(n) {
		e.writeNewline()
	}
}

func (e *sectionExtractor) flush() {
	body := strings.TrimSpace(e.body.String())
	if e.currentHeading != "" && body != "" {
		e.chunks = append(e.chunks, Chunk{
			URL:     e.currentURL,
			Title:   e.title,
			Heading: e.currentHeading,
			Body:    body,
		})
	}
	e.body.Reset()
}

func (e *sectionExtractor) writeText(s string) {
	s = cleanText(s)
	if s == "" {
		return
	}
	current := e.body.String()
	if current != "" && !strings.HasSuffix(current, "\n") && !strings.HasSuffix(current, " ") {
		e.body.WriteByte(' ')
	}
	e.body.WriteString(s)
}

func (e *sectionExtractor) writeNewline() {
	current := e.body.String()
	if current != "" && !strings.HasSuffix(current, "\n") {
		e.body.WriteByte('\n')
	}
}

func isHeading(n *html.Node) bool {
	if len(n.Data) != 2 || n.Data[0] != 'h' {
		return false
	}
	return n.Data[1] >= '1' && n.Data[1] <= '6'
}

func shouldSkipElement(n *html.Node) bool {
	switch strings.ToLower(n.Data) {
	case "head", "script", "style", "noscript", "nav", "header", "footer", "aside", "svg", "form", "button", "input", "select", "textarea", "iframe", "canvas":
		return true
	}
	if _, ok := nodeAttr(n, "hidden"); ok {
		return true
	}
	if strings.EqualFold(attrOrEmpty(n, "aria-hidden"), "true") {
		return true
	}
	class := attrOrEmpty(n, "class")
	for _, skip := range []string{"copy-popup", "site-footer", "site-header", "search-container", "toc"} {
		if strings.Contains(class, skip) {
			return true
		}
	}
	return false
}

func isBlockElement(n *html.Node) bool {
	switch strings.ToLower(n.Data) {
	case "address", "article", "blockquote", "dd", "div", "dl", "dt", "figcaption", "figure", "hr", "li", "main", "ol", "p", "pre", "section", "table", "tbody", "td", "tfoot", "th", "thead", "tr", "ul":
		return true
	default:
		return false
	}
}

func textContent(n *html.Node) string {
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			if b.Len() > 0 {
				b.WriteByte(' ')
			}
			b.WriteString(n.Data)
			return
		}
		if n.Type == html.ElementNode && shouldSkipElement(n) {
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return cleanText(b.String())
}

func cleanText(s string) string {
	s = strings.Map(func(r rune) rune {
		switch r {
		case '\u00a0', '\u200b':
			return ' '
		default:
			return r
		}
	}, s)
	return strings.Join(strings.FieldsFunc(s, func(r rune) bool {
		return unicode.IsSpace(r)
	}), " ")
}

func headingFragment(n *html.Node) string {
	if id := attrOrEmpty(n, "id"); id != "" {
		return id
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if id := attrOrEmpty(c, "id"); id != "" {
			return id
		}
		if name := attrOrEmpty(c, "name"); name != "" {
			return name
		}
	}
	for p := n.Parent; p != nil; p = p.Parent {
		if id := attrOrEmpty(p, "id"); id != "" {
			return id
		}
		if name := attrOrEmpty(p, "name"); name != "" {
			return name
		}
	}
	return ""
}

func withFragment(rawURL, fragment string) string {
	if fragment == "" {
		return rawURL
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	u.Fragment = fragment
	return u.String()
}

func nodeAttr(n *html.Node, key string) (string, bool) {
	for _, attr := range n.Attr {
		if strings.EqualFold(attr.Key, key) {
			return attr.Val, true
		}
	}
	return "", false
}

func attrOrEmpty(n *html.Node, key string) string {
	if n == nil {
		return ""
	}
	val, _ := nodeAttr(n, key)
	return val
}

func refreshURL(content string) string {
	for _, part := range strings.Split(content, ";") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(strings.ToLower(part), "url=") {
			return strings.Trim(strings.TrimSpace(part[4:]), `"'`)
		}
	}
	return ""
}

func findGradleVersion(root *html.Node, title string) string {
	if n := findElement(root, func(n *html.Node) bool { return attrOrEmpty(n, "id") == "revnumber" }); n != nil {
		if v := versionFromText(textContent(n)); v != "" {
			return v
		}
	}
	if n := findElement(root, func(n *html.Node) bool { return strings.Contains(attrOrEmpty(n, "class"), "site-header-version") }); n != nil {
		if v := versionFromText(textContent(n)); v != "" {
			return v
		}
	}
	if v := versionFromText(title); v != "" {
		return v
	}
	return ""
}

var versionPattern = regexp.MustCompile(`(?i)(?:version|api)?\s*([0-9]+\.[0-9]+(?:\.[0-9]+)?)`)

func versionFromText(s string) string {
	match := versionPattern.FindStringSubmatch(s)
	if len(match) == 2 {
		return match[1]
	}
	return ""
}
