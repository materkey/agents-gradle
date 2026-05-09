#!/bin/bash
# Clone or update Gradle source code repository
# Usage: ./clone_gradle.sh [version]
#   version: optional git tag (e.g., v9.3.0) to checkout

set -e

GRADLE_SOURCES_DIR="${GRADLE_SOURCES_DIR:-$HOME/.gradle-sources}"
GRADLE_REPO_DIR="$GRADLE_SOURCES_DIR/gradle"
REPO_URL="https://github.com/gradle/gradle.git"
VERSION="${1:-}"

# Create base directory if needed
mkdir -p "$GRADLE_SOURCES_DIR"

if [ -d "$GRADLE_REPO_DIR/.git" ]; then
    echo "Updating existing Gradle repository..."
    cd "$GRADLE_REPO_DIR"
    git fetch --tags --prune

    if [ -n "$VERSION" ]; then
        echo "Checking out version: $VERSION"
        git checkout "$VERSION"
    else
        echo "Pulling latest from master..."
        git checkout master
        git pull origin master
    fi
else
    echo "Cloning Gradle repository (this may take a while)..."
    # Use shallow clone with history for faster initial clone
    git clone --depth 100 "$REPO_URL" "$GRADLE_REPO_DIR"
    cd "$GRADLE_REPO_DIR"

    if [ -n "$VERSION" ]; then
        echo "Fetching and checking out version: $VERSION"
        git fetch --depth 1 origin tag "$VERSION"
        git checkout "$VERSION"
    fi
fi

echo ""
echo "Gradle sources ready at: $GRADLE_REPO_DIR"
echo "Current commit: $(git rev-parse --short HEAD)"
echo "Branch/Tag: $(git describe --tags --always 2>/dev/null || git branch --show-current)"
