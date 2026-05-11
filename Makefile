GRADLE_DOCS_URL ?= https://docs.gradle.org/current/userguide/userguide.html
DB_PATH ?= cmd/gradle-rag/db/gradle.db
BINARY_PATH ?= gradle-rag/skills/gradle-rag/references/gradle-rag
DIST_DIR ?= dist
PLUGINS := gradle-rag gradle-grill agp-sources gradle-sources kotlin-sources ksp-sources
PYTHON := $(shell command -v python3.13 || command -v python3.12 || command -v python3.11 || command -v python3 || true)
ROOT := $(CURDIR)

.PHONY: build-db smoke-db build build-fast test install install-plugins install-local clean

build-db:
	go run ./cmd/crawler --start "$(GRADLE_DOCS_URL)" --db "$(DB_PATH)"

smoke-db:
	go run ./cmd/crawler --start "$(GRADLE_DOCS_URL)" --db "$(DB_PATH)" --max-pages 80 --workers 4

build: build-db
	mkdir -p "$(dir $(BINARY_PATH))"
	VERSION="v0.$$(date +%Y.%m%d)"; \
		go build -ldflags "-s -w -X main.version=$$VERSION" -o "$(BINARY_PATH)" ./cmd/gradle-rag
	chmod +x gradle-rag/skills/gradle-rag/bin/gradle-rag

build-fast:
	mkdir -p "$(dir $(BINARY_PATH))"
	go build -o "$(BINARY_PATH)" ./cmd/gradle-rag
	chmod +x gradle-rag/skills/gradle-rag/bin/gradle-rag

test:
	go test ./...

install-local: install-plugins

install: install-plugins

install-plugins: build-fast
	@for path in "$$HOME/.claude/skills/gradle" "$$HOME/.codex/skills/gradle" "$$HOME/.claude/skills/gradle-rag" "$$HOME/.codex/skills/gradle-rag"; do \
		if [ -L "$$path" ]; then \
			rm "$$path"; \
		elif [ -e "$$path" ]; then \
			echo "Refusing to remove non-symlink legacy skill at $$path" >&2; \
			exit 1; \
		fi; \
	done
	@config="$$HOME/.codex/config.toml"; \
	if [ -f "$$config" ]; then \
		tmp="$${config}.tmp"; \
		awk '\
			/^\[plugins\."(gradle-docs|gradle)@agents-gradle"\]/ { skip = 1; next } \
			skip && /^\[/ { skip = 0 } \
			!skip { print } \
		' "$$config" > "$$tmp"; \
		mv "$$tmp" "$$config"; \
	fi
	@if [ -z "$(PYTHON)" ] || ! "$(PYTHON)" -c 'import sys; sys.exit(0 if sys.version_info >= (3, 9) else 1)'; then \
		echo "Need Python 3.9+ (found: $(PYTHON))" >&2; \
		exit 1; \
	fi; \
	"$(PYTHON)" "$(ROOT)/scripts/install-codex-plugins.py" --root "$(ROOT)" $(PLUGINS)
	claude plugin marketplace add "$(ROOT)"
	claude plugin uninstall gradle-docs@agents-gradle --scope user --keep-data || true
	claude plugin uninstall gradle@agents-gradle --scope user --keep-data || true
	@for plugin in $(PLUGINS); do \
		claude plugin uninstall "$${plugin}@agents-gradle" --scope user --keep-data || true; \
		claude plugin install "$${plugin}@agents-gradle" --scope user; \
	done
	scripts/install-gradle-rag-bin.sh

clean:
	rm -rf "$(DIST_DIR)"
	rm -f "$(DB_PATH)"
	rm -f "$(BINARY_PATH)"
