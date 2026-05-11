#!/bin/bash
# KSP Sources Explorer
# Clones/updates google/ksp and provides search utilities

set -e

KSP_DIR="${KSP_SOURCES_DIR:-$HOME/.ksp-sources/ksp}"
REPO_URL="https://github.com/google/ksp.git"
DEFAULT_REF="${KSP_REF:-main}"

# Ensure repo is cloned
ensure_repo() {
    if [ ! -d "$KSP_DIR" ]; then
        echo "Cloning KSP repository..."
        mkdir -p "$(dirname "$KSP_DIR")"
        git clone --depth 100 "$REPO_URL" "$KSP_DIR"
    else
        echo "KSP repository found at $KSP_DIR"
    fi
}

checkout_ref() {
    local ref="${1:-$DEFAULT_REF}"
    cd "$KSP_DIR"

    if [ "$ref" = "main" ]; then
        git fetch --depth 1 origin main
        git checkout main
        git pull --ff-only origin main
    else
        if git fetch --depth 1 origin "$ref" 2>/dev/null; then
            git checkout --detach FETCH_HEAD
        else
            git fetch --depth 1 origin tag "$ref"
            git checkout "$ref"
        fi
    fi

    echo "KSP sources ready at: $KSP_DIR"
    echo "Current ref: $(git describe --tags --always 2>/dev/null || git branch --show-current)"
    echo "Current commit: $(git rev-parse --short HEAD)"
}

# Search for a class/interface
search_class() {
    local name="$1"
    echo "Searching for class/interface: $name"
    grep -rn "class $name\|interface $name\|object $name" "$KSP_DIR" --include="*.kt" | head -20
}

# Search for any pattern
search_pattern() {
    local pattern="$1"
    echo "Searching for pattern: $pattern"
    grep -rn "$pattern" "$KSP_DIR" --include="*.kt" | head -30
}

# List key API files
list_api() {
    echo "Key KSP API files:"
    find "$KSP_DIR/api" -name "*.kt" -type f 2>/dev/null | head -30
}

# Show help
show_help() {
    cat << EOF
KSP Sources Explorer

Usage: $0 <command> [args]

Commands:
    init [ref]        Clone KSP repo if not present and checkout ref (default: ${DEFAULT_REF})
    update [ref]      Fetch and checkout ref (default: ${DEFAULT_REF})
    search <pattern>  Search for pattern in Kotlin files
    class <name>      Search for class/interface/object definition
    api               List key API files
    help              Show this help

Examples:
    $0 init
    $0 init <tag-or-commit>
    $0 search "getSymbolsWithAnnotation"
    $0 class KSClassDeclaration
    $0 api
EOF
}

# Main
case "${1:-help}" in
    init)
        ensure_repo
        checkout_ref "${2:-$DEFAULT_REF}"
        ;;
    update)
        ensure_repo
        checkout_ref "${2:-$DEFAULT_REF}"
        ;;
    search)
        ensure_repo
        search_pattern "$2"
        ;;
    class)
        ensure_repo
        search_class "$2"
        ;;
    api)
        ensure_repo
        list_api
        ;;
    help|*)
        show_help
        ;;
esac
