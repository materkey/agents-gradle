#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SOURCE="${ROOT}/gradle-rag/skills/gradle-rag/references/gradle-rag"
INSTALL_DIR="${GRADLE_RAG_INSTALL_DIR:-${HOME}/.local/bin}"
TARGET="${INSTALL_DIR}/gradle-rag"

case "$(uname -s)" in
  Darwin|Linux) ;;
  *)
    echo "install-gradle-rag-bin.sh: unsupported OS: $(uname -s); only Darwin and Linux are supported" >&2
    exit 1
    ;;
esac

if [ ! -f "$SOURCE" ] || [ ! -x "$SOURCE" ]; then
  echo "install-gradle-rag-bin.sh: built gradle-rag binary is missing or not executable: $SOURCE" >&2
  echo "Run task build-fast or task build before installing." >&2
  exit 1
fi

mkdir -p "$INSTALL_DIR"

if [ ! -d "$INSTALL_DIR" ] || [ ! -w "$INSTALL_DIR" ]; then
  echo "install-gradle-rag-bin.sh: install directory is not writable: $INSTALL_DIR" >&2
  exit 1
fi

if [ -d "$TARGET" ]; then
  echo "install-gradle-rag-bin.sh: refusing to replace directory: $TARGET" >&2
  exit 1
fi

if [ -e "$TARGET" ] && [ ! -f "$TARGET" ] && [ ! -L "$TARGET" ]; then
  echo "install-gradle-rag-bin.sh: refusing to replace non-file target: $TARGET" >&2
  exit 1
fi

if [ -e "$TARGET" ] && [ ! -L "$TARGET" ] && [ ! -w "$TARGET" ]; then
  echo "install-gradle-rag-bin.sh: refusing to replace non-writable file: $TARGET" >&2
  exit 1
fi

tmp="$(mktemp "${INSTALL_DIR}/.gradle-rag.XXXXXX")"
trap 'rm -f "$tmp"' EXIT
cp "$SOURCE" "$tmp"
chmod 0755 "$tmp"
mv -f "$tmp" "$TARGET"
trap - EXIT

echo "Installed gradle-rag -> $TARGET"

case ":${PATH:-}:" in
  *":${INSTALL_DIR}:"*) ;;
  *)
    shell_name="$(basename "${SHELL:-}")"
    echo "Warning: $INSTALL_DIR is not in PATH." >&2

    case "$shell_name" in
      fish)
        echo "Add this in fish:" >&2
        if [ "$INSTALL_DIR" = "${HOME}/.local/bin" ]; then
          echo '  fish_add_path $HOME/.local/bin' >&2
        else
          echo "  fish_add_path \"${INSTALL_DIR}\"" >&2
        fi
        ;;
      bash|zsh)
        echo "Add this to ~/.${shell_name}rc:" >&2
        if [ "$INSTALL_DIR" = "${HOME}/.local/bin" ]; then
          echo '  export PATH="$HOME/.local/bin:$PATH"' >&2
        else
          echo "  export PATH=\"${INSTALL_DIR}:\$PATH\"" >&2
        fi
        ;;
      *)
        echo "Add this to your shell profile." >&2
        echo "For zsh/bash:" >&2
        if [ "$INSTALL_DIR" = "${HOME}/.local/bin" ]; then
          echo '  export PATH="$HOME/.local/bin:$PATH"' >&2
        else
          echo "  export PATH=\"${INSTALL_DIR}:\$PATH\"" >&2
        fi
        echo "For fish:" >&2
        if [ "$INSTALL_DIR" = "${HOME}/.local/bin" ]; then
          echo '  fish_add_path $HOME/.local/bin' >&2
        else
          echo "  fish_add_path \"${INSTALL_DIR}\"" >&2
        fi
        ;;
    esac
    ;;
esac
