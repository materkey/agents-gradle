#!/bin/bash
# Clone or update the jrodbx/agp-sources repository (per-release bundled AGP sources).
# Usage: ./clone_agp_sources.sh
#   Honors $AGP_SOURCES_DIR (default: $HOME/projects/agp-sources).

set -e

AGP_SOURCES_DIR="${AGP_SOURCES_DIR:-$HOME/projects/agp-sources}"
REPO_URL="https://github.com/jrodbx/agp-sources.git"

if [ -d "$AGP_SOURCES_DIR/.git" ]; then
    echo "AGP sources already present at: $AGP_SOURCES_DIR"
    echo "Skipping clone; updating instead..."
    git -C "$AGP_SOURCES_DIR" pull --ff-only || echo "Skipping update (not a fast-forward); using existing checkout."
else
    echo "Cloning AGP sources from $REPO_URL (this may take a while)..."
    git clone --depth 1 "$REPO_URL" "$AGP_SOURCES_DIR"
fi

echo ""
echo "AGP sources ready at: $AGP_SOURCES_DIR"
echo "Available versions: $(ls -d "$AGP_SOURCES_DIR"/8.* "$AGP_SOURCES_DIR"/9.* 2>/dev/null | wc -l | tr -d ' ')"
