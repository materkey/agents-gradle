GRADLE_DOCS_URL ?= https://docs.gradle.org/current/userguide/userguide.html
DB_PATH ?= cmd/gradle-rag/db/gradle.db
BINARY_PATH ?= plugins/gradle-rag/skills/gradle-rag/references/gradle-rag
DIST_DIR ?= dist

.PHONY: crawl-docs crawl-docs-sample build build-cli install build-db smoke-db build-fast test clean

crawl-docs:
	go run ./cmd/crawler --start "$(GRADLE_DOCS_URL)" --db "$(DB_PATH)"

crawl-docs-sample:
	go run ./cmd/crawler --start "$(GRADLE_DOCS_URL)" --db "$(DB_PATH)" --max-pages 80 --workers 4

build: crawl-docs
	mkdir -p "$(dir $(BINARY_PATH))"
	VERSION="v0.$$(date +%Y.%m%d)"; \
		go build -ldflags "-s -w -X main.version=$$VERSION" -o "$(BINARY_PATH)" ./cmd/gradle-rag
	chmod +x plugins/gradle-rag/skills/gradle-rag/bin/gradle-rag

build-cli:
	mkdir -p "$(dir $(BINARY_PATH))"
	go build -o "$(BINARY_PATH)" ./cmd/gradle-rag
	chmod +x plugins/gradle-rag/skills/gradle-rag/bin/gradle-rag

install: build
	scripts/install-gradle-rag-bin.sh

build-db: crawl-docs

smoke-db: crawl-docs-sample

build-fast: build-cli

test:
	go test ./...

clean:
	rm -rf "$(DIST_DIR)"
	rm -f "$(DB_PATH)"
	rm -f "$(BINARY_PATH)"
